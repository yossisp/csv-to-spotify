package kafkahelper

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/yossisp/csv-to-spotify/pkg/utils"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/yossisp/csv-to-spotify/pkg/config"
)

var (
	conf               config.Config = config.NewConfig()
	trackProgressTopic string        = conf.KafkaTrackProgressTopic
	logger                           = utils.NewLogger("kafkahelper")
)

type messageType int

const (
	// TrackProgress - that the message carries progress information
	TrackProgress messageType = iota
	// JobFinished - that the job is finished
	JobFinished
	// CSVFileError - an error with CSV file occurred
	CSVFileError
)

// Producer holds kafka producer
type Producer struct {
	*kafka.Producer
}

// Consumer holds kafka consumer
type Consumer struct {
	*kafka.Consumer
}

// Message is used to communicate between producer and consumer
type Message struct {
	MsgType messageType `json:"msgType"`
	Msg     interface{} `json:"msg"`
	UserID  string
}

// trackProgress is used to communicate how many tracks were
// added/not added by the runner
type trackProgress struct {
	TracksAdded    int `json:"tracksAdded"`
	TracksNotAdded int `json:"tracksNotAdded"`
}

// getTrackProgressMsg creates track progress message
func getTrackProgressMsg(trackData []interface{}) ([]byte, error) {
	const funcName = "getTrackProgressMsg"
	tracksAdded, ok := trackData[0].(int)
	if !ok {
		return nil, fmt.Errorf("%s: couldn't get tracksAdded", funcName)
	}
	tracksNotAdded, ok := trackData[1].(int)
	if !ok {
		return nil, fmt.Errorf("%s: couldn't get tracksNotAdded", funcName)
	}
	progressMsg := trackProgress{
		TracksAdded:    tracksAdded,
		TracksNotAdded: tracksNotAdded,
	}
	return json.Marshal(Message{
		MsgType: TrackProgress,
		Msg:     progressMsg,
	})
}

// getJobFinishedMsg notifies that the job has finished
func getJobFinishedMsg() ([]byte, error) {
	return json.Marshal(Message{
		MsgType: JobFinished,
	})
}

// getCSVFileErrorMsg notifies that the job has finished
func getCSVFileErrorMsg() ([]byte, error) {
	return json.Marshal(Message{
		MsgType: CSVFileError,
		Msg: map[string]string{
			"error": "CSV file error",
		},
	})
}

// getConfig returns kafka config for producer/consumer
func getConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"metadata.broker.list": conf.KafkaBrokers,
		"security.protocol":    "SASL_SSL",
		"sasl.mechanisms":      "SCRAM-SHA-256",
		"sasl.username":        conf.KafkaUsername,
		"sasl.password":        conf.KafkaPassword,
		"group.id":             conf.KafkaGroupID,
		"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"},
		//"debug":                           "generic,broker,security",
	}
}

// LogDeliveredMessages logs delivered messages by producer
func (producer *Producer) LogDeliveredMessages() {
	const funcName = "LogDeliveredMessages"
	for event := range producer.Events() {
		switch eventType := event.(type) {
		case *kafka.Message:
			if eventType.TopicPartition.Error != nil {
				logger("%s: Delivery failed: %v\n", funcName, eventType.TopicPartition)
			} else {
				logger("%s: Delivered message to %v\n", funcName, eventType.TopicPartition)
			}
		}
	}
}

// NewProducer returns a producer
func NewProducer() *Producer {
	producer, err := kafka.NewProducer(getConfig())
	if err != nil {
		log.Fatalln("kafka.NewProducer: ", err)
	}
	return &Producer{producer}
}

// NewConsumer returns new consumer
func NewConsumer() *Consumer {
	consumer, err := kafka.NewConsumer(getConfig())
	if err != nil {
		log.Fatalln("kafka.NewConsumer: ", err)
	}
	return &Consumer{consumer}
}

// ProduceMessage produces kafka message
func (producer *Producer) ProduceMessage(userID string, msgType messageType, msgParams ...interface{}) {
	const funcName = "ProduceMessage"
	var (
		msg []byte
		err error
	)

	switch msgType {
	case TrackProgress:
		msg, err = getTrackProgressMsg(msgParams)
	case JobFinished:
		msg, err = getJobFinishedMsg()
	case CSVFileError:
		msg, err = getCSVFileErrorMsg()
	default:
		logger("%s: unknown message type: %v", funcName, msgType)
	}
	if err != nil {
		logger("%s: get message %v", funcName, err)
	} else {
		producer.Produce(&kafka.Message{TopicPartition: kafka.TopicPartition{
			Topic: &trackProgressTopic, Partition: kafka.PartitionAny},
			Key:       []byte(userID),
			Value:     msg,
			Timestamp: time.Now(),
		}, nil)
	}
}

// ConsumeMessages consumes kafka messages
func (consumer *Consumer) ConsumeMessages(msgChan chan Message) {
	const funcName = "ConsumeMessages"
	err := consumer.SubscribeTopics([]string{trackProgressTopic}, nil)
	if err != nil {
		log.Fatalln("consumer.SubscribeTopics: ", err)
	}
	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			logger("%s: Message on %s: %s ts: %v", funcName, msg.TopicPartition, string(msg.Value), msg.Timestamp)
			kafkaMsg := Message{}
			err = json.Unmarshal(msg.Value, &kafkaMsg)
			if err != nil {
				logger("%s: json.Unmarshal: %v", funcName, err)
			} else {
				kafkaMsg.UserID = string(msg.Key)
				msgChan <- kafkaMsg
			}
		} else {
			// The client will automatically try to recover from all errors.
			logger("%s: consumer.ReadMessage: %v", funcName, err)
		}
	}
}

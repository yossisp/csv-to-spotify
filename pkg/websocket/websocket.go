package websocket

/**
websocket connection quits when:
1. client sends close message (e.g. page refresh)
2. user id was not found
3. job has finished

cases checked/handled:
1. user sends playlist to server, refreshes the page, still gets the results
2. refresh without starting copy
3. user login -> ws opened
4. copy job started, user exits before it's finished
5. 2 clients starting copy job in parallel
*/

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/yossisp/csv-to-spotify/pkg/utils"

	"github.com/yossisp/csv-to-spotify/pkg/db"
	"github.com/yossisp/csv-to-spotify/pkg/kafkahelper"

	"github.com/gorilla/websocket"
)

type clientPayload struct {
	MessageType    string      `json:"type"`
	MessagePayload interface{} `json:"payload,omitempty"`
}

type writePayload struct {
	message       interface{}
	isJobFinished bool
}

// Websocket makes sure to that only 1 writer is active
// at any time because Gorilla
// connections support one concurrent reader and one concurrent writer.
// (https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency)
type Websocket struct {
	*websocket.Conn
	writeChan chan writePayload
	quitChan  chan bool
	now       time.Time // TODO: remove
	userID    string
}

// SafeMap is intended for thread-safe access to map
type safeMap struct {
	smap  map[string]*Websocket
	mutex sync.Mutex
}

const (
	user        = "USER"
	update      = "UPDATE"
	jobFinished = "JOB_FINISHED"
	pongWait    = 10 * time.Second
	logPrefix   = "websocket.go"
)

var (
	upgrader                                  = websocket.Upgrader{}
	consumer                                  = kafkahelper.NewConsumer()
	msgChan          chan kafkahelper.Message = make(chan kafkahelper.Message)
	wsConnectionsMap *safeMap                 = &safeMap{smap: make(map[string]*Websocket)}
	logger                                    = utils.NewLogger("websocket.go")
)

func init() {
	go consumer.ConsumeMessages(msgChan)
	go wsConnectionsMap.processKafkaMessage()
}

// listens on socket connection and quits if some read error occurred
// tells WSConnectionHandler to quit via quitChan
func (ws *Websocket) listen() {
	defer func() {
		log.Println("defer")
		close(ws.quitChan)
	}()

	for {
		clientMessage := clientPayload{}
		_, message, err := ws.ReadMessage()
		if err != nil {
			logger("user: %s listen -> ReadMessage: %v", ws.userID, err)
			break

		}
		err = json.Unmarshal(message, &clientMessage)
		if err != nil {
			logger("user: %s listen -> json.Unmarshal: %v", ws.userID, err)
		}

		switch clientMessage.MessageType {
		case user:
			userID, ok := clientMessage.MessagePayload.(string)
			if ok {
				log.Println("ws listen: userID: ", userID)
				dbUser := db.FindSpotifyUser(userID)
				if dbUser != nil {
					ws.userID = userID
					wsConnectionsMap.mutex.Lock()
					wsConnectionsMap.smap[userID] = ws
					wsConnectionsMap.mutex.Unlock()
					log.Println("ws listen: user found in db: ", dbUser.UserID)
					ws.writeChan <- writePayload{
						message: map[string]interface{}{
							"type":    user,
							"payload": true,
						},
					}
					wsConnectionsMap.mutex.Lock()
					log.Println("wsConnectionsMap", wsConnectionsMap.smap)
					wsConnectionsMap.mutex.Unlock()
				} else {
					ws.writeChan <- writePayload{
						message: map[string]interface{}{
							"type":    user,
							"payload": false,
						},
					}
					logger("user: %s not found", ws.userID)
					return
				}
			}
		default:
			logger("unknown message type received: %v", clientMessage)
		}
	}
	log.Println("end of listen")
}

func (connectionsMap *safeMap) get(userID string) (connection *Websocket, found bool) {
	connectionsMap.mutex.Lock()
	connection, found = connectionsMap.smap[userID]
	connectionsMap.mutex.Unlock()
	return
}

// processKafkaMessage is started at init()
// it forwards results to WSConnectionHandler which writes to client
func (connectionsMap *safeMap) processKafkaMessage() {
	var (
		msg kafkahelper.Message
	)
	for {
		select {
		case msg = <-msgChan:
			log.Println("processKafkaMessage msg: ", msg)
			switch msg.MsgType {
			case kafkahelper.TrackProgress:
				connection, found := connectionsMap.get(msg.UserID)
				if found {
					connection.writeChan <- writePayload{
						message: clientPayload{
							MessageType:    update,
							MessagePayload: msg.Msg,
						},
					}
				} else {
					logger("processKafkaMessage: user id: %s not found in wsConnectionsMap", msg.UserID)
				}

			case kafkahelper.JobFinished:
				connection, found := connectionsMap.get(msg.UserID)
				if found {
					connection.writeChan <- writePayload{
						message: clientPayload{
							MessageType: jobFinished,
						},
						isJobFinished: true,
					}
					log.Println("processKafkaMessage: JobFinished")
				} else {
					logger("processKafkaMessage: user id: %s not found in wsConnectionsMap", msg.UserID)
				}
			case kafkahelper.CSVFileError:
				connection, found := connectionsMap.get(msg.UserID)
				if found {
					connection.writeChan <- writePayload{
						message: clientPayload{
							MessageType:    jobFinished,
							MessagePayload: msg.Msg,
						},
						isJobFinished: true,
					}
				}
			default:
				logger("processKafkaMessage: unknown message: %v", msg)
			}
		}
	}
}

// WSConnectionHandler (/websocket route) handles
// communication with the client via websocket.
// It tells listen() to quit when the job finishes by closing socket connection
// also removes user id from connections map so that
// processKafkaMessage stops sending via writeChan
func WSConnectionHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: filter allowed origins
	upgrader.CheckOrigin = func(req *http.Request) bool {
		return true
	}
	connection, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		logger("upgrade error: %v", err)
		return
	}
	defer connection.Close()
	log.Println("new ws upgraded")
	websocket := &Websocket{
		connection,
		make(chan writePayload),
		make(chan bool),
		time.Now(),
		"",
	}
	defer func() {
		close(websocket.writeChan)
		wsConnectionsMap.mutex.Lock()
		if _, found := wsConnectionsMap.smap[websocket.userID]; found {
			delete(wsConnectionsMap.smap, websocket.userID)
			log.Println("found websocket.user in wsConnectionsMap")
		}
		wsConnectionsMap.mutex.Unlock()
		log.Println("WSConnectionHandler quitChan")
	}()
	go websocket.listen()
	for {
		select {
		case update := <-websocket.writeChan:
			err := websocket.WriteJSON(update.message)
			log.Println("WriteJSON: msg", update.message)
			if err != nil {
				logger("WriteJSON: %v", err)
			}
			if update.isJobFinished {
				log.Println("websocket.isJobFinished")
				return
			}
		case <-websocket.quitChan:
			log.Println("<-websocket.quitChan")
			return
		}
	}
}

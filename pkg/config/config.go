package config

import (
	"os"

	"github.com/yossisp/csv-to-spotify/pkg/utils"

	"github.com/joho/godotenv"
)

var (
	logger = utils.NewLogger("config")
)

// LoadEnv loads environment variables
func LoadEnv() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		logger("No .env file found")
	}
}

func init() {
	LoadEnv()
}

// Config object
type Config struct {
	// base64 encoded client_id:client_secret
	SpotifySecret           string
	Env                     string
	Market                  string
	MongoDBName             string
	MongoConnectionString   string
	KafkaBrokers            string
	KafkaUsername           string
	KafkaPassword           string
	KafkaGroupID            string
	KafkaTrackProgressTopic string
	InputFileExt            string
	ClientTimeout           int64
	Port                    string
	AllowedOrigins          string
	TrackLookupInterval     string
	TestRefreshToken        string
}

// NewConfig returns config
func NewConfig() Config {
	return Config{
		SpotifySecret:           getEnvVar("SPOTIFY_CLIENT_ID_SECRET_BASE64", ""),
		Env:                     getEnvVar("ENV", "development"),
		Market:                  getEnvVar("MARKET", "US"),
		MongoDBName:             getEnvVar("MONGO_DB_NAME", ""),
		MongoConnectionString:   getEnvVar("MONGO_ATLAS_CONNECTION", ""),
		KafkaBrokers:            getEnvVar("KAFKA_BROKERS", ""),
		KafkaUsername:           getEnvVar("KAFKA_USERNAME", ""),
		KafkaPassword:           getEnvVar("KAFKA_PASSWORD", ""),
		KafkaGroupID:            getEnvVar("KAFKA_GROUP_ID", ""),
		KafkaTrackProgressTopic: getEnvVar("KAFKA_TRACK_PROGRESS_TOPIC", ""),
		InputFileExt:            getEnvVar("INPUT_FILE_EXT", ".csv"),
		ClientTimeout:           30,
		Port:                    getEnvVar("PORT", "8000"),
		AllowedOrigins:          getEnvVar("ALLOWED_ORIGINS", "http://localhost:3000"),
		TrackLookupInterval:     getEnvVar("TRACK_LOOKUP_INTERVAL", "5"),
		TestRefreshToken:        getEnvVar("TEST_REFRESH_TOKEN", ""),
	}
}

func getEnvVar(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

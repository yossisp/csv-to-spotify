package db

import (
	"context"
	"log"
	"time"

	"github.com/yossisp/csv-to-spotify/pkg/utils"

	"github.com/yossisp/csv-to-spotify/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ctx    = context.TODO()
	client *mongo.Client
	conf   config.Config = config.NewConfig()
	logger               = utils.NewLogger("mongoClient")
)

func InitMongoConnection() {
	clientOptions := options.Client().ApplyURI(conf.MongoConnectionString)
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalln("mongo.Connect", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalln("mongo ping", err)
	}
}

func init() {
	InitMongoConnection()
}

// InsertSpotifyUser adds/updates spotify user data
func InsertSpotifyUser(user SpotifyUser) {
	const funcName = "InsertSpotifyUser"
	opts := options.Update().SetUpsert(true)
	collection := client.Database(conf.MongoDBName).Collection("test")

	user.UpdatedAt = time.Now()
	_, err := collection.UpdateOne(ctx, bson.D{{"userId", user.UserID}},
		bson.D{{"$set", user}}, opts)

	if err != nil {
		logger("%s UpdateOne: %v", funcName, err)
	}
}

// FindSpotifyUser finds spotify user
func FindSpotifyUser(userID string) *SpotifyUser {
	const funcName = "FindSpotifyUser"
	collection := client.Database(conf.MongoDBName).Collection("test")
	var user SpotifyUser
	err := collection.FindOne(ctx, bson.D{{"userId", userID}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		logger("%s UpdateOne: %v", funcName, err)
	}
	return &user
}

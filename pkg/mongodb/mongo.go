package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"lkrouter/config"
	"sync"
)

/*
	Used to create a singleton object of MongoDB client.

Initialized and exposed through  GetMongoClient().
*/
var clientInstance *mongo.Client

// Used during creation of singleton client object in GetMongoClient().
var clientInstanceError error

// Used to execute client creation procedure only once.
var mongoOnce sync.Once

// GetMongoClient - Return mongodb connection to work with
func GetMongoClient() (*mongo.Client, error) {
	//Perform connection creation operation only once.
	mongoOnce.Do(func() {
		cfg := config.GetConfig()
		// Set client options
		clientOptions := options.Client().ApplyURI(cfg.MongoDbConnection)
		// Connect to MongoDB
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			clientInstanceError = err
		}
		// Check the connection
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			clientInstanceError = err
		}
		clientInstance = client
	})
	return clientInstance, clientInstanceError
}

package mrequests

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"lkrouter/pkg/mongodb"
)

func GetCallByRoom(room string) (bson.M, error) {
	//update mongo calls item
	var call bson.M
	logger := logrus.New()
	mongoClient, err := mongodb.GetMongoClient()
	if err == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("calls")

		err = callsCollection.FindOne(ctx, bson.D{{"url", room}}).Decode(&call)
		if err != nil {
			call = nil
			if err == mongo.ErrNoDocuments {
				// This error means your query did not match any documents.
				logger.Infof("Can't find room %v Documents! \n", room)
			}
		}
		logger.Infof("Find call document %v by room %v! \n", call["_id"], room)
	}
	return call, err
}

func UpdateCallByBsonFilter(filter bson.M, actions bson.M) error {
	mongoClient, errClient := mongodb.GetMongoClient()
	if errClient == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("calls")
		_, err := callsCollection.UpdateOne(ctx, filter, actions)

		return err
	}
	return errClient
}

func SetRecordStatus(room string, status bool) error {
	mongoClient, errClient := mongodb.GetMongoClient()
	if errClient == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("calls")
		_, err := callsCollection.UpdateOne(ctx, bson.M{"url": room}, bson.M{"$set": bson.M{"rec": status}})
		return err
	}
	return errClient
}

package mrequests

import (
	"context"
	"fmt"
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

func UpdateTranscribeTextStatus(room string, status string) error {
	//update mongo calls item
	mongoClient, err := mongodb.GetMongoClient()
	if err == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("calls")

		_, err = callsCollection.UpdateOne(ctx, bson.D{{"url", room}}, bson.D{
			{"$set", bson.D{{
				"file_transcribe_status",
				status}}}})

		if err != nil {
			fmt.Printf("Can't update transcribe text status %v for room %v! \n", status, room)
		}
		fmt.Printf("Update transcribe text status for room %v! \n", room)
	}
	return err
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

func UpdateTranscribeText(room string, textData []map[string]interface{}) error {
	//update mongo calls item
	mongoClient, err := mongodb.GetMongoClient()
	if err == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("calls")

		_, err = callsCollection.UpdateOne(ctx, bson.D{{"url", room}}, bson.D{
			{"$set", bson.D{{
				"transcrib_rec",
				textData}}},
			{"$set", bson.D{{
				"file_transcribe_status",
				"success"}}}})

		if err != nil {
			fmt.Printf("Can't update transcribe text %v for room %v! \n", textData, room)
		}
		fmt.Printf("Update transcribe text for room %v! \n", room)
	}
	return err
}

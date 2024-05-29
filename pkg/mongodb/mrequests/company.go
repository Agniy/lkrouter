package mrequests

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"lkrouter/pkg/mongodb"
	"math"
)

func GetCompany(companyId string) (bson.M, error) {
	//update mongo calls item
	var company bson.M
	logger := logrus.New()
	mongoClient, err := mongodb.GetMongoClient()
	if err == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("company")

		err = callsCollection.FindOne(ctx, bson.D{{"companyId", companyId}}).Decode(&company)
		if err != nil {
			company = nil
			if err == mongo.ErrNoDocuments {
				// This error means your query did not match any documents.
				logger.Infof("Can't find company %v Documents! \n", companyId)
			}
		}
		logger.Infof("Find company document %v by companyId %v! \n", company["_id"], companyId)
	}
	return company, err
}

func IsRoomActive(roomUrl string) (bool, error) {
	logger := logrus.New()
	mongoClient, err := mongodb.GetMongoClient()
	if err == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("calls")

		var call bson.M

		// find call by roomUrl and endDate is not null
		err = callsCollection.FindOne(ctx, bson.D{{"$and", []interface{}{
			bson.D{{"url", roomUrl}},
			bson.D{{"endDate", nil}},
		},
		}}).Decode(&call)

		if err != nil {
			call = nil
			if err == mongo.ErrNoDocuments {
				// This error means your query did not match any documents.
				logger.Infof("Can't find active room %v Documents! \n", roomUrl)
			}
		}
		logger.Infof("Find active room document %v by roomUrl %v! \n", call["_id"], roomUrl)
		return call != nil, err
	}
	return false, err
}

// UpdateCallStt
// sttAddMilliseconds - in miliseeconds
func UpdateCallStt(roomUrl string, sttAddMilliseconds int32) error {
	logger := logrus.New()
	//update mongo call
	mongoClient, errClient := mongodb.GetMongoClient()
	if errClient == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("calls")
		result, err := callsCollection.UpdateOne(
			ctx,
			bson.M{"url": roomUrl},
			bson.D{
				{"$inc", bson.D{
					{"stt_total", sttAddMilliseconds},
				},
				},
			},
		)
		if err != nil {
			logger.Infof("error when try to UpdateCompanyFilesSum: %v", err)
			return err
		}
		logger.Infof("Updated %v Documents! \n", result.ModifiedCount)
	}

	return errClient
}

// UpdateCompanyStt
// sttAddMilliseconds - in seconds
func UpdateCompanyStt(companyId string, sttAddSeconds int32) error {
	logger := logrus.New()
	//update mongo call
	mongoClient, errClient := mongodb.GetMongoClient()
	if errClient == nil {
		ctx := context.Background()
		callsCollection := mongoClient.Database("teleporta").Collection("company")
		result, err := callsCollection.UpdateOne(
			ctx,
			bson.M{"companyId": companyId},
			bson.D{
				{"$inc", bson.D{
					{"sttCurrent", sttAddSeconds},
				},
				},
			},
		)
		if err != nil {
			logger.Infof("error when try to UpdateCompanyStt: %v", err)
			return err
		}
		logger.Infof("Updated %v Documents! \n", result.ModifiedCount)
	}
	return errClient
}

// UpdateCompanySttStatsByRoom
// update company and call stt
// sttAddMilliseconds - time in milliseconds
func UpdateCompanySttStatsByRoom(roomUrl string, sttAddMilliseconds int32) error {
	logger := logrus.New()
	call, err := GetCallByRoom(roomUrl)
	if err != nil {
		logger.Infof("Error when try to get room %v \n", roomUrl)
		return err
	}

	if call["companyId"] == nil {
		return fmt.Errorf("call for %v has not companyId", call["_id"])
	}

	//update call stt
	err = UpdateCallStt(roomUrl, sttAddMilliseconds)
	if err != nil {
		logger.Infof("Error when try to update call stt_total")
	}

	companyId := call["companyId"].(string)
	addSttSeconds := int32(math.Round(float64(sttAddMilliseconds) / 1000))
	err = UpdateCompanyStt(companyId, addSttSeconds)
	if err != nil {
		logger.Infof("UpdateCompanyStt error for company: %v with error: %v", companyId, err)
		return err
	} else {
		logger.Infof("UpdateCompanyStt success for company: %v with sumSttCurrent: %v", companyId, addSttSeconds)
	}

	return nil
}

// CheckCompanySttLimit	- check if company stt limit is reached
// sttAdd - in seconds
// companyId - company id
func CheckCompanySttLimit(companyId string, sttAdd int32) error {
	logger := logrus.New()
	company, err := GetCompany(companyId)
	if err != nil {
		logger.Infof("Error when try to get company by call %v \n", companyId)
		return err
	}
	sttLimit := company["sttLimit"].(int32) * 60
	sttCurrent := company["sttCurrent"].(int32)
	if sttCurrent+sttAdd > sttLimit {
		return fmt.Errorf("sttCurrent %v + sttAdd %v > sttLimit %v", sttCurrent, sttAdd, sttLimit)
	}
	return nil
}

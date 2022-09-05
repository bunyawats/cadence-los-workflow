package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

const (
	mongoUri                      = "MONGO_URI"
	mongoDatabase                 = "MONGO_DATABASE"
	loanApplicationCollectionName = "loan_application"
)

var (
	mongoDBClient *mongo.Client
	ctx           context.Context
)

func init() {

	ctx = context.Background()
	mongoDBClient, _ = mongo.Connect(
		ctx,
		options.Client().ApplyURI(os.Getenv(mongoUri)),
	)
	mongoDBClient.Ping(ctx, nil)
	log.Println("mongo connected")
}

func UpdateLoanApplicationTaskToken(appID string, lastState string, taskToken string) error {

	collection := mongoDBClient.Database(
		os.Getenv(mongoDatabase)).Collection(loanApplicationCollectionName)

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"taskToken": taskToken,
			"lastState": lastState,
		},
	}

	result, err := collection.UpdateOne(
		context.Background(),
		filter,
		update,
	)

	if err != nil {
		fmt.Println("UpdateOne() result ERROR:", err)
	} else if result.MatchedCount == 0 {
		fmt.Println("UpdateOne() result:", result)
		return fmt.Errorf("Application ID: `%v` not found", appID)
	}

	return err
}

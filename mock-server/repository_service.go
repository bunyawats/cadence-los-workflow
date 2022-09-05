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
	collectionNameLoanApplication = "loan_application"
)

var (
	client     *mongo.Client
	ctx        context.Context
	collection *mongo.Collection
)

func init() {
	ctx = context.Background()
	client, _ = mongo.Connect(
		ctx,
		options.Client().ApplyURI(os.Getenv(mongoUri)),
	)
	client.Ping(ctx, nil)
	log.Println("mongo connected")

	collection = client.Database(
		os.Getenv(mongoDatabase)).Collection(collectionNameLoanApplication)
}

func CreateNewLoanApplication(appID string) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	if err := collection.FindOne(ctx, filter).Decode(&LoanApplication{}); err == nil {
		return fmt.Errorf("Application ID: `%v` duplicated", appID)
	}

	insert := bson.M{
		"appID":     appID,
		"lastState": "NEW_APPLICATION",
	}

	collection.InsertOne(
		ctx,
		insert,
	)

	return nil
}

func CreateLoanApplication(loanApp *LoanApplication) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": loanApp.AppID,
		},
	}

	if err := collection.FindOne(ctx, filter).Decode(&LoanApplication{}); err == nil {
		return fmt.Errorf("Application ID: `%v` duplicated", loanApp.AppID)
	}

	insert := bson.M{
		"fname":     loanApp.Fname,
		"lname":     loanApp.Lname,
		"appID":     loanApp.AppID,
		"lastState": loanApp.LastState,
	}

	collection.InsertOne(
		ctx,
		insert,
	)

	return nil
}

func SaveFormOne(loanApp *LoanApplication) (*LoanApplication, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": loanApp.AppID,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"fname": loanApp.Fname,
			"lname": loanApp.Lname,
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
		return nil, fmt.Errorf("Application ID: `%v` not found", loanApp.AppID)
	}

	updatedLoanApp := LoanApplication{}
	collection.FindOne(ctx, filter).Decode(&updatedLoanApp)

	return &updatedLoanApp, err
}

func SaveFormTwo(loanApp *LoanApplication) (*LoanApplication, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": loanApp.AppID,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"email":   loanApp.Email,
			"phoneNo": loanApp.PhoneNo,
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
		return nil, fmt.Errorf("Application ID: `%v` not found", loanApp.AppID)
	}

	updatedLoanApp := LoanApplication{}
	collection.FindOne(ctx, filter).Decode(&updatedLoanApp)

	return &updatedLoanApp, err
}

func GetTokenByAppID(appID string) (string, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	loanApplication := &LoanApplication{}
	if err := collection.FindOne(ctx, filter).Decode(loanApplication); err != nil {
		return "NA", err
	}
	log.Printf("loan application \n %v\n", loanApplication)

	return loanApplication.TaskToken, nil
}

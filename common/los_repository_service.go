package common

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
	mongoDBClient   *mongo.Client
	mongoCollection *mongo.Collection
	ctx             context.Context
)

func init() {

	ctx = context.Background()
	mongoDBClient, _ = mongo.Connect(
		ctx,
		options.Client().ApplyURI(os.Getenv(mongoUri)),
	)
	mongoDBClient.Ping(ctx, nil)
	log.Println("mongo connected")

	mongoCollection = mongoDBClient.Database(
		os.Getenv(mongoDatabase)).Collection(loanApplicationCollectionName)
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

func CreateNewLoanApplication(appID string) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	if err := mongoCollection.FindOne(ctx, filter).Decode(&LoanApplication{}); err == nil {
		return fmt.Errorf("Application ID: `%v` duplicated", appID)
	}

	insert := bson.M{
		"appID":     appID,
		"lastState": "NEW_APPLICATION",
	}

	mongoCollection.InsertOne(
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

	if err := mongoCollection.FindOne(ctx, filter).Decode(&LoanApplication{}); err == nil {
		return fmt.Errorf("Application ID: `%v` duplicated", loanApp.AppID)
	}

	insert := bson.M{
		"fname":     loanApp.Fname,
		"lname":     loanApp.Lname,
		"appID":     loanApp.AppID,
		"lastState": loanApp.LastState,
	}

	mongoCollection.InsertOne(
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

	result, err := mongoCollection.UpdateOne(
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
	mongoCollection.FindOne(ctx, filter).Decode(&updatedLoanApp)

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

	result, err := mongoCollection.UpdateOne(
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
	mongoCollection.FindOne(ctx, filter).Decode(&updatedLoanApp)

	return &updatedLoanApp, err
}

func GetTokenByAppID(appID string) (string, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	loanApplication := &LoanApplication{}
	if err := mongoCollection.FindOne(ctx, filter).Decode(loanApplication); err != nil {
		return "NA", err
	}

	return loanApplication.TaskToken, nil
}

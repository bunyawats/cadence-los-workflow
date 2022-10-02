package common

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const (
	loanApplicationCollectionName = "loan_application"
)

type (
	MongodbHelper struct {
		getCon         func() *mongo.Database
		collectionName string
		//mongoCollection *mongo.Collection
		ctx context.Context
	}

	MongodbConfig struct {
		MongoUri      string
		MongoDatabase string
	}
)

func NewMongodbHelper(config MongodbConfig) *MongodbHelper {

	ctx := context.Background()

	getConn := func() *mongo.Database {

		mongoDBClient, _ := mongo.Connect(
			ctx,
			options.Client().ApplyURI(config.MongoUri),
		)
		mongoDBClient.Ping(ctx, nil)
		log.Println("mongo connected")

		return mongoDBClient.Database(config.MongoDatabase)
	}

	return &MongodbHelper{
		ctx:            ctx,
		getCon:         getConn,
		collectionName: loanApplicationCollectionName,
	}
}

func NewMongodbHelperWithCallBack(getDB func() *mongo.Database) *MongodbHelper {
	return &MongodbHelper{
		ctx:            context.Background(),
		getCon:         getDB,
		collectionName: loanApplicationCollectionName,
	}
}

func (m *MongodbHelper) getCollection() *mongo.Collection {
	return m.getCon().Collection(m.collectionName)
}

func (m *MongodbHelper) UpdateLoanApplicationTaskToken(appID string, lastState string, workflowID string, runID string) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"workflowID": workflowID,
			"runID":      runID,
			"lastState":  lastState,
		},
	}

	result, err := m.getCollection().UpdateOne(
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

func (m *MongodbHelper) CreateNewLoanApplication(appID string) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	if err := m.getCollection().FindOne(m.ctx, filter).Decode(&LoanApplication{}); err == nil {
		return fmt.Errorf("Application ID: `%v` duplicated", appID)
	}

	insert := bson.M{
		"appID":     appID,
		"lastState": "NEW_APPLICATION",
	}

	m.getCollection().InsertOne(
		m.ctx,
		insert,
	)

	return nil
}

func (m *MongodbHelper) CreateLoanApplication(loanApp *LoanApplication) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": loanApp.AppID,
		},
	}

	if err := m.getCollection().FindOne(m.ctx, filter).Decode(&LoanApplication{}); err == nil {
		return fmt.Errorf("Application ID: `%v` duplicated", loanApp.AppID)
	}

	insert := bson.M{
		"fname":     loanApp.Fname,
		"lname":     loanApp.Lname,
		"appID":     loanApp.AppID,
		"lastState": loanApp.LastState,
	}

	m.getCollection().InsertOne(
		m.ctx,
		insert,
	)

	return nil
}

func (m *MongodbHelper) SaveFormOne(loanApp *LoanApplication) (*LoanApplication, error) {

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

	result, err := m.getCollection().UpdateOne(
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
	m.getCollection().FindOne(m.ctx, filter).Decode(&updatedLoanApp)

	return &updatedLoanApp, err
}

func (m *MongodbHelper) SaveFormTwo(loanApp *LoanApplication) (*LoanApplication, error) {

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

	result, err := m.getCollection().UpdateOne(
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
	m.getCollection().FindOne(m.ctx, filter).Decode(&updatedLoanApp)

	return &updatedLoanApp, err
}

func (m *MongodbHelper) GetWorkflowIdByAppID(appID string) (string, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	loanApplication := &LoanApplication{}
	if err := m.getCollection().FindOne(m.ctx, filter).Decode(loanApplication); err != nil {
		return "NA", err
	}

	return loanApplication.WorkflowID, nil
}

func (m *MongodbHelper) GetLoanApplicationByAppID(appID string) (*LoanApplication, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	loanApplication := &LoanApplication{}
	if err := m.getCollection().FindOne(m.ctx, filter).Decode(loanApplication); err != nil {
		return nil, err
	}

	return loanApplication, nil
}

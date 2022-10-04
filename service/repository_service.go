package service

import (
	"cadence-los-workflow/model"
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
	MongodbService struct {
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

func NewMongodbService(config MongodbConfig) *MongodbService {

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

	return &MongodbService{
		ctx:            ctx,
		getCon:         getConn,
		collectionName: loanApplicationCollectionName,
	}
}

func NewMongodbHelperWithCallBack(getDB func() *mongo.Database) *MongodbService {
	return &MongodbService{
		ctx:            context.Background(),
		getCon:         getDB,
		collectionName: loanApplicationCollectionName,
	}
}

func (m *MongodbService) getCollection() *mongo.Collection {
	return m.getCon().Collection(m.collectionName)
}

func (m *MongodbService) UpdateLoanApplicationTaskToken(appID string, lastState string, workflowID string, runID string) error {

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

func (m *MongodbService) CreateNewLoanApplication(appID string) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	if err := m.getCollection().FindOne(m.ctx, filter).Decode(&model.LoanApplication{}); err == nil {
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

func (m *MongodbService) CreateLoanApplication(loanApp *model.LoanApplication) error {

	filter := bson.M{
		"appID": bson.M{
			"$eq": loanApp.AppID,
		},
	}

	if err := m.getCollection().FindOne(m.ctx, filter).Decode(&model.LoanApplication{}); err == nil {
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

func (m *MongodbService) SaveFormOne(loanApp *model.LoanApplication) (*model.LoanApplication, error) {

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

	updatedLoanApp := model.LoanApplication{}
	m.getCollection().FindOne(m.ctx, filter).Decode(&updatedLoanApp)

	return &updatedLoanApp, err
}

func (m *MongodbService) SaveFormTwo(loanApp *model.LoanApplication) (*model.LoanApplication, error) {

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

	updatedLoanApp := model.LoanApplication{}
	m.getCollection().FindOne(m.ctx, filter).Decode(&updatedLoanApp)

	return &updatedLoanApp, err
}

func (m *MongodbService) GetWorkflowIdByAppID(appID string) (string, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	loanApplication := &model.LoanApplication{}
	if err := m.getCollection().FindOne(m.ctx, filter).Decode(loanApplication); err != nil {
		return "NA", err
	}

	return loanApplication.WorkflowID, nil
}

func (m *MongodbService) GetLoanApplicationByAppID(appID string) (*model.LoanApplication, error) {

	filter := bson.M{
		"appID": bson.M{
			"$eq": appID,
		},
	}

	loanApplication := &model.LoanApplication{}
	if err := m.getCollection().FindOne(m.ctx, filter).Decode(loanApplication); err != nil {
		return nil, err
	}

	return loanApplication, nil
}

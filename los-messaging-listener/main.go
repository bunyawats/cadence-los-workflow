package main

import (
	"cadence-los-workflow/common"
	"cadence-los-workflow/service"
	"fmt"
	"os"
)

const (
	rabbitMqUri      = "RABBITMQ_URI"
	rabbitMqInQueue  = "RABBITMQ_IN_QUEUE"
	rabbitMqOutQueue = "RABBITMQ_OUT_QUEUE"

	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

var (
	workflowHelper  common.WorkflowHelper
	rabbitMqService *service.RabbitMqService
	mongodbService  *service.MongodbService
)

func init() {

	rabbitMqService = service.NewRabbitMqService(service.RabbitMqConfig{
		RabbitMqUri:  os.Getenv(rabbitMqUri),
		InQueueName:  os.Getenv(rabbitMqInQueue),
		OutQueueName: os.Getenv(rabbitMqOutQueue),
	})

	mongodbService = service.NewMongodbService(service.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	workflowHelper.SetupServiceConfig()
}

func main() {

	fmt.Println("start messaging listener")
	rabbitMqService.ConsumeRabbitMqMessage(mongodbService, &workflowHelper)

	select {}
}

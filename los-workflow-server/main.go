package main

import (
	"cadence-los-workflow/common"
	"cadence-los-workflow/service"
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
	workflowService service.WorkflowService
)

func init() {

	mg := service.NewMongodbService(service.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	var wh common.WorkflowHelper
	wh.SetupServiceConfig()

	rb := service.NewRabbitMqService(service.RabbitMqConfig{
		RabbitMqUri:  os.Getenv(rabbitMqUri),
		InQueueName:  os.Getenv(rabbitMqInQueue),
		OutQueueName: os.Getenv(rabbitMqOutQueue),
	}, mg, &wh)

	workflowService = service.WorkflowService{
		MongodbService:  mg,
		WorkflowHelper:  &wh,
		RabbitMqService: rb,
	}
}

func main() {

	workflowService.RegisterWorkflowAndActivity()
	workflowService.StartWorkers()

	select {}
}

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
	w service.WorkflowService
)

func init() {

	r := service.NewRabbitMqService(service.RabbitMqConfig{
		RabbitMqUri:  os.Getenv(rabbitMqUri),
		InQueueName:  os.Getenv(rabbitMqInQueue),
		OutQueueName: os.Getenv(rabbitMqOutQueue),
	})

	m := service.NewMongodbService(service.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	var h common.WorkflowHelper
	h.SetupServiceConfig()

	w = service.WorkflowService{
		MongodbService:  m,
		WorkflowHelper:  &h,
		RabbitMqService: r,
	}
}

func main() {

	w.RegisterWorkflowAndActivity()
	w.StartWorkers()

	select {}
}

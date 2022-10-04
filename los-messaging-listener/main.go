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
	h common.WorkflowHelper
	r *service.RabbitMqService
	m *service.MongodbService
)

func init() {

	r = service.NewRabbitMqService(service.RabbitMqConfig{
		RabbitMqUri:  os.Getenv(rabbitMqUri),
		InQueueName:  os.Getenv(rabbitMqInQueue),
		OutQueueName: os.Getenv(rabbitMqOutQueue),
	})

	m = service.NewMongodbService(service.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	h.SetupServiceConfig()
}

func main() {

	fmt.Println("start messaging listener")
	r.ConsumeRabbitMqMessage(m, &h)

	select {}
}

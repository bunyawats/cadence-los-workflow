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
	rabbitMqService *service.RabbitMqService
)

func init() {

	mg := service.NewMongodbService(service.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	var wh common.WorkflowHelper
	wh.SetupServiceConfig()

	rabbitMqService = service.NewRabbitMqService(
		service.RabbitMqConfig{
			RabbitMqUri:  os.Getenv(rabbitMqUri),
			InQueueName:  os.Getenv(rabbitMqInQueue),
			OutQueueName: os.Getenv(rabbitMqOutQueue),
		}, mg, &wh)

}

func main() {

	fmt.Println("start messaging listener")
	rabbitMqService.ConsumeRabbitMqMessage()

	select {}
}

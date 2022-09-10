package main

import (
	"cadence-los-workflow/common"
	"fmt"
	"os"
)

const (
	rabbitMqUri   = "RABBITMQ_URI"
	rabbitMqQueue = "RABBITMQ_QUEUE"

	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

func main() {

	var h common.LosHelper

	r := common.NewRabbitMqHelper(common.RabbitMqConfig{
		RabbitMqUri:   os.Getenv(rabbitMqUri),
		RabbitMqQueue: os.Getenv(rabbitMqQueue),
	})

	m := common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	h.SetupServiceConfig()
	var err error
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	fmt.Println("start messaging listener")
	r.ConsumeRabbitMqMessage(m, workflowClient)

	select {}
}

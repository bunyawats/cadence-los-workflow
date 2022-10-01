package main

import (
	"cadence-los-workflow/common"
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
	w LosWorkFlowHelper
)

func init() {

	r := common.NewRabbitMqHelper(common.RabbitMqConfig{
		RabbitMqUri:  os.Getenv(rabbitMqUri),
		InQueueName:  os.Getenv(rabbitMqInQueue),
		OutQueueName: os.Getenv(rabbitMqOutQueue),
	})

	m := common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	var h common.LosHelper
	h.SetupServiceConfig()

	w = LosWorkFlowHelper{
		MongodbHelper:  m,
		LosHelper:      &h,
		RabbitMqHelper: r,
	}
}

func main() {

	w.RegisterWorkflowAndActivity()
	w.StartWorkers()

	select {}
}

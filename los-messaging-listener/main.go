package main

import (
	"cadence-los-workflow/common"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	cadence_client "go.uber.org/cadence/client"
	"log"
	"os"
)

const (
	rabbitMqUri   = "RABBITMQ_URI"
	rabbitMqQueue = "RABBITMQ_QUEUE"

	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

var (
	consumeChannelAmqp *amqp.Channel
	amqpConnection     *amqp.Connection

	m              *common.MongodbHelper
	h              common.LosHelper
	workflowClient cadence_client.Client
)

func init() {

	m = common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	h.SetupServiceConfig()
	var err error
	workflowClient, err = h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	amqpConnection, err = amqp.Dial(os.Getenv(rabbitMqUri))
	if err != nil {
		log.Fatalln("rabbit mq error: ", os.Getenv(rabbitMqUri), err)
	}
	log.Println("rabbit mq connected")

}

func main() {
	fmt.Println("start messaging listener")
	ConsumeRabbitMqMessage()

	select {}
}

func ConsumeRabbitMqMessage() {

	consumeChannelAmqp, _ = amqpConnection.Channel()
	msgs, _ := consumeChannelAmqp.Consume(
		os.Getenv(rabbitMqQueue),
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var request common.DEResult
			json.Unmarshal(d.Body, &request)

			common.CompleteActivity(m, workflowClient, request.AppID, request.Status)
		}
	}()
}

package main

import (
	"cadence-los-workflow/common"
	"context"
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
)

var (
	consumeChannelAmqp *amqp.Channel
	amqpConnection     *amqp.Connection

	h              common.LosHelper
	workflowClient cadence_client.Client
)

func init() {

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

			CompleteActivity(request.AppID, request.Status)
		}
	}()
}

func CompleteActivity(appID string, lastState string) {
	taskToken, err := common.GetTokenByAppID(appID)
	if err != nil {
		fmt.Printf("Failed to find taskToken by error : %+v\n", err)
	} else {

		log.Printf("AppID: %v : TaskToken %v \n", appID, taskToken)

		err = workflowClient.CompleteActivity(context.Background(), []byte(taskToken), lastState, nil)
		if err != nil {
			fmt.Printf("Failed to complete activity with error: %+v\n", err)
		} else {
			fmt.Printf("Successfully complete activity: %s\n", taskToken)
		}
	}
}

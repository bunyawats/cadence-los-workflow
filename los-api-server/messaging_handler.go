package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
)

const (
	rabbitMqUri   = "RABBITMQ_URI"
	rabbitMqQueue = "RABBITMQ_QUEUE"
)

type (
	DEResult struct {
		AppID  string `json:"appID"`
		Status string `json:"status"`
	}
)

var (
	publishChannelAmqp *amqp.Channel
	consumeChannelAmqp *amqp.Channel
	amqpConnection     *amqp.Connection
)

func init() {

	amqpConnection, err := amqp.Dial(os.Getenv(rabbitMqUri))
	if err != nil {
		log.Fatalln("rabbit mq error: ", os.Getenv(rabbitMqUri), err)
	}
	publishChannelAmqp, _ = amqpConnection.Channel()
	log.Println("rabbit mq connected")

	ConsumeRabbitMqMessage()
}

func Publish2RabbitMQ(payload *DEResult) {

	data, _ := json.Marshal(payload)
	err := publishChannelAmqp.Publish(
		"",
		os.Getenv(rabbitMqQueue),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		},
	)
	if err != nil {
		fmt.Println(err)
	}
}

func ConsumeRabbitMqMessage() {
	amqpConnection, err := amqp.Dial(os.Getenv(rabbitMqUri))
	if err != nil {
		log.Fatalln(err)
	}

	consumeChannelAmqp, _ = amqpConnection.Channel()
	msgs, err := consumeChannelAmqp.Consume(
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
			var request DEResult
			json.Unmarshal(d.Body, &request)

			CompleteActivity(request.AppID, request.Status)
		}
	}()
}

package main

import (
	"cadence-los-workflow/common"
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

	//	ConsumeRabbitMqMessage()
}

func Publish2RabbitMQ(payload *common.DEResult) {

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

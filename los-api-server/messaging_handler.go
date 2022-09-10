package main

import (
	"cadence-los-workflow/common"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
)

func Publish2RabbitMQ(payload *common.DEResult, queueName string) {

	data, _ := json.Marshal(payload)
	err := publishChannelAmqp.Publish(
		"",
		queueName,
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

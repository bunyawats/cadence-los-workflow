package main

import (
	"cadence-los-workflow/common"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"os"
)

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

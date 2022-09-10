package common

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	cadence_client "go.uber.org/cadence/client"
	"log"
	"os"
)

type (
	RabbitMqHelper struct {
		amqpConnection *amqp.Connection
		//publishChannelAmqp *amqp.Channel
		queueName string
	}

	RabbitMqConfig struct {
		RabbitMqUri   string
		RabbitMqQueue string
	}
)

func NewRabbitMqHelper(config RabbitMqConfig) *RabbitMqHelper {

	amqpConnection, err := amqp.Dial(config.RabbitMqUri)
	if err != nil {
		log.Fatalln("rabbit mq error: ", config.RabbitMqUri, err)
	}

	log.Println("rabbit mq connected")
	queueName := config.RabbitMqQueue

	return &RabbitMqHelper{
		amqpConnection: amqpConnection,
		//publishChannelAmqp: publishChannelAmqp,
		queueName: queueName,
	}
}

func (r *RabbitMqHelper) Publish2RabbitMQ(payload *DEResult) {

	data, _ := json.Marshal(payload)

	publishChannelAmqp, _ := r.amqpConnection.Channel()
	err := publishChannelAmqp.Publish(
		"",
		r.queueName,
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

func (r *RabbitMqHelper) ConsumeRabbitMqMessage(m *MongodbHelper, workflowClient cadence_client.Client) {

	consumeChannelAmqp, _ := r.amqpConnection.Channel()
	msgs, _ := consumeChannelAmqp.Consume(
		os.Getenv(r.queueName),
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

			CompleteActivity(m, workflowClient, request.AppID, request.Status)
		}
	}()
}

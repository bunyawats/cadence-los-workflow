package common

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type (
	RabbitMqHelper struct {
		amqpConnection *amqp.Connection
		queueName      string
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

	fmt.Printf("\n RabbitMqUri: %v \n RabbitMqQueue: %v\n", config.RabbitMqUri, config.RabbitMqQueue)

	return &RabbitMqHelper{
		amqpConnection: amqpConnection,
		queueName:      queueName,
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

func (r *RabbitMqHelper) ConsumeRabbitMqMessage(m *MongodbHelper, h *LosHelper) {

	consumeChannelAmqp, _ := r.amqpConnection.Channel()
	msgs, _ := consumeChannelAmqp.Consume(
		r.queueName,
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

			taskTokenStr, err := m.GetTokenByAppID(request.AppID)
			if err != nil {
				return
			}

			taskToken := DeserializeTaskToken([]byte(taskTokenStr))
			h.SignalWorkflow(
				taskToken.WorkflowID,
				SignalName,
				&SignalPayload{
					Action:  DEOneResultNotification,
					Content: Approve,
				},
			)
			time.Sleep(time.Second * 5)
			state := QueryApplicationState(m, h, request.AppID)
			fmt.Printf("current state: %v", state)
		}
	}()
}

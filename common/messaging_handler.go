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
		inQueueName    string
		outQueueName   string
	}

	RabbitMqConfig struct {
		RabbitMqUri  string
		InQueueName  string
		OutQueueName string
	}
)

func NewRabbitMqHelper(config RabbitMqConfig) *RabbitMqHelper {

	amqpConnection, err := amqp.Dial(config.RabbitMqUri)
	if err != nil {
		log.Fatalln("rabbit mq error: ", config.RabbitMqUri, err)
	}

	log.Println("rabbit mq connected")
	inQueueName := config.InQueueName
	outQueueName := config.OutQueueName

	return &RabbitMqHelper{
		amqpConnection: amqpConnection,
		inQueueName:    inQueueName,
		outQueueName:   outQueueName,
	}
}

func (r *RabbitMqHelper) PublishAppDEOne(payload *LoanApplication) {

	fmt.Println("Call PublishAppDEOne API", payload)

	data, _ := json.Marshal(payload)
	queueName := r.outQueueName
	publishMessage(r, queueName, data)
}

func (r *RabbitMqHelper) PublishDEResult(payload *DEResult) {

	data, _ := json.Marshal(payload)
	queueName := r.inQueueName
	publishMessage(r, queueName, data)
}

func publishMessage(r *RabbitMqHelper, queueName string, data []byte) {
	publishChannelAmqp, _ := r.amqpConnection.Channel()
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

func (r *RabbitMqHelper) ConsumeRabbitMqMessage(m *MongodbHelper, h *LosHelper) {

	consumeChannelAmqp, _ := r.amqpConnection.Channel()
	msgs, _ := consumeChannelAmqp.Consume(
		r.inQueueName,
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

			loanApp, err := m.GetLoanApplicationByAppID(request.AppID)
			if err != nil {
				return
			}

			//taskToken := DeserializeTaskToken([]byte(taskTokenStr))
			h.SignalWorkflow(
				loanApp.WorkflowID,
				SignalName,
				&SignalPayload{
					Action: DEOneResultNotification,
					Content: Content{
						"appID":  request.AppID,
						"status": request.Status,
					},
				},
			)
			time.Sleep(time.Second * 5)
			state := QueryApplicationState(m, h, request.AppID)
			fmt.Printf("current state: %v", state)
		}
	}()
}

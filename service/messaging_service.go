package service

import (
	"cadence-los-workflow/common"
	"cadence-los-workflow/model"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type (
	RabbitMqService struct {
		amqpConnection *amqp.Connection
		inQueueName    string
		outQueueName   string

		*MongodbService
		*common.WorkflowHelper
	}

	RabbitMqConfig struct {
		RabbitMqUri  string
		InQueueName  string
		OutQueueName string
	}
)

func NewRabbitMqService(config RabbitMqConfig, mg *MongodbService, wh *common.WorkflowHelper) *RabbitMqService {

	amqpConnection, err := amqp.Dial(config.RabbitMqUri)
	if err != nil {
		log.Fatalln("rabbit mq error: ", config.RabbitMqUri, err)
	}

	log.Println("rabbit mq connected")
	inQueueName := config.InQueueName
	outQueueName := config.OutQueueName

	return &RabbitMqService{
		amqpConnection: amqpConnection,
		inQueueName:    inQueueName,
		outQueueName:   outQueueName,

		MongodbService: mg,
		WorkflowHelper: wh,
	}
}

func (r *RabbitMqService) PublishAppDEOne(payload *model.LoanApplication) {

	fmt.Println("Call PublishAppDEOne API", payload)

	data, _ := json.Marshal(payload)
	queueName := r.outQueueName
	r.publishMessage(queueName, data)
}

func (r *RabbitMqService) PublishDEResult(payload *model.DEResult) {

	fmt.Println("Call PublishDEResult API", payload)

	data, _ := json.Marshal(payload)
	queueName := r.inQueueName
	r.publishMessage(queueName, data)
}

func (r *RabbitMqService) publishMessage(queueName string, data []byte) {
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

func (r *RabbitMqService) ConsumeRabbitMqMessage() {

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

			var dr model.DEResult
			json.Unmarshal(d.Body, &r)

			cb, _ := json.Marshal(&r)

			la, err := r.GetLoanApplicationByAppID(dr.AppID)
			if err != nil {
				return
			}

			r.SignalWorkflow(
				la.WorkflowID,
				model.SignalName,
				&model.SignalPayload{
					Action:  model.DEOneResultNotification,
					Content: cb,
				},
			)
			time.Sleep(time.Second * 5)
			//state := QueryApplicationState(m, h, r.AppID)
			//fmt.Printf("current state: %v", state)
		}
	}()
}

package service

import (
	"cadence-los-workflow/common"
	"cadence-los-workflow/model"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
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

func (s *RabbitMqService) PublishAppDEOne(payload *model.LoanApplication) {

	fmt.Println("Call PublishAppDEOne API", payload)

	data, _ := json.Marshal(payload)
	queueName := s.outQueueName
	s.publishMessage(queueName, data)
}

func (s *RabbitMqService) PublishDEResult(payload *model.DEResult) {

	fmt.Println("Call PublishDEResult API", payload)

	data, _ := json.Marshal(payload)
	queueName := s.inQueueName
	s.publishMessage(queueName, data)
}

func (s *RabbitMqService) publishMessage(queueName string, data []byte) {
	publishChannelAmqp, _ := s.amqpConnection.Channel()
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

func (s *RabbitMqService) ConsumeRabbitMqMessage() {

	log.Println("start messaging listener")

	consumeChannelAmqp, _ := s.amqpConnection.Channel()
	msgs, _ := consumeChannelAmqp.Consume(
		s.inQueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	go func() {
		for m := range msgs {
			log.Printf("Received a message: %s", m.Body)

			var deResult model.DEResult
			json.Unmarshal(m.Body, &deResult)

			cb, _ := json.Marshal(&deResult)

			la, err := s.GetLoanApplicationByAppID(deResult.AppID)

			log.Printf("Loan App: %v", la)
			if err != nil {
				log.Print(err.Error())
			} else {
				s.SignalWorkflow(
					la.WorkflowID,
					model.SignalName,
					&model.SignalPayload{
						Action:  model.DEOneResultNotification,
						Content: cb,
					},
				)
			}
		}
	}()
}

package main

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	v1 "cadence-los-workflow/los-api-server/losapis/impl/v1"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	cadence_client "go.uber.org/cadence/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
)

const (
	rabbitMqUri   = "RABBITMQ_URI"
	rabbitMqQueue = "RABBITMQ_QUEUE"

	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

var (
	publishChannelAmqp *amqp.Channel
	amqpConnection     *amqp.Connection
	queueName          string

	m              *common.MongodbHelper
	h              common.LosHelper
	workflowClient cadence_client.Client
)

func init() {

	var err error
	amqpConnection, err = amqp.Dial(os.Getenv(rabbitMqUri))
	if err != nil {
		log.Fatalln("rabbit mq error: ", os.Getenv(rabbitMqUri), err)
	}
	publishChannelAmqp, _ = amqpConnection.Channel()
	log.Println("rabbit mq connected")
	queueName = os.Getenv(rabbitMqQueue)

	m = common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	h.SetupServiceConfig()

	workflowClient, err = h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}
}

func main() {

	runGin()
	//runGrpc()
}

func runGin() error {
	router := gin.Default()

	router.POST("/nlos/notification/de_one", NlosNotificationHandler)
	router.POST("/nlos/create/application/:appId", CreateNewLoanApplicationHandler)
	router.POST("/nlos/submit/form_one/:appId", SubmitFormOneHandler)
	router.POST("/nlos/submit/form_two/:appId", SubmitFormTwoHandler)
	router.GET("/nlos/query/state/:appId", QueryStateHandler)

	log.Printf(" [*] Waiting for message. To exit press CYRL+C")
	err := router.Run(":5500")
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func runGrpc() error {
	listenOn := "127.0.0.1:8080"
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenOn, err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	los.RegisterLOSServer(
		server,
		v1.NewLosApiServer(
			context.Background(),
		),
	)

	log.Println("Listening on", listenOn)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

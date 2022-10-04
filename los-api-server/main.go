package main

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	v1 "cadence-los-workflow/los-api-server/losapis/impl/v1"
	"cadence-los-workflow/service"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
)

const (
	rabbitMqUri      = "RABBITMQ_URI"
	rabbitMqInQueue  = "RABBITMQ_IN_QUEUE"
	rabbitMqOutQueue = "RABBITMQ_OUT_QUEUE"

	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

var (
	ginHandlerHelper GinHandlerHelper
	losApiServer     *v1.LosApiServer
)

func init() {

	mg := service.NewMongodbService(service.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	var wh common.WorkflowHelper
	wh.SetupServiceConfig()

	r := service.NewRabbitMqService(service.RabbitMqConfig{
		RabbitMqUri:  os.Getenv(rabbitMqUri),
		InQueueName:  os.Getenv(rabbitMqInQueue),
		OutQueueName: os.Getenv(rabbitMqOutQueue),
	}, mg, &wh)

	var err error
	workflowClient, err := wh.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	wf := service.WorkflowService{
		MongodbService:  mg,
		WorkflowHelper:  &wh,
		RabbitMqService: r,
	}

	ginHandlerHelper = GinHandlerHelper{
		Service:       wf,
		CadenceClient: workflowClient,
	}

	losApiServer = &v1.LosApiServer{
		Context:       context.Background(),
		Service:       wf,
		CadenceClient: workflowClient,
	}

}

func main() {

	runGin()
	//runGrpc()
}

func runGin() error {

	router := gin.Default()
	ginHandlerHelper.RegisterRouter(router)

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

	s := grpc.NewServer()
	reflection.Register(s)
	los.RegisterLOSServer(s, losApiServer)

	log.Println("Listening on", listenOn)
	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC s: %w", err)
	}

	return nil
}

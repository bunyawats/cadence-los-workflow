package main

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	v1 "cadence-los-workflow/los-api-server/losapis/impl/v1"
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
	rabbitMqUri   = "RABBITMQ_URI"
	rabbitMqQueue = "RABBITMQ_QUEUE"

	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

var (
	g GinHandlerHelper
)

func init() {

	r := common.NewRabbitMqHelper(common.RabbitMqConfig{
		RabbitMqUri:   os.Getenv(rabbitMqUri),
		RabbitMqQueue: os.Getenv(rabbitMqQueue),
	})

	m := common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	var h common.LosHelper
	h.SetupServiceConfig()

	var err error
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	g = GinHandlerHelper{
		M: m,
		R: r,
		H: &h,
		W: workflowClient,
	}

}

func main() {

	runGin()
	//runGrpc()
}

func runGin() error {
	router := gin.Default()

	router.POST("/nlos/notification/de_one", g.NlosNotificationHandler)
	router.POST("/nlos/create/application/:appId", g.CreateNewLoanApplicationHandler)
	router.POST("/nlos/submit/form_one/:appId", g.SubmitFormOneHandler)
	router.POST("/nlos/submit/form_two/:appId", g.SubmitFormTwoHandler)
	router.GET("/nlos/query/state/:appId", g.QueryStateHandler)

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

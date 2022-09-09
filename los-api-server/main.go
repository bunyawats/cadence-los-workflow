package main

import (
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	v1 "cadence-los-workflow/los-api-server/losapis/impl/v1"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {

	runGin()
	//runGrpc()
}

func runGin() error {
	router := gin.Default()

	router.POST("/nlos/notification/de_one", NLOS_NotificationHandler)
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

package main

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	v1 "cadence-los-workflow/los-api-server/losapis/impl/v1"
	"cadence-los-workflow/service"
	"context"
	"flag"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkmongo "github.com/rookie-ninja/rk-db/mongodb"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"log"
)

const (
	rabbitMqUri      = "RABBITMQ_URI"
	rabbitMqInQueue  = "RABBITMQ_IN_QUEUE"
	rabbitMqOutQueue = "RABBITMQ_OUT_QUEUE"

	configName = "ssc-config"
)

var (
	rabbitMqService *service.RabbitMqService
	workflowService service.WorkflowService

	losApiServer *v1.LosApiServer
	boot         *rkboot.Boot
)

func getConfigString(name string) string {
	return rkentry.GlobalAppCtx.GetConfigEntry(configName).GetString(name)
}

func init() {

	boot = rkboot.NewBoot()

	mg := service.NewMongodbHelperWithCallBack(
		func() *mongo.Database {
			return rkmongo.GetMongoDB("ssc-mongo", "test")
		},
	)

	var wh common.WorkflowHelper
	wh.SetupServiceConfig()

	rabbitMqService = service.NewRabbitMqService(service.RabbitMqConfig{
		RabbitMqUri:  getConfigString(rabbitMqUri),
		InQueueName:  getConfigString(rabbitMqInQueue),
		OutQueueName: getConfigString(rabbitMqOutQueue),
	}, mg, &wh)

	var err error
	wc, err := wh.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	workflowService = service.WorkflowService{
		MongodbService:  mg,
		WorkflowHelper:  &wh,
		RabbitMqService: rabbitMqService,
	}

	losApiServer = &v1.LosApiServer{
		Context:         context.Background(),
		WorkflowService: workflowService,
		Client:          wc,
	}

}

func main() {

	var mode string
	flag.StringVar(&mode, "m", "all", "Mode is all or api only.")
	flag.Parse()

	if mode == "all" {
		rabbitMqService.ConsumeRabbitMqMessage()

		go func() {

			log.Println("start workflow worker")

			workflowService.RegisterWorkflowAndActivity()
			workflowService.StartWorkers()
		}()
	}

	// register grpc
	eGrpc := rkgrpc.GetGrpcEntry("los-grpc")
	eGrpc.AddRegFuncGrpc(registerLosServer)
	eGrpc.AddRegFuncGw(los.RegisterLOSHandlerFromEndpoint)

	eGw := rkgrpc.GetGrpcEntry("los-gw")
	eGw.AddRegFuncGrpc(registerLosServer)
	eGw.AddRegFuncGw(los.RegisterLOSHandlerFromEndpoint)

	boot.Bootstrap(context.TODO())
	boot.WaitForShutdownSig(context.TODO())
}

func registerLosServer(s *grpc.Server) {
	los.RegisterLOSServer(s, losApiServer)
}

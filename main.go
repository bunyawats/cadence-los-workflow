package main

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	v1 "cadence-los-workflow/los-api-server/losapis/impl/v1"
	"cadence-los-workflow/service"
	"context"
	"fmt"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkmongo "github.com/rookie-ninja/rk-db/mongodb"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

const (
	rabbitMqUri      = "RABBITMQ_URI"
	rabbitMqInQueue  = "RABBITMQ_IN_QUEUE"
	rabbitMqOutQueue = "RABBITMQ_OUT_QUEUE"

	configName = "ssc-config"
)

var (
	losApiServer *v1.LosApiServer
	boot         *rkboot.Boot
)

func getConfigString(name string) string {
	return rkentry.GlobalAppCtx.GetConfigEntry(configName).GetString(name)
}

func init() {

	boot = rkboot.NewBoot()

	r := service.NewRabbitMqService(service.RabbitMqConfig{
		RabbitMqUri:  getConfigString(rabbitMqUri),
		InQueueName:  getConfigString(rabbitMqInQueue),
		OutQueueName: getConfigString(rabbitMqOutQueue),
	})

	m := service.NewMongodbHelperWithCallBack(
		func() *mongo.Database {
			return rkmongo.GetMongoDB("ssc-mongo", "test")
		},
	)

	var h common.WorkflowHelper
	h.SetupServiceConfig()

	var err error
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	w := service.WorkflowService{
		MongodbService:  m,
		WorkflowHelper:  &h,
		RabbitMqService: r,
	}

	losApiServer = &v1.LosApiServer{
		Context:       context.Background(),
		Service:       w,
		CadenceClient: workflowClient,
	}

}

func main() {
	fmt.Println("Hello Cadence")

	// register grpc
	entry := rkgrpc.GetGrpcEntry("ssc-grpc")
	entry.AddRegFuncGrpc(registerLosServer)
	//entry.AddRegFuncGw(nil)

	boot.Bootstrap(context.TODO())
	boot.WaitForShutdownSig(context.TODO())
}

func registerLosServer(s *grpc.Server) {
	los.RegisterLOSServer(s, losApiServer)
}

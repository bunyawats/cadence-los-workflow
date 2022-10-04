package main

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	v1 "cadence-los-workflow/los-api-server/losapis/impl/v1"
	"cadence-los-workflow/service"
	"context"
	"fmt"
	_ "github.com/rookie-ninja/rk-boot/v2"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	_ "github.com/rookie-ninja/rk-db/mongodb"
	rkmongo "github.com/rookie-ninja/rk-db/mongodb"
	_ "github.com/rookie-ninja/rk-entry/v2"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	_ "github.com/rookie-ninja/rk-grpc/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.uber.org/cadence"
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

	boot *rkboot.Boot
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

	losApiServer = v1.NewLosApiServer(
		context.Background(),
		m,
		r,
		&h,
		workflowClient,
	)

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

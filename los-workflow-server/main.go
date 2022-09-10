package main

import (
	"cadence-los-workflow/common"
	"os"
)

const (
	mongoUri      = "MONGO_URI"
	mongoDatabase = "MONGO_DATABASE"
)

var (
	w LosWorkFlowHelper
)

func init() {

	m := common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	var h common.LosHelper
	h.SetupServiceConfig()

	w = LosWorkFlowHelper{
		M: m,
		H: &h,
	}
}

func main() {

	w.RegisterWorkflowAndActivity()
	w.StartWorkers()

	select {}
}

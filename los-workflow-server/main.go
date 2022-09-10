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
	m *common.MongodbHelper
	h common.LosHelper
)

func init() {

	m = common.NewMongodbHelper(common.MongodbConfig{
		MongoUri:      os.Getenv(mongoUri),
		MongoDatabase: os.Getenv(mongoDatabase),
	})

	h.SetupServiceConfig()
}

func main() {

	RegisterWorkflowAndActivity(&h)
	StartWorkers(&h)

	select {}
}

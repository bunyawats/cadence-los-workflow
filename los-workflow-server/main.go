package main

import (
	"cadence-los-workflow/common"
)

func main() {

	h := common.LosHelper{}
	h.SetupServiceConfig()

	RegisterWorkflowAndActivity(&h)
	StartWorkers(&h)

	select {}
}

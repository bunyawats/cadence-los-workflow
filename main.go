package main

import (
	"cadence-los-workflow/common"
)

func main() {

	h := common.SampleHelper{}
	h.SetupServiceConfig()

	RegisterWorkflowAndActivity(&h)
	StartWorkers(&h)

	select {}
}

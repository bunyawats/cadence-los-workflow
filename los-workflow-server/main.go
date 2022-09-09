package main

import (
	"cadence-los-workflow/common"
)

var h common.LosHelper

func init() {
	h.SetupServiceConfig()
}

func main() {

	RegisterWorkflowAndActivity(&h)
	StartWorkers(&h)

	select {}
}

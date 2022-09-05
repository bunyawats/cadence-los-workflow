package main

import (
	"flag"
	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker, trigger or shadower.")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		RegisterWorkflowAndActivity(&h)
		StartWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	}
}

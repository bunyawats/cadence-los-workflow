package main

import (
	"cadence-los-workflow/common"
	"context"
	"fmt"
	"github.com/pborman/uuid"
	cadence_client "go.uber.org/cadence/client"
	"go.uber.org/zap"
	"log"
	"time"
)

const (
	applicationName            = "loanOnBoardingGroup"
	loanOnBoardingWorkflowName = "loanOnBoardingWorkflow"
)

var (
	h              common.LosHelper
	workflowClient cadence_client.Client
)

func init() {

	h.SetupServiceConfig()
	var err error
	workflowClient, err = h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}
}

func StartWorkflow(appID string) {
	workflowOptions := cadence_client.StartWorkflowOptions{
		ID:                              "loan_on_boarding_" + uuid.New(),
		TaskList:                        applicationName,
		ExecutionStartToCloseTimeout:    10 * time.Minute,
		DecisionTaskStartToCloseTimeout: 10 * time.Minute,
	}
	execution := h.StartWorkflow(workflowOptions, loanOnBoardingWorkflowName, appID)
	log.Println("Started work flow!", zap.String("WorkflowId", execution.ID), zap.String("RunId", execution.RunID))
}

func CompleteActivity(appID string, lastState string) {
	taskToken, err := GetTokenByAppID(appID)
	if err != nil {
		fmt.Printf("Failed to find taskToken by error : %+v\n", err)
	} else {

		log.Printf("AppID: %v : TaskToken %v \n", appID, taskToken)

		err = workflowClient.CompleteActivity(context.Background(), []byte(taskToken), lastState, nil)
		if err != nil {
			fmt.Printf("Failed to complete activity with error: %+v\n", err)
		} else {
			fmt.Printf("Successfully complete activity: %s\n", taskToken)
		}
	}
}

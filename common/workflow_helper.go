package common

import (
	"context"
	"fmt"
	"github.com/pborman/uuid"
	cadence_client "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"log"
	"time"
)

func StartWorkflow(h *LosHelper) *workflow.Execution {
	workflowOptions := cadence_client.StartWorkflowOptions{
		ID:                              "los_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    20 * time.Minute,
		DecisionTaskStartToCloseTimeout: 20 * time.Minute,
	}
	execution := h.StartWorkflow(workflowOptions, LoanOnBoardingWorkflowName)
	h.Logger.Info("Started work flow!", zap.String("WorkflowId", execution.ID), zap.String("RunId", execution.RunID))
	return execution
}

func CompleteActivity(m *MongodbHelper, workflowClient cadence_client.Client, appID string, lastState string) {
	taskToken, err := m.GetWorkflowIdByAppID(appID)
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

func QueryApplicationState(m *MongodbHelper, h *LosHelper, appID string) *QueryResult {

	loanApp, err := m.GetLoanApplicationByAppID(appID)

	if err != nil {
		return nil
	}

	var result QueryResult
	err = h.ConsistentQueryWorkflow(&result, loanApp.WorkflowID, loanApp.RunID, QueryName, true)
	if err != nil {
		panic("failed to query workflow")
	}

	return &result
}

package main

import (
	"context"
	"fmt"
	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// applicationName is the task list for this sample
const (
	applicationName            = "loanOnBoardingGroup"
	loanOnBoardingWorkflowName = "loanOnBoardingWorkflow"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func StartWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}
	h.StartWorkers(h.Config.DomainName, applicationName, workerOptions)
}

func RegisterWorkflowAndActivity(h *common.SampleHelper) {

	h.RegisterWorkflowWithAlias(loanOnBoardingWorkflow, loanOnBoardingWorkflowName)

	h.RegisterActivity(createNewAppActivity)
	h.RegisterActivity(submitFormOneActivity)
	h.RegisterActivity(submitFormTwoActivity)
	h.RegisterActivity(submitDE1Activity)
	h.RegisterActivity(approveActivity)
	h.RegisterActivity(rejectActivity)
}

// helloWorkflow workflow decider
func loanOnBoardingWorkflow(ctx workflow.Context, loanAppID string) error {

	activityResult := "NA"
	lastState := "NA"

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		//HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("loan on boarding workflow started loanAppID: " + loanAppID)

	err := workflow.ExecuteActivity(ctx, createNewAppActivity, loanAppID).Get(ctx, &activityResult)
	logger.Info("\n-----submitFormOneActivity completed.-----\n", zap.String("Result", activityResult))
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, submitFormOneActivity, loanAppID).Get(ctx, &activityResult)
	logger.Info("\n-----submitFormOneActivity completed.-----\n", zap.String("Result", activityResult))
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, submitFormTwoActivity, loanAppID).Get(ctx, &lastState)
	logger.Info("\n-----submitFormTwoActivity completed.-----\n", zap.String("Result", lastState))
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, submitDE1Activity, loanAppID).Get(ctx, &lastState)
	logger.Info("\n-----submitDE1Activity completed.-----\n", zap.String("Result", lastState))
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	logger.Info(fmt.Sprintf("\n\n\n+++++before choice+++++ : %v \n\n\n", lastState))
	switch lastState {
	case "APPROVE":
		err = workflow.ExecuteActivity(ctx, approveActivity, loanAppID).Get(ctx, &activityResult)
		logger.Info("\n-----approveActivity completed.-----\n", zap.String("Result", activityResult))
	case "REJECT":
		err = workflow.ExecuteActivity(ctx, rejectActivity, loanAppID).Get(ctx, &activityResult)
		logger.Info("\n-----rejectActivity completed.-----\n", zap.String("Result", activityResult))
	}
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}
	return nil
}

func createNewAppActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++createNewAppActivity  started+++++\n " + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	UpdateLoanApplicationTaskToken(loanAppID, "NEW_APPLICATION", taskToken)

	return "SUCCESS", nil
}

func submitFormOneActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormOneActivity  started+++++\n " + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_FORM_ONE", taskToken)

	return "", activity.ErrResultPending
}

func submitFormTwoActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormTwoActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_FORM_TWO", taskToken)

	return "", activity.ErrResultPending
}

func submitDE1Activity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitDE1Activity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_DE_ONE", taskToken)

	return "", activity.ErrResultPending
}

func approveActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++approveActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	UpdateLoanApplicationTaskToken(loanAppID, "APPROVED", taskToken)

	return "SUCCESS", nil
}

func rejectActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++rejectActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	UpdateLoanApplicationTaskToken(loanAppID, "REJECTED", taskToken)

	return "SUCCESS", nil
}

package main

import (
	los_common "cadence-los-workflow/common"
	"context"
	"fmt"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	applicationName            = "loanOnBoardingGroup"
	loanOnBoardingWorkflowName = "loanOnBoardingWorkflow"
)

type (
	LosWorkFlowHelper struct {
		M *los_common.MongodbHelper
		H *los_common.LosHelper
	}
)

func (w LosWorkFlowHelper) StartWorkers() {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: w.H.WorkerMetricScope,
		Logger:       w.H.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}
	w.H.StartWorkers(w.H.Config.DomainName, applicationName, workerOptions)
}

func (w LosWorkFlowHelper) RegisterWorkflowAndActivity() {

	w.H.RegisterWorkflowWithAlias(w.loanOnBoardingWorkflow, loanOnBoardingWorkflowName)

	w.H.RegisterActivity(w.createNewAppActivity)
	w.H.RegisterActivity(w.submitFormOneActivity)
	w.H.RegisterActivity(w.submitFormTwoActivity)
	w.H.RegisterActivity(w.submitDE1Activity)
	w.H.RegisterActivity(w.approveActivity)
	w.H.RegisterActivity(w.rejectActivity)
}

// helloWorkflow workflow decider
func (w LosWorkFlowHelper) loanOnBoardingWorkflow(ctx workflow.Context, loanAppID string) error {

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

	// setup query handler for query type "state"
	err := workflow.SetQueryHandler(ctx, "state", func(input []byte) (string, error) {
		return lastState, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	err = workflow.ExecuteActivity(ctx, w.createNewAppActivity, loanAppID).Get(ctx, &activityResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}
	lastState = "NEW_APPLICATION_CREATED"
	logger.Info("\n-----createNewAppActivity completed.-----\n", zap.String("Result", lastState))

	err = workflow.ExecuteActivity(ctx, w.submitFormOneActivity, loanAppID).Get(ctx, &activityResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}
	lastState = "FORM_ONE_SUBMITTED"
	logger.Info("\n-----submitFormOneActivity completed.-----\n", zap.String("Result", lastState))

	err = workflow.ExecuteActivity(ctx, w.submitFormTwoActivity, loanAppID).Get(ctx, &activityResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}
	lastState = "FORM_TWO_SUBMITTED"
	logger.Info("\n-----submitFormTwoActivity completed.-----\n", zap.String("Result", lastState))

	err = workflow.ExecuteActivity(ctx, w.submitDE1Activity, loanAppID).Get(ctx, &activityResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}
	lastState = "DE_ONE_SUBMITTED"
	logger.Info("\n-----submitDE1Activity completed.-----\n", zap.String("Result", lastState))

	logger.Info(fmt.Sprintf("\n\n\n+++++before choice+++++ : %v \n\n\n", activityResult))
	switch activityResult {
	case "APPROVE":
		err = workflow.ExecuteActivity(ctx, w.approveActivity, loanAppID).Get(ctx, &activityResult)
		lastState = "APPLICATION_APPROVED"
		logger.Info("\n-----approveActivity completed.-----\n", zap.String("Result", lastState))
	case "REJECT":
		err = workflow.ExecuteActivity(ctx, w.rejectActivity, loanAppID).Get(ctx, &activityResult)
		lastState = "APPLICATION_REJECTED"
		logger.Info("\n-----rejectActivity completed.-----\n", zap.String("Result", lastState))
	}
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}
	return nil
}

func (w LosWorkFlowHelper) createNewAppActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++createNewAppActivity  started+++++\n " + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "NEW_APPLICATION", taskToken)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitFormOneActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormOneActivity  started+++++\n " + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_FORM_ONE", taskToken)

	return "", activity.ErrResultPending
}

func (w LosWorkFlowHelper) submitFormTwoActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormTwoActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_FORM_TWO", taskToken)

	return "", activity.ErrResultPending
}

func (w LosWorkFlowHelper) submitDE1Activity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitDE1Activity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_DE_ONE", taskToken)

	return "", activity.ErrResultPending
}

func (w LosWorkFlowHelper) approveActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++approveActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "APPROVED", taskToken)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) rejectActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++rejectActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "REJECTED", taskToken)

	return "SUCCESS", nil
}

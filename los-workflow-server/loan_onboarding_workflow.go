package main

import (
	los_common "cadence-los-workflow/common"
	"context"
	"fmt"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
	"time"

	"go.uber.org/cadence/workflow"
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
	w.H.StartWorkers(w.H.Config.DomainName, los_common.ApplicationName, workerOptions)
}

func (w LosWorkFlowHelper) RegisterWorkflowAndActivity() {

	w.H.RegisterWorkflowWithAlias(w.loanOnBoardingWorkflow, los_common.LoanOnBoardingWorkflowName)

	w.H.RegisterActivity(w.createNewAppActivity)
	w.H.RegisterActivity(w.submitFormOneActivity)
	w.H.RegisterActivity(w.submitFormTwoActivity)
	w.H.RegisterActivity(w.submitDE1Activity)
	w.H.RegisterActivity(w.approveActivity)
	w.H.RegisterActivity(w.rejectActivity)
	w.H.RegisterActivity(w.cancelActivity)
}

// helloWorkflow workflow decider
func (w LosWorkFlowHelper) loanOnBoardingWorkflow(ctx workflow.Context, loanAppID string) (los_common.State, error) {

	ch := workflow.GetSignalChannel(ctx, los_common.SignalName)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		//HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("loan on boarding workflow started loanAppID: " + loanAppID)

	state := los_common.Initialized
	activityResult := "NA"
	var content string

	err := workflow.SetQueryHandler(ctx, los_common.QueryName, func(includeContent bool) (los_common.QueryResult, error) {
		result := los_common.QueryResult{State: state}
		if includeContent {
			result.Content = content
		}
		return result, nil
	})
	if err != nil {
		return state, err
	}

	for {
		var signal los_common.SignalPayload
		if more := ch.Receive(ctx, &signal); !more {
			logger.Info("Signal channel closed")
			return state, cadence.NewCustomError("signal_channel_closed")
		}

		logger.Info("Signal received.", zap.Any("signal", signal))

		switch signal.Action {
		case los_common.Create:
			if state == los_common.Initialized {
				state = los_common.Received
				err := workflow.ExecuteActivity(ctx, w.createNewAppActivity, loanAppID).Get(ctx, &activityResult)
				if err != nil {
					logger.Error("Failed to create loan application.")
				} else {
					logger.Info("State is now created.")
					state = los_common.Created
				}

			}
		case los_common.SubmitFormOne:
			if state == los_common.Created {
				err := workflow.ExecuteActivity(ctx, w.submitFormOneActivity, loanAppID).Get(ctx, &activityResult)
				if err != nil {
					logger.Error("Failed to submit loan application form one.")
				} else {
					logger.Info("State is now form one submitted.")
					state = los_common.FormOneSubmitted
				}
			}
		case los_common.SubmitFormTwo:
			if state == los_common.FormOneSubmitted {
				err := workflow.ExecuteActivity(ctx, w.submitFormTwoActivity, loanAppID).Get(ctx, &activityResult)
				if err != nil {
					logger.Error("Failed to submit loan application form two.")
				} else {
					logger.Info("State is now form two submitted.")
					state = los_common.FormTwoSubmitted
				}
			}
		case los_common.SubmitDEOne:
			if state == los_common.FormTwoSubmitted {
				err := workflow.ExecuteActivity(ctx, w.submitDE1Activity, loanAppID).Get(ctx, &activityResult)
				if err != nil {
					logger.Error("Failed to submit DE one.")
				} else {
					logger.Info("State is now deOneSubmitted.")
					state = los_common.DEOneSubmitted
				}
			}
		case los_common.DEOneResultNotification:
			result := signal.Content
			if state == los_common.DEOneSubmitted {
				if result == los_common.Approve {
					err := workflow.ExecuteActivity(ctx, w.approveActivity, loanAppID).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to approve loan application.")
					} else {
						logger.Info("State is now approved.")
						state = los_common.Approved
					}
					return state, nil
				} else if result == los_common.Reject {
					err := workflow.ExecuteActivity(ctx, w.rejectActivity, loanAppID).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to reject loan application.")
					} else {
						logger.Info("State is now rejected.")
						state = los_common.Rejected
					}
					return state, nil
				} else {
					logger.Error(fmt.Sprintf("Wrong DE result :%v.", result))
				}
			}
		case los_common.Cancel:
			if state != los_common.Approved || state != los_common.Rejected {

				err := workflow.ExecuteActivity(ctx, w.rejectActivity, loanAppID).Get(ctx, &activityResult)
				if err != nil {
					logger.Error("Failed to reject loan application.")
				} else {
					logger.Info("State is now canceled.")
					state = los_common.Canceled
					return state, nil
				}
			}
		}
	}
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

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitFormTwoActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormTwoActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_FORM_TWO", taskToken)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitDE1Activity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitDE1Activity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "SUBMIT_DE_ONE", taskToken)

	return "SUCCESS", nil
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

func (w LosWorkFlowHelper) cancelActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++rejectActivity  started+++++\n" + loanAppID)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "CANCELED", taskToken)

	return "SUCCESS", nil
}

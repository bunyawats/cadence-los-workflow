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
	"reflect"
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
func (w LosWorkFlowHelper) loanOnBoardingWorkflow(ctx workflow.Context) (los_common.State, error) {

	ch := workflow.GetSignalChannel(ctx, los_common.SignalName)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
		//HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("loan on boarding workflow started")

	state := los_common.Initialized
	var content los_common.Content

	activityResult := "NA"

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
		logger.Info("Signal signal.Content type.", zap.Any("type", reflect.TypeOf(signal.Content)))

		content = signal.Content
		switch signal.Action {
		case los_common.Create:
			if signal.Content != nil {

				appID := signal.Content["appID"]
				if state == los_common.Initialized {
					state = los_common.Received
					err := workflow.ExecuteActivity(ctx, w.createNewAppActivity, appID).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to create loan application.")
					} else {
						logger.Info("State is now created.")
						state = los_common.Created
					}

				}
			}

		case los_common.SubmitFormOne:
			if signal.Content != nil {

				loanApp := &los_common.LoanApplication{
					AppID: signal.Content["appID"].(string),
					Fname: signal.Content["fname"].(string),
					Lname: signal.Content["lname"].(string),
				}

				if state == los_common.Created {
					err := workflow.ExecuteActivity(ctx, w.submitFormOneActivity, loanApp).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to submit loan application form one.")
					} else {
						logger.Info("State is now form one submitted.")
						state = los_common.FormOneSubmitted
					}
				}
			}

		case los_common.SubmitFormTwo:
			if signal.Content != nil {

				loanApp := &los_common.LoanApplication{
					AppID:   signal.Content["appID"].(string),
					Email:   signal.Content["email"].(string),
					PhoneNo: signal.Content["phoneNo"].(string),
				}

				if state == los_common.FormOneSubmitted {
					err := workflow.ExecuteActivity(ctx, w.submitFormTwoActivity, loanApp).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to submit loan application form two.")
					} else {
						logger.Info("State is now form two submitted.")
						state = los_common.FormTwoSubmitted
					}
				}
			}

		case los_common.SubmitDEOne:
			if signal.Content != nil {
				appID := signal.Content["appID"]
				if state == los_common.FormTwoSubmitted {
					err := workflow.ExecuteActivity(ctx, w.submitDE1Activity, appID).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to submit DE one.")
					} else {
						logger.Info("State is now deOneSubmitted.")
						state = los_common.DEOneSubmitted
					}
				}
			}

		case los_common.DEOneResultNotification:
			if signal.Content != nil {

				appID := signal.Content["appID"]
				status := signal.Content["status"]

				if state == los_common.DEOneSubmitted {
					if status == los_common.Approve {
						err := workflow.ExecuteActivity(ctx, w.approveActivity, appID).Get(ctx, &activityResult)
						if err != nil {
							logger.Error("Failed to approve loan application.")
						} else {
							logger.Info("State is now approved.")
							state = los_common.Approved
						}
						return state, nil
					} else if status == los_common.Reject {
						err := workflow.ExecuteActivity(ctx, w.rejectActivity, appID).Get(ctx, &activityResult)
						if err != nil {
							logger.Error("Failed to reject loan application.")
						} else {
							logger.Info("State is now rejected.")
							state = los_common.Rejected
						}
						return state, nil
					} else {
						logger.Error(fmt.Sprintf("Wrong DE result :%v.", status))
					}
				}
			}

		case los_common.Cancel:

			if signal.Content != nil {

				appId := signal.Content

				if state != los_common.Approved || state != los_common.Rejected {

					err := workflow.ExecuteActivity(ctx, w.rejectActivity, appId).Get(ctx, &activityResult)
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
}

func (w LosWorkFlowHelper) createNewAppActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++createNewAppActivity  started+++++\n " + loanAppID)

	if err := w.M.CreateNewLoanApplication(loanAppID); err != nil {
		return "FAIL", err
	}

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanAppID, "NEW_APPLICATION", taskToken)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitFormOneActivity(ctx context.Context, loanApp *los_common.LoanApplication) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormOneActivity  started+++++\n " + loanApp.AppID)

	_, err := w.M.SaveFormOne(loanApp)
	if err != nil {
		return "FAIL", err
	}
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanApp.AppID, "SUBMIT_FORM_ONE", taskToken)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitFormTwoActivity(ctx context.Context, loanApp *los_common.LoanApplication) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormTwoActivity  started+++++\n" + loanApp.AppID)

	_, err := w.M.SaveFormTwo(loanApp)
	if err != nil {
		return "FAIL", err
	}
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)

	w.M.UpdateLoanApplicationTaskToken(loanApp.AppID, "SUBMIT_FORM_TWO", taskToken)

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

package main

import (
	loscommon "cadence-los-workflow/common"
	"context"
	"encoding/json"
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
		MongodbHelper  *loscommon.MongodbHelper
		LosHelper      *loscommon.LosHelper
		RabbitMqHelper *loscommon.RabbitMqHelper
	}
)

func (w LosWorkFlowHelper) StartWorkers() {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: w.LosHelper.WorkerMetricScope,
		Logger:       w.LosHelper.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}
	w.LosHelper.StartWorkers(w.LosHelper.Config.DomainName, loscommon.ApplicationName, workerOptions)
}

func (w LosWorkFlowHelper) RegisterWorkflowAndActivity() {

	w.LosHelper.RegisterWorkflowWithAlias(w.loanOnBoardingWorkflow, loscommon.LoanOnBoardingWorkflowName)

	w.LosHelper.RegisterActivity(w.createNewAppActivity)
	w.LosHelper.RegisterActivity(w.submitFormOneActivity)
	w.LosHelper.RegisterActivity(w.submitFormTwoActivity)
	w.LosHelper.RegisterActivity(w.submitDE1Activity)
	w.LosHelper.RegisterActivity(w.approveActivity)
	w.LosHelper.RegisterActivity(w.rejectActivity)
	w.LosHelper.RegisterActivity(w.cancelActivity)
}

func (w LosWorkFlowHelper) updateCurrentState(ctx workflow.Context, loanAppID string, state string) {
	info := workflow.GetInfo(ctx)
	workflowId := info.WorkflowExecution.ID
	runID := info.WorkflowExecution.RunID
	_ = w.MongodbHelper.UpdateLoanApplicationTaskToken(loanAppID, state, workflowId, runID)
}

func (w LosWorkFlowHelper) loanOnBoardingWorkflow(ctx workflow.Context) (loscommon.State, error) {

	ch := workflow.GetSignalChannel(ctx, loscommon.SignalName)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
		//HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("loan on boarding workflow started")

	state := loscommon.Initialized
	var content loscommon.Content

	activityResult := "NA"

	err := workflow.SetQueryHandler(ctx, loscommon.QueryName, func(includeContent bool) (loscommon.QueryResult, error) {
		result := loscommon.QueryResult{State: state}
		if includeContent {
			result.Content = content
		}
		return result, nil
	})
	if err != nil {
		return state, err
	}

	for {
		var signal loscommon.SignalPayload
		if more := ch.Receive(ctx, &signal); !more {
			logger.Info("Signal channel closed")
			return state, cadence.NewCustomError("signal_channel_closed")
		}

		logger.Info("Signal received.", zap.Any("signal", signal))
		logger.Info("Signal signal.Content type.", zap.Any("type", reflect.TypeOf(signal.Content)))

		content = signal.Content
		switch signal.Action {
		case loscommon.Create:
			if signal.Content != nil {

				var appID string
				json.Unmarshal(content, &appID)

				if state == loscommon.Initialized {
					state = loscommon.Received
					err := workflow.ExecuteActivity(ctx, w.createNewAppActivity, appID).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to create loan application.")
					} else {
						logger.Info("State is now created.")
						state = loscommon.Created

						w.updateCurrentState(ctx, appID, string(state))
					}

				}
			}

		case loscommon.SubmitFormOne:
			if signal.Content != nil {

				var la loscommon.LoanApplication
				json.Unmarshal(content, &la)

				if state == loscommon.Created {
					err := workflow.ExecuteActivity(ctx, w.submitFormOneActivity, la).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to submit loan application form one.")
					} else {
						logger.Info("State is now form one submitted.")
						state = loscommon.FormOneSubmitted

						w.updateCurrentState(ctx, la.AppID, string(state))
					}
				}
			}

		case loscommon.SubmitFormTwo:
			if signal.Content != nil {

				var la loscommon.LoanApplication
				json.Unmarshal(content, &la)

				if state == loscommon.FormOneSubmitted {
					err := workflow.ExecuteActivity(ctx, w.submitFormTwoActivity, la).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to submit loan application form two.")
					} else {
						logger.Info("State is now form two submitted.")
						state = loscommon.FormTwoSubmitted

						w.updateCurrentState(ctx, la.AppID, string(state))
					}
				}
			}

		case loscommon.SubmitDEOne:
			if signal.Content != nil {

				var appID string
				json.Unmarshal(content, &appID)

				if state == loscommon.FormTwoSubmitted {
					err := workflow.ExecuteActivity(ctx, w.submitDE1Activity, appID).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to submit DE one.")
					} else {
						logger.Info("State is now deOneSubmitted.")
						state = loscommon.DEOneSubmitted

						w.updateCurrentState(ctx, appID, string(state))
					}
				}
			}

		case loscommon.DEOneResultNotification:
			if signal.Content != nil {

				var r loscommon.DEResult
				json.Unmarshal(content, &r)

				if state == loscommon.DEOneSubmitted {
					if r.Status == loscommon.Approve {
						err := workflow.ExecuteActivity(ctx, w.approveActivity, r.AppID).Get(ctx, &activityResult)
						if err != nil {
							logger.Error("Failed to approve loan application.")
						} else {
							logger.Info("State is now approved.")
							state = loscommon.Approved

							w.updateCurrentState(ctx, r.AppID, string(state))
						}
						return state, nil
					} else if r.Status == loscommon.Reject {
						err := workflow.ExecuteActivity(ctx, w.rejectActivity, r.AppID).Get(ctx, &activityResult)
						if err != nil {
							logger.Error("Failed to reject loan application.")
						} else {
							logger.Info("State is now rejected.")
							state = loscommon.Rejected

							w.updateCurrentState(ctx, r.AppID, string(state))
						}
						return state, nil
					} else {
						logger.Error(fmt.Sprintf("Wrong DE result :%v.", r.Status))
					}
				}
			}

		case loscommon.Cancel:

			if signal.Content != nil {

				var appID string
				json.Unmarshal(content, &appID)

				if state != loscommon.Approved && state != loscommon.Rejected {

					err := workflow.ExecuteActivity(ctx, w.rejectActivity, appID).Get(ctx, &activityResult)
					if err != nil {
						logger.Error("Failed to reject loan application.")
					} else {
						logger.Info("State is now canceled.")
						state = loscommon.Canceled

						w.updateCurrentState(ctx, appID, string(state))

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

	if err := w.MongodbHelper.CreateNewLoanApplication(loanAppID); err != nil {
		return "FAIL", err
	}

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitFormOneActivity(ctx context.Context, loanApp *loscommon.LoanApplication) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormOneActivity  started+++++\n " + loanApp.AppID)

	_, err := w.MongodbHelper.SaveFormOne(loanApp)
	if err != nil {
		return "FAIL", err
	}

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitFormTwoActivity(ctx context.Context, loanApp *loscommon.LoanApplication) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormTwoActivity  started+++++\n" + loanApp.AppID)

	_, err := w.MongodbHelper.SaveFormTwo(loanApp)
	if err != nil {
		return "FAIL", err
	}

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) submitDE1Activity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitDE1Activity  started+++++\n" + loanAppID)

	loanApp, err := w.MongodbHelper.GetLoanApplicationByAppID(loanAppID)
	if err != nil {
		return "FAIL", err
	}

	w.RabbitMqHelper.PublishAppDEOne(loanApp)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) approveActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++approveActivity  started+++++\n" + loanAppID)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) rejectActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++rejectActivity  started+++++\n" + loanAppID)

	return "SUCCESS", nil
}

func (w LosWorkFlowHelper) cancelActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++rejectActivity  started+++++\n" + loanAppID)

	return "SUCCESS", nil
}

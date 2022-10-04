package service

import (
	"cadence-los-workflow/common"
	"cadence-los-workflow/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
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
	WorkflowService struct {
		MongodbService  *MongodbService
		WorkflowHelper  *common.WorkflowHelper
		RabbitMqService *RabbitMqService
	}
)

func (w WorkflowService) StartWorkers() {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: w.WorkflowHelper.WorkerMetricScope,
		Logger:       w.WorkflowHelper.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}
	w.WorkflowHelper.StartWorkers(w.WorkflowHelper.Config.DomainName, model.ApplicationName, workerOptions)
}

func (w WorkflowService) checkAllowState(currentState model.State, allowState ...model.State) bool {

	ok := false
	for _, as := range allowState {
		if currentState == as {
			ok = true
			break
		}
	}
	return ok
}

func (w WorkflowService) RegisterWorkflowAndActivity() {

	w.WorkflowHelper.RegisterWorkflowWithAlias(w.loanOnBoardingWorkflow, model.LoanOnBoardingWorkflowName)

	w.WorkflowHelper.RegisterActivity(w.createNewAppActivity)
	w.WorkflowHelper.RegisterActivity(w.submitFormOneActivity)
	w.WorkflowHelper.RegisterActivity(w.submitFormTwoActivity)
	w.WorkflowHelper.RegisterActivity(w.submitDE1Activity)
	w.WorkflowHelper.RegisterActivity(w.approveActivity)
	w.WorkflowHelper.RegisterActivity(w.rejectActivity)
	w.WorkflowHelper.RegisterActivity(w.cancelActivity)
}

func (w WorkflowService) updateCurrentState(ctx workflow.Context, loanAppID string, state string) {
	info := workflow.GetInfo(ctx)
	workflowId := info.WorkflowExecution.ID
	runID := info.WorkflowExecution.RunID
	_ = w.MongodbService.UpdateLoanApplicationTaskToken(loanAppID, state, workflowId, runID)
}

func (w WorkflowService) loanOnBoardingWorkflow(ctx workflow.Context) (model.State, error) {

	ch := workflow.GetSignalChannel(ctx, model.SignalName)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
		//HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("loan on boarding workflow started")

	state := model.Initialized
	var content model.Content

	activityResult := "NA"

	err := workflow.SetQueryHandler(ctx, model.QueryName, func(includeContent bool) (model.QueryResult, error) {
		result := model.QueryResult{State: state}
		if includeContent {
			result.Content = content
		}
		return result, nil
	})
	if err != nil {
		return state, err
	}

	for {
		var signal model.SignalPayload
		if more := ch.Receive(ctx, &signal); !more {
			logger.Info("Signal channel closed")
			return state, cadence.NewCustomError("signal_channel_closed")
		}

		logger.Info("Signal received.", zap.Any("signal", signal))
		logger.Info("Signal signal.Content type.", zap.Any("type", reflect.TypeOf(signal.Content)))

		content = signal.Content
		switch signal.Action {
		case model.Create:

			if !w.checkAllowState(state, model.Initialized) ||
				signal.Content == nil {
				continue
			}

			var appID string
			json.Unmarshal(content, &appID)

			state = model.Received
			err := workflow.ExecuteActivity(ctx, w.createNewAppActivity, appID).Get(ctx, &activityResult)
			if err != nil {
				logger.Error("Failed to create loan application.")
			} else {

				state = model.Created
				w.updateCurrentState(ctx, appID, string(state))
			}

		case model.SubmitFormOne:

			if !w.checkAllowState(state, model.Created) ||
				signal.Content == nil {
				continue
			}

			var la model.LoanApplication
			json.Unmarshal(content, &la)

			err := workflow.ExecuteActivity(ctx, w.submitFormOneActivity, la).Get(ctx, &activityResult)
			if err != nil {
				logger.Error("Failed to submit loan application form one.")
			} else {

				state = model.FormOneSubmitted
				w.updateCurrentState(ctx, la.AppID, string(state))
			}

		case model.SubmitFormTwo:

			if !w.checkAllowState(state, model.FormOneSubmitted) ||
				signal.Content == nil {
				continue
			}

			var la model.LoanApplication
			json.Unmarshal(content, &la)

			err := workflow.ExecuteActivity(ctx, w.submitFormTwoActivity, la).Get(ctx, &activityResult)
			if err != nil {
				logger.Error("Failed to submit loan application form two.")
			} else {

				state = model.FormTwoSubmitted
				w.updateCurrentState(ctx, la.AppID, string(state))
			}

		case model.SubmitDEOne:

			if !w.checkAllowState(state, model.FormTwoSubmitted) ||
				signal.Content == nil {
				continue
			}

			var appID string
			json.Unmarshal(content, &appID)

			err := workflow.ExecuteActivity(ctx, w.submitDE1Activity, appID).Get(ctx, &activityResult)
			if err != nil {
				logger.Error("Failed to submit DE one.")
			} else {

				state = model.DEOneSubmitted
				w.updateCurrentState(ctx, appID, string(state))
			}

		case model.DEOneResultNotification:

			if !w.checkAllowState(state, model.DEOneSubmitted) ||
				signal.Content == nil {
				continue
			}

			var r model.DEResult
			json.Unmarshal(content, &r)

			if r.Status == model.Approve {
				err := workflow.ExecuteActivity(ctx, w.approveActivity, r.AppID).Get(ctx, &activityResult)
				if err != nil {
					logger.Error("Failed to approve loan application.")
				} else {

					state = model.Approved

					w.updateCurrentState(ctx, r.AppID, string(state))
				}
				return state, nil
			}
			if r.Status == model.Reject {
				err := workflow.ExecuteActivity(ctx, w.rejectActivity, r.AppID).Get(ctx, &activityResult)
				if err != nil {
					logger.Error("Failed to reject loan application.")
				} else {

					state = model.Rejected
					w.updateCurrentState(ctx, r.AppID, string(state))
				}
				return state, nil
			}
			logger.Error(fmt.Sprintf("Wrong DE result :%v.", r.Status))

		case model.Cancel:

			if !w.checkAllowState(state, model.Approved, model.Rejected) ||
				signal.Content == nil {
				continue
			}

			var appID string
			json.Unmarshal(content, &appID)

			err := workflow.ExecuteActivity(ctx, w.rejectActivity, appID).Get(ctx, &activityResult)
			if err != nil {
				logger.Error("Failed to reject loan application.")
			} else {

				state = model.Canceled
				w.updateCurrentState(ctx, appID, string(state))

				return state, nil
			}
		}

		logger.Info(fmt.Sprintf("State is now %v.", state))
	}
}

func (w WorkflowService) createNewAppActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++createNewAppActivity  started+++++\n " + loanAppID)

	if err := w.MongodbService.CreateNewLoanApplication(loanAppID); err != nil {
		return "FAIL", err
	}

	return "SUCCESS", nil
}

func (w WorkflowService) submitFormOneActivity(ctx context.Context, loanApp *model.LoanApplication) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormOneActivity  started+++++\n " + loanApp.AppID)

	_, err := w.MongodbService.SaveFormOne(loanApp)
	if err != nil {
		return "FAIL", err
	}

	return "SUCCESS", nil
}

func (w WorkflowService) submitFormTwoActivity(ctx context.Context, loanApp *model.LoanApplication) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitFormTwoActivity  started+++++\n" + loanApp.AppID)

	_, err := w.MongodbService.SaveFormTwo(loanApp)
	if err != nil {
		return "FAIL", err
	}

	return "SUCCESS", nil
}

func (w WorkflowService) submitDE1Activity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++submitDE1Activity  started+++++\n" + loanAppID)

	loanApp, err := w.MongodbService.GetLoanApplicationByAppID(loanAppID)
	if err != nil {
		return "FAIL", err
	}

	w.RabbitMqService.PublishAppDEOne(loanApp)

	return "SUCCESS", nil
}

func (w WorkflowService) approveActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++approveActivity  started+++++\n" + loanAppID)

	return "SUCCESS", nil
}

func (w WorkflowService) rejectActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++rejectActivity  started+++++\n" + loanAppID)

	return "SUCCESS", nil
}

func (w WorkflowService) cancelActivity(ctx context.Context, loanAppID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("\n\n+++++rejectActivity  started+++++\n" + loanAppID)

	return "SUCCESS", nil
}

func (w WorkflowService) StartWorkflow() *workflow.Execution {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "los_" + uuid.New(),
		TaskList:                        model.ApplicationName,
		ExecutionStartToCloseTimeout:    20 * time.Minute,
		DecisionTaskStartToCloseTimeout: 20 * time.Minute,
	}
	execution := w.WorkflowHelper.StartWorkflow(workflowOptions, model.LoanOnBoardingWorkflowName)
	w.WorkflowHelper.Logger.Info("Started work flow!", zap.String("WorkflowId", execution.ID), zap.String("RunId", execution.RunID))
	return execution
}

func (w WorkflowService) QueryApplicationState(appID string) *model.QueryResult {

	loanApp, err := w.MongodbService.GetLoanApplicationByAppID(appID)

	if err != nil {
		return nil
	}

	var result model.QueryResult
	err = w.WorkflowHelper.ConsistentQueryWorkflow(&result, loanApp.WorkflowID, loanApp.RunID, model.QueryName, true)
	if err != nil {
		panic("failed to query workflow")
	}

	return &result
}

func AssertState(expected, actual model.State) {
	if expected != actual {
		message := fmt.Sprintf("Workflow in wrong state. Expected %v Actual %v", expected, actual)
		panic(message)
	}
}

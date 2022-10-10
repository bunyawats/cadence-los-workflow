package v1

import (
	"cadence-los-workflow/losapis/gen/v1"
	"cadence-los-workflow/model"
	"cadence-los-workflow/service"
	"context"
	"encoding/json"

	"go.uber.org/cadence/client"
	"time"
)

type LosApiServer struct {
	Context context.Context
	client.Client
	service.WorkflowService
}

func (s *LosApiServer) CreateNewApp(_ context.Context, in *los.CreateNewAppRequest) (*los.CreateNewAppResponse, error) {

	appID := in.AppID

	cb, _ := json.Marshal(appID)

	ex := s.WorkflowService.StartWorkflow()
	s.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := s.QueryApplicationState(appID)
	service.AssertState(model.Created, st.State)

	return &los.CreateNewAppResponse{
		AppID: appID,
	}, nil
}

func (s *LosApiServer) SubmitFormOne(_ context.Context, in *los.SubmitFormOneRequest) (*los.SubmitFormOneResponse, error) {

	r := model.LoanApplication{
		AppID: in.AppID,
		Fname: in.FName,
		Lname: in.LName,
	}

	id, err := s.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(&r)

	s.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := s.QueryApplicationState(r.AppID)
	service.AssertState(model.FormOneSubmitted, st.State)

	return &los.SubmitFormOneResponse{
		LoanApp: &los.LoanApplication{
			AppID: r.AppID,
			FName: r.Fname,
			LName: r.Lname,
		},
	}, nil
}

func (s *LosApiServer) SubmitFormTwo(_ context.Context, in *los.SubmitFormTwoRequest) (*los.SubmitFormTwoResponse, error) {

	r := model.LoanApplication{
		AppID:   in.AppID,
		Email:   in.Email,
		PhoneNo: in.PhoneNo,
	}

	id, err := s.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(&r)

	s.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := s.QueryApplicationState(r.AppID)
	service.AssertState(model.FormTwoSubmitted, st.State)

	return &los.SubmitFormTwoResponse{
		LoanApp: &los.LoanApplication{
			AppID:   r.AppID,
			FName:   r.Fname,
			LName:   r.Lname,
			Email:   r.Email,
			PhoneNo: r.PhoneNo,
		},
	}, nil
}

func (s *LosApiServer) SubmitDeOne(_ context.Context, in *los.SubmitDeOneRequest) (*los.SubmitDeOneResponse, error) {

	id, err := s.MongodbService.GetWorkflowIdByAppID(in.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(in.AppID)

	s.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitDEOne,
			Content: cb,
		},
	)

	time.Sleep(time.Second)
	st := s.QueryApplicationState(in.AppID)
	service.AssertState(model.DEOneSubmitted, st.State)

	return &los.SubmitDeOneResponse{
		AppID: in.AppID,
	}, nil
}

func (s *LosApiServer) NotificationDE1(_ context.Context, in *los.NotificationDE1Request) (*los.NotificationDE1Response, error) {

	s.PublishDEResult(&model.DEResult{
		AppID:  in.AppID,
		Status: in.Status,
	})

	return &los.NotificationDE1Response{
		AppID:  in.AppID,
		Status: in.Status,
	}, nil
}

func (s *LosApiServer) QueryState(_ context.Context, in *los.QueryStateRequest) (*los.QueryStateResponse, error) {

	st := s.QueryApplicationState(in.AppID)

	return &los.QueryStateResponse{
		LoanAppState: &los.LoanAppState{
			AppID: in.AppID,
			State: string(st.State),
		},
	}, nil
}

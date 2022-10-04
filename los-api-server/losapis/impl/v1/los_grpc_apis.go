package v1

import (
	v1 "cadence-los-workflow/los-api-server/losapis/gen/v1"
	"cadence-los-workflow/model"
	"cadence-los-workflow/service"
	"context"
	"encoding/json"

	"go.uber.org/cadence/client"
	"time"
)

type LosApiServer struct {
	v1.UnimplementedLOSServer
	Context       context.Context
	CadenceClient client.Client
	Service       service.WorkflowService
}

func (s *LosApiServer) CreateNewApp(_ context.Context, in *v1.CreateNewAppRequest) (*v1.CreateNewAppResponse, error) {

	appID := in.AppID

	cb, _ := json.Marshal(appID)

	ex := s.Service.StartWorkflow()
	s.Service.WorkflowHelper.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := s.Service.QueryApplicationState(appID)
	service.AssertState(model.Created, st.State)

	return &v1.CreateNewAppResponse{
		AppID: appID,
	}, nil
}

func (s *LosApiServer) SubmitFormOne(_ context.Context, in *v1.SubmitFormOneRequest) (*v1.SubmitFormOneResponse, error) {

	r := model.LoanApplication{
		AppID: in.AppID,
		Fname: in.FName,
		Lname: in.LName,
	}

	id, err := s.Service.MongodbService.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(&r)

	s.Service.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := s.Service.QueryApplicationState(r.AppID)
	service.AssertState(model.FormOneSubmitted, st.State)

	return &v1.SubmitFormOneResponse{
		LoanApp: &v1.LoanApplication{
			AppID: r.AppID,
			FName: r.Fname,
			LName: r.Lname,
		},
	}, nil
}

func (s *LosApiServer) SubmitFormTwo(_ context.Context, in *v1.SubmitFormTwoRequest) (*v1.SubmitFormTwoResponse, error) {

	r := model.LoanApplication{
		AppID:   in.AppID,
		Email:   in.Email,
		PhoneNo: in.PhoneNo,
	}

	id, err := s.Service.MongodbService.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(&r)

	s.Service.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := s.Service.QueryApplicationState(r.AppID)
	service.AssertState(model.FormTwoSubmitted, st.State)

	return &v1.SubmitFormTwoResponse{
		LoanApp: &v1.LoanApplication{
			AppID:   r.AppID,
			FName:   r.Fname,
			LName:   r.Lname,
			Email:   r.Email,
			PhoneNo: r.PhoneNo,
		},
	}, nil
}

func (s *LosApiServer) SubmitDeOne(_ context.Context, in *v1.SubmitDeOneRequest) (*v1.SubmitDeOneResponse, error) {

	id, err := s.Service.MongodbService.GetWorkflowIdByAppID(in.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(in.AppID)

	s.Service.WorkflowHelper.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitDEOne,
			Content: cb,
		},
	)

	time.Sleep(time.Second)
	st := s.Service.QueryApplicationState(in.AppID)
	service.AssertState(model.DEOneSubmitted, st.State)

	return &v1.SubmitDeOneResponse{
		AppID: in.AppID,
	}, nil
}

func (s *LosApiServer) NotificationDE1(_ context.Context, in *v1.NotificationDE1Request) (*v1.NotificationDE1Response, error) {

	s.Service.RabbitMqService.PublishDEResult(&model.DEResult{
		AppID:  in.AppID,
		Status: in.Status,
	})

	return &v1.NotificationDE1Response{
		AppID:  in.AppID,
		Status: in.Status,
	}, nil
}

func (s *LosApiServer) QueryState(_ context.Context, in *v1.QueryStateRequest) (*v1.QueryStateResponse, error) {

	st := s.Service.QueryApplicationState(in.AppID)

	return &v1.QueryStateResponse{
		LoanAppState: &v1.LoanAppState{
			AppID: in.AppID,
			State: string(st.State),
		},
	}, nil
}

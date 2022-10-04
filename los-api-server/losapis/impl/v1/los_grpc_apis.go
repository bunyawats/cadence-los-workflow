package v1

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	"cadence-los-workflow/model"
	"cadence-los-workflow/service"
	"context"
	"encoding/json"
	cadence_client "go.uber.org/cadence/client"
	"time"
)

type LosApiServer struct {
	los.UnimplementedLOSServer
	Context context.Context

	M *service.MongodbService
	R *service.RabbitMqService
	H *common.WorkflowHelper
	W cadence_client.Client
}

func NewLosApiServer(
	cx context.Context,
	m *service.MongodbService,
	r *service.RabbitMqService,
	h *common.WorkflowHelper,
	w cadence_client.Client,
) *LosApiServer {

	return &LosApiServer{
		Context: cx,
		M:       m,
		R:       r,
		H:       h,
		W:       w,
	}
}

func (s *LosApiServer) CreateNewApp(_ context.Context, in *los.CreateNewAppRequest) (*los.CreateNewAppResponse, error) {

	appID := in.AppID

	cb, _ := json.Marshal(appID)

	ex := service.StartWorkflow(s.H)
	s.H.SignalWorkflow(
		ex.ID,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.Create,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := service.QueryApplicationState(s.M, s.H, appID)
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

	id, err := s.M.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(&r)

	s.H.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormOne,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := service.QueryApplicationState(s.M, s.H, r.AppID)
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

	id, err := s.M.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(&r)

	s.H.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitFormTwo,
			Content: cb,
		},
	)
	time.Sleep(time.Second)
	st := service.QueryApplicationState(s.M, s.H, r.AppID)
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

	id, err := s.M.GetWorkflowIdByAppID(in.AppID)
	if err != nil {
		return nil, err
	}

	cb, _ := json.Marshal(in.AppID)

	s.H.SignalWorkflow(
		id,
		model.SignalName,
		&model.SignalPayload{
			Action:  model.SubmitDEOne,
			Content: cb,
		},
	)

	time.Sleep(time.Second)
	st := service.QueryApplicationState(s.M, s.H, in.AppID)
	service.AssertState(model.DEOneSubmitted, st.State)

	return &los.SubmitDeOneResponse{
		AppID: in.AppID,
	}, nil
}

func (s *LosApiServer) NotificationDE1(_ context.Context, in *los.NotificationDE1Request) (*los.NotificationDE1Response, error) {

	s.R.PublishDEResult(&model.DEResult{
		AppID:  in.AppID,
		Status: in.Status,
	})

	return &los.NotificationDE1Response{
		AppID:  in.AppID,
		Status: in.Status,
	}, nil
}

func (s *LosApiServer) QueryState(_ context.Context, in *los.QueryStateRequest) (*los.QueryStateResponse, error) {

	st := service.QueryApplicationState(s.M, s.H, in.AppID)

	return &los.QueryStateResponse{
		LoanAppState: &los.LoanAppState{
			AppID: in.AppID,
			State: string(st.State),
		},
	}, nil
}

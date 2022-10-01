package v1

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	"context"
	cadence_client "go.uber.org/cadence/client"
	"time"
)

type LosApiServer struct {
	los.UnimplementedLOSServer
	Context context.Context

	M *common.MongodbHelper
	R *common.RabbitMqHelper
	H *common.LosHelper
	W cadence_client.Client
}

func NewLosApiServer(
	cx context.Context,
	m *common.MongodbHelper,
	r *common.RabbitMqHelper,
	h *common.LosHelper,
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

	ex := common.StartWorkflow(s.H)
	s.H.SignalWorkflow(
		ex.ID,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.Create,
			Content: common.Content{"appID": appID},
		},
	)
	time.Sleep(time.Second)
	state := common.QueryApplicationState(s.M, s.H, appID)
	common.AssertState(common.Created, state.State)

	return &los.CreateNewAppResponse{
		AppID: appID,
	}, nil
}

func (s *LosApiServer) SubmitFormOne(_ context.Context, in *los.SubmitFormOneRequest) (*los.SubmitFormOneResponse, error) {

	r := common.LoanApplication{
		AppID: in.AppID,
		Fname: in.FName,
		Lname: in.LName,
	}

	id, err := s.M.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	s.H.SignalWorkflow(
		id,
		common.SignalName,
		&common.SignalPayload{
			Action: common.SubmitFormOne,
			Content: common.Content{
				"appID": r.AppID,
				"fname": r.Fname,
				"lname": r.Lname,
			},
		},
	)
	time.Sleep(time.Second)
	state := common.QueryApplicationState(s.M, s.H, r.AppID)
	common.AssertState(common.FormOneSubmitted, state.State)

	return &los.SubmitFormOneResponse{
		LoanApp: &los.LoanApplication{
			AppID: r.AppID,
			FName: r.Fname,
			LName: r.Lname,
		},
	}, nil
}

func (s *LosApiServer) SubmitFormTwo(_ context.Context, in *los.SubmitFormTwoRequest) (*los.SubmitFormTwoResponse, error) {

	r := common.LoanApplication{
		AppID:   in.AppID,
		Email:   in.Email,
		PhoneNo: in.PhoneNo,
	}

	id, err := s.M.GetWorkflowIdByAppID(r.AppID)
	if err != nil {
		return nil, err
	}

	s.H.SignalWorkflow(
		id,
		common.SignalName,
		&common.SignalPayload{
			Action: common.SubmitFormTwo,
			Content: common.Content{
				"appID":   r.AppID,
				"email":   r.Email,
				"phoneNo": r.PhoneNo,
			},
		},
	)
	time.Sleep(time.Second)
	state := common.QueryApplicationState(s.M, s.H, r.AppID)
	common.AssertState(common.FormTwoSubmitted, state.State)

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

	s.H.SignalWorkflow(
		id,
		common.SignalName,
		&common.SignalPayload{
			Action:  common.SubmitDEOne,
			Content: common.Content{"appID": in.AppID},
		},
	)

	time.Sleep(time.Second)
	state := common.QueryApplicationState(s.M, s.H, in.AppID)
	common.AssertState(common.DEOneSubmitted, state.State)

	return &los.SubmitDeOneResponse{
		AppID: in.AppID,
	}, nil
}

func (s *LosApiServer) NotificationDE1(_ context.Context, in *los.NotificationDE1Request) (*los.NotificationDE1Response, error) {

	s.R.PublishDEResult(&common.DEResult{
		AppID:  in.AppID,
		Status: in.Status,
	})

	return &los.NotificationDE1Response{
		AppID:  in.AppID,
		Status: in.Status,
	}, nil
}

func (s *LosApiServer) QueryState(_ context.Context, in *los.QueryStateRequest) (*los.QueryStateResponse, error) {

	queryResult := common.QueryApplicationState(s.M, s.H, in.AppID)

	return &los.QueryStateResponse{
		LoanAppState: &los.LoanAppState{
			AppID: in.AppID,
			State: string(queryResult.State),
		},
	}, nil
}

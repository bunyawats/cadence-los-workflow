package v1

import (
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	"context"
)

type LosApiServer struct {
	los.UnimplementedLOSServer
	Context context.Context
}

func NewLosApiServer(cx context.Context) *LosApiServer {
	return &LosApiServer{
		Context: cx,
	}
}

func (s *LosApiServer) CreateNewApp(ctx context.Context, in *los.CreateNewAppRequest) (*los.CreateNewAppResponse, error) {

	appId := in.AppID

	return &los.CreateNewAppResponse{
		AppID: appId,
	}, nil
}

func (s *LosApiServer) SubmitFormOne(ctx context.Context, in *los.SubmitFormOneRequest) (*los.SubmitFormOneResponse, error) {

	appId := in.AppID

	return &los.SubmitFormOneResponse{
		LoanApp: &los.LoanApplication{
			AppID: appId,
			FName: in.FName,
			LName: in.LnNme,
		},
	}, nil
}

func (s *LosApiServer) SubmitFormTwo(ctx context.Context, in *los.SubmitFormTwoRequest) (*los.SubmitFormTwoResponse, error) {

	appId := in.AppID

	return &los.SubmitFormTwoResponse{
		LoanApp: &los.LoanApplication{
			AppID:   appId,
			Email:   in.Email,
			PhoneNo: in.PhoneNo,
		},
	}, nil
}

func (s *LosApiServer) NotificationDE1(ctx context.Context, in *los.NotificationDE1Request) (*los.NotificationDE1Response, error) {

	appId := in.AppID

	return &los.NotificationDE1Response{
		AppID:  appId,
		Status: in.Status,
	}, nil
}

func (s *LosApiServer) QueryState(ctx context.Context, in *los.QueryStateRequest) (*los.QueryStateResponse, error) {

	appId := in.AppID

	return &los.QueryStateResponse{
		LoanAppState: &los.LoanAppState{
			AppID:      appId,
			RunID:      "run_id",
			WorkflowID: "workflow_id",
			State:      "current_state",
		},
	}, nil
}

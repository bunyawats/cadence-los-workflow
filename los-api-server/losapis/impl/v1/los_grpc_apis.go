package v1

import (
	"cadence-los-workflow/common"
	los "cadence-los-workflow/los-api-server/losapis/gen/v1"
	"context"
	cadence_client "go.uber.org/cadence/client"
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

	if err := s.M.CreateNewLoanApplication(appID); err != nil {
		return nil, err
	}

	common.StartWorkflow(s.H)

	return &los.CreateNewAppResponse{
		AppID: appID,
	}, nil
}

func (s *LosApiServer) SubmitFormOne(_ context.Context, in *los.SubmitFormOneRequest) (*los.SubmitFormOneResponse, error) {

	request := common.LoanApplication{
		AppID: in.AppID,
		Fname: in.FName,
		Lname: in.LName,
	}

	a, err := s.M.SaveFormOne(&request)
	if err != nil {
		return nil, err
	}

	common.CompleteActivity(s.M, s.W, in.AppID, "SUCCESS")

	return &los.SubmitFormOneResponse{
		LoanApp: &los.LoanApplication{
			AppID: a.AppID,
			FName: a.Fname,
			LName: a.Lname,
		},
	}, nil
}

func (s *LosApiServer) SubmitFormTwo(_ context.Context, in *los.SubmitFormTwoRequest) (*los.SubmitFormTwoResponse, error) {

	request := common.LoanApplication{
		AppID:   in.AppID,
		Email:   in.Email,
		PhoneNo: in.PhoneNo,
	}

	a, err := s.M.SaveFormTwo(&request)
	if err != nil {
		return nil, err
	}

	common.CompleteActivity(s.M, s.W, in.AppID, "SUCCESS")

	return &los.SubmitFormTwoResponse{
		LoanApp: &los.LoanApplication{
			AppID:   a.AppID,
			FName:   a.Fname,
			LName:   a.Lname,
			Email:   a.Email,
			PhoneNo: a.PhoneNo,
		},
	}, nil
}

func (s *LosApiServer) NotificationDE1(_ context.Context, in *los.NotificationDE1Request) (*los.NotificationDE1Response, error) {

	s.R.Publish2RabbitMQ(&common.DEResult{
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

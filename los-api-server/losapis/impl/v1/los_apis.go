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

func (s *LosApiServer) Onboard(ctx context.Context, req *los.OnboardRequest) (*los.OnboardResponse, error) {

	appId := req.AppID

	return &los.OnboardResponse{
		AppID: appId,
	}, nil
}

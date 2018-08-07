package healthcheck

import (
	"context"
	"time"

	healthcheckPb "github.com/linkerd/linkerd2/controller/gen/common/healthcheck"
	"google.golang.org/grpc"
)

type grpcStatusChecker interface {
	SelfCheck(ctx context.Context, in *healthcheckPb.SelfCheckRequest, opts ...grpc.CallOption) (*healthcheckPb.SelfCheckResponse, error)
}

type statusCheckerProxy struct {
	delegate grpcStatusChecker
	prefix   string
}

func (proxy *statusCheckerProxy) SelfCheck() []*healthcheckPb.CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	selfCheckResponse, err := proxy.delegate.SelfCheck(ctx, &healthcheckPb.SelfCheckRequest{})
	if err != nil {
		return []*healthcheckPb.CheckResult{
			&healthcheckPb.CheckResult{
				SubsystemName:         proxy.prefix,
				CheckDescription:      "can query the Linkerd API",
				Status:                healthcheckPb.CheckStatus_ERROR,
				FriendlyMessageToUser: err.Error(),
			},
		}
	}

	return selfCheckResponse.Results
}

func NewGrpcStatusChecker(name string, grpClient grpcStatusChecker) StatusChecker {
	return &statusCheckerProxy{
		prefix:   name,
		delegate: grpClient,
	}
}

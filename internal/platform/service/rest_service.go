package service

import (
	"context"

	api "github.com/dohernandez/qonto/pkg/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// QontoRESTService Wrapper on top of the GRPC server to be able to use the interceptor for
// REST request as it is used for grpc request.
type QontoRESTService struct {
	*QontoService

	unaryInt grpc.UnaryServerInterceptor
}

// NewQontoRESTService creates an instance of QontoService.
func NewQontoRESTService(service *QontoService) *QontoRESTService {
	return &QontoRESTService{
		QontoService: service,
	}
}

// TransferBulk is wrapper on the unary RPC to performs given transfers for REST calls.
func (s *QontoRESTService) TransferBulk(ctx context.Context, req *api.TransferBulkRequest) (*emptypb.Empty, error) {
	info := &grpc.UnaryServerInfo{
		Server:     s.QontoService,
		FullMethod: "/api.qonto/TransferBulk",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.QontoService.TransferBulk(ctx, req.(*api.TransferBulkRequest))
	}

	resp, err := s.unaryInt(ctx, req, info, handler)
	if err != nil {
		s := status.Convert(err)

		if s.Code() == codes.FailedPrecondition {
			err = &runtime.HTTPStatusError{
				HTTPStatus: 422,
				Err:        err,
			}
		}
	}

	return resp.(*emptypb.Empty), err
}

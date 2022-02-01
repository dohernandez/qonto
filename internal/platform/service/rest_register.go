package service

import (
	"context"

	"github.com/dohernandez/qonto/pkg/grpc/rest"
	api "github.com/dohernandez/qonto/pkg/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// RegisterHandlerService registers the service implementation to mux.
func (s *QontoRESTService) RegisterHandlerService(mux *runtime.ServeMux) error {
	return api.RegisterQontoServiceHandlerServer(context.Background(), mux, s)
}

// WithUnaryServerInterceptor set the UnaryServerInterceptor for the REST service.
func (s *QontoRESTService) WithUnaryServerInterceptor(i grpc.UnaryServerInterceptor) rest.ServiceServer {
	s.unaryInt = i

	return s
}

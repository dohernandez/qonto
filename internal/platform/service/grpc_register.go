package service

import (
	api "github.com/dohernandez/qonto/pkg/proto"
	"google.golang.org/grpc"
)

// RegisterService registers the service implementation to grpc service.
func (s *QontoService) RegisterService(sr grpc.ServiceRegistrar) {
	api.RegisterQontoServiceServer(sr, s)
}

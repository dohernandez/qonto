package service

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// kitTemplateRESTService Wrapper on top of the GRPC server to be able to use the interceptor for
// REST request as it is used for grpc request.
type kitTemplateRESTService struct {
	*KitTemplateService

	unaryInt grpc.UnaryServerInterceptor
}

// PostFuncName is wrapper on the unary RPC to ... for REST calls.
func (s *kitTemplateRESTService) PostFuncName(ctx context.Context, req interface{}) (interface{}, error) {
	info := &grpc.UnaryServerInfo{
		Server:     s.KitTemplateService,
		FullMethod: "/kit.template.Service/PostFuncName",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.KitTemplateService.PostFuncName(ctx, req)
	}

	resp, err := s.unaryInt(ctx, req, info, handler)

	return resp, err
}

// KitTemplateRESTServiceRegister registers the service  into the REST service.
type KitTemplateRESTServiceRegister struct {
	kitTemplateRESTService *kitTemplateRESTService
}

// NewRESTServiceRegister create an instance used to register the REST point service.
func NewRESTServiceRegister(service *KitTemplateService) *KitTemplateRESTServiceRegister {
	return &KitTemplateRESTServiceRegister{
		kitTemplateRESTService: &kitTemplateRESTService{
			KitTemplateService: service,
		},
	}
}

// RegisterHandlerService registers the service implementation to mux.
func (rsr *KitTemplateRESTServiceRegister) RegisterHandlerService(mux *runtime.ServeMux) error {
	// register
	return nil
}

// WithUnaryServerInterceptor set the UnaryServerInterceptor for the REST service.
func (rsr *KitTemplateRESTServiceRegister) WithUnaryServerInterceptor(i grpc.UnaryServerInterceptor) *KitTemplateRESTServiceRegister {
	rsr.kitTemplateRESTService.unaryInt = i

	return rsr
}

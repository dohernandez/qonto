package service

import (
	"google.golang.org/grpc"
)

// KitTemplateGRPCServiceRegister wrap the service ... .
type KitTemplateGRPCServiceRegister struct {
	service *KitTemplateService
}

// NewGRPCServiceRegister create an instance used to register the grpc service.
func NewGRPCServiceRegister(service *KitTemplateService) *KitTemplateGRPCServiceRegister {
	return &KitTemplateGRPCServiceRegister{
		service: service,
	}
}

// RegisterService registers the service implementation to grpc service.
func (sr *KitTemplateGRPCServiceRegister) RegisterService(s grpc.ServiceRegistrar) {
}

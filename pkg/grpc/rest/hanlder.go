package rest

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// ServiceServer is an interface for a server that provides services.
type ServiceServer interface {
	RegisterHandlerService(mux *runtime.ServeMux) error
}

// ServiceHandlerServerFunc is the function to register the http handlers for service to "mux".
type ServiceHandlerServerFunc func(mux *runtime.ServeMux) error

// HandlerPathFunc allows users to configure custom path handlers for mux service.
type HandlerPathFunc func(mux *runtime.ServeMux) error

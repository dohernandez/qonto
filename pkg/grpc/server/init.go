package server

import (
	"context"
	"net"

	grpcZapLogger "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type InitGRPCServiceConfig struct {
	Listener       net.Listener
	Service        ServiceServer
	Logger         *zap.Logger
	UInterceptor   []grpc.UnaryServerInterceptor
	WithReflective bool
	Options        []Option
}

// InitGRPCService initialize an instance of grpc service, with all the instrumentation.
func InitGRPCService(
	_ context.Context,
	cfg InitGRPCServiceConfig,
) (*Server, error) {
	grpcZapLogger.ReplaceGrpcLoggerV2(cfg.Logger)

	opts := append(cfg.Options,
		WithListener(cfg.Listener, true),
		// registering point service using the point service registerer
		WithService(cfg.Service),
		ChainUnaryInterceptor(cfg.UInterceptor...),
	)

	// Enabling reflection in dev and testing env.
	if cfg.WithReflective {
		opts = append(opts, WithReflective())
	}

	return NewServer(opts...), nil
}

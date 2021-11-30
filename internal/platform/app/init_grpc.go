package app

import (
	"context"
	"fmt"
	"net"

	"github.com/dohernandez/kit-template/internal/platform/config"
	"github.com/dohernandez/kit-template/internal/platform/service"
	grpcMetrics "github.com/dohernandez/kit-template/pkg/grpc/metrics"
	grpcServer "github.com/dohernandez/kit-template/pkg/grpc/server"
	grpcZapLogger "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcCtxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcOpentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
)

// NewGRPCService creates an instance of grpc service, with all the instrumentation.
func NewGRPCService(
	_ context.Context,
	cfg *config.Config,
	locator *Locator,
	srv *service.KitTemplateService,
	interceptors []grpc.UnaryServerInterceptor,
	metricsServer *grpcMetrics.Server,
) (*grpcServer.Server, error) {
	grpcServiceRegister := service.NewGRPCServiceRegister(srv)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	if err != nil {
		return nil, err
	}

	grpcZapLogger.ReplaceGrpcLoggerV2(locator.ZapLogger())

	opts := []grpcServer.Option{
		grpcServer.WithListener(grpcListener, true),
		// registering point service using the point service registerer
		grpcServer.WithService(grpcServiceRegister),
		grpcServer.WithMetrics(metricsServer.ServerMetrics()),
		grpcServer.ChainUnaryInterceptor(interceptors...),
	}

	// Enabling reflection in dev env.
	if cfg.IsDev() {
		opts = append(opts, grpcServer.WithReflective())
	}

	return grpcServer.NewServer(opts...), nil
}

// InitGRPCUnitaryInterceptors initialize unitary interceptors used by the grpc service.
func InitGRPCUnitaryInterceptors(
	l *Locator,
	metricsServer *grpcMetrics.Server,
) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		// recovering from panic
		grpcRecovery.UnaryServerInterceptor(),
		// adding tracing
		grpcOpentracing.UnaryServerInterceptor(),
		// adding metrics
		metricsServer.ServerMetrics().UnaryServerInterceptor(),
		// adding logger
		grpcCtxtags.UnaryServerInterceptor(grpcCtxtags.WithFieldExtractor(grpcCtxtags.CodeGenRequestFieldExtractor)),
		grpcZapLogger.UnaryServerInterceptor(l.ZapLogger()),
	}
}

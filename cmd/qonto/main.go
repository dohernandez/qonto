package main

import (
	"context"
	"fmt"
	"net"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/qonto/internal/platform/app"
	"github.com/dohernandez/qonto/internal/platform/config"
	grpcMetrics "github.com/dohernandez/qonto/pkg/grpc/metrics"
	grpcRest "github.com/dohernandez/qonto/pkg/grpc/rest"
	grpcServer "github.com/dohernandez/qonto/pkg/grpc/server"
	"github.com/dohernandez/qonto/pkg/must"
	"github.com/dohernandez/qonto/pkg/servicing"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// load configurations
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	metricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppMetricsPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init Metrics service listener"))

	srvMetrics, err := grpcMetrics.NewMetricsService(ctx, metricsListener)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init Metrics service"))

	// initialize locator
	deps, err := app.NewServiceLocator(cfg, func(l *app.Locator) {
		l.GRPCUnitaryInterceptors = append(l.GRPCUnitaryInterceptors,
			// adding metrics
			srvMetrics.ServerMetrics().UnaryServerInterceptor(),
		)
	})
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init locator"))

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service listener"))

	srvGRPC, err := grpcServer.InitGRPCService(
		ctx,
		grpcServer.InitGRPCServiceConfig{
			Listener:       grpcListener,
			Service:        deps.QontoService,
			Logger:         deps.ZapLogger(),
			UInterceptor:   deps.GRPCUnitaryInterceptors,
			WithReflective: cfg.IsDev(),
			Options: []grpcServer.Option{
				grpcServer.WithMetrics(srvMetrics.ServerMetrics()),
			},
		},
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service"))

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service listener"))

	srvREST, err := grpcRest.InitRESTService(
		ctx,
		grpcRest.InitRESTServiceConfig{
			Listener:         restTListener,
			Service:          deps.QontoRESTService,
			UInterceptor:     deps.GRPCUnitaryInterceptors,
			Handlers:         deps.Handlers,
			ResponseModifier: deps.ResponseModifier,
		},
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service"))

	services := servicing.WithGracefulSutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	err = services.Start(
		ctx,
		func(ctx context.Context, msg string) {
			deps.CtxdLogger().Important(ctx, msg)
		},
		srvMetrics,
		srvGRPC,
		srvREST,
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
}

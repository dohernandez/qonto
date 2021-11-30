package main

import (
	"context"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/internal/platform/app"
	"github.com/dohernandez/kit-template/internal/platform/config"
	"github.com/dohernandez/kit-template/internal/platform/service"
	"github.com/dohernandez/kit-template/pkg/must"
	"github.com/dohernandez/kit-template/pkg/servicing"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// load configurations
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	// initialize locator
	l, err := app.NewServiceLocator(cfg)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init locator"))

	srvMetrics, err := app.NewMetricsService(ctx, cfg)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init Metrics service"))

	srv := service.NewKitTemplateService()

	// enabling interceptor for grpc and rest
	interceptors := app.InitGRPCUnitaryInterceptors(l, srvMetrics)

	srvGRPC, err := app.NewGRPCService(ctx, cfg, l, srv, interceptors, srvMetrics)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service"))

	srvREST, err := app.NewRESTService(ctx, cfg, l, srv, interceptors)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service"))

	services := servicing.WithGracefulSutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, l)
		},
	)

	err = services.Start(
		ctx,
		func(ctx context.Context, msg string) {
			l.CtxdLogger().Important(ctx, msg)
		},
		srvMetrics,
		srvGRPC,
		srvREST,
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
}

package kit_template_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"testing"

	"github.com/bool64/ctxd"
	"github.com/bool64/dbdog"
	"github.com/bool64/httpdog"
	"github.com/bool64/sqluct"
	"github.com/cucumber/godog"
	"github.com/dohernandez/qonto/internal/domain/model"
	"github.com/dohernandez/qonto/internal/platform/app"
	"github.com/dohernandez/qonto/internal/platform/config"
	grpcRest "github.com/dohernandez/qonto/pkg/grpc/rest"
	grpcServer "github.com/dohernandez/qonto/pkg/grpc/server"
	"github.com/dohernandez/qonto/pkg/must"
	"github.com/dohernandez/qonto/pkg/servicing"
	"github.com/dohernandez/qonto/pkg/test/feature"
	dbdogcleaner "github.com/dohernandez/qonto/pkg/test/feature/database"
	"github.com/nhatthm/clockdog"
)

func TestIntegration(t *testing.T) {
	ctx := context.Background()

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// load configurations
	err := config.WithEnvFiles(".env.integration-test")
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load env from .env.integration-test"))
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	cfg.Environment = "test"
	cfg.Log.Output = ioutil.Discard

	clock := clockdog.New()

	deps, err := app.NewServiceLocator(cfg, func(l *app.Locator) {
		l.ClockProvider = clock
	})
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init service locator"))

	dbm := initDBManager(deps.Storage)
	dbmCleaner := initDBMCleaner(dbm)

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
				grpcServer.WithAddrAssigned(),
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
			Options: []grpcRest.Option{
				grpcRest.WithAddrAssigned(),
			},
		},
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service"))

	services := servicing.WithGracefulSutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	go func() {
		err = services.Start(ctx,
			func(ctx context.Context, msg string) {
				deps.CtxdLogger().Important(ctx, msg)
			},
			srvGRPC,
			srvREST,
		)
		must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
	}()

	baseRESTURL := <-srvREST.AddrAssigned
	local := httpdog.NewLocal(baseRESTURL)

	feature.RunFeatures(t, "features", func(_ *testing.T, s *godog.ScenarioContext) {
		local.RegisterSteps(s)

		dbm.RegisterSteps(s)
		dbmCleaner.RegisterSteps(s)

		clock.RegisterContext(s)
	})

	must.NotFail(services.Close())
}

func initDBManager(storage *sqluct.Storage) *dbdog.Manager {
	tableMapper := dbdog.NewTableMapper()

	dbm := dbdog.Manager{
		TableMapper: tableMapper,
	}

	dbm.Instances = map[string]dbdog.Instance{
		"postgres": {
			Storage: storage,
			Tables: map[string]interface{}{
				"transactions":  new(model.Transaction),
				"bank_accounts": new(model.BankAccount),
			},
			PostCleanup: map[string][]string{
				"transactions":  {"ALTER SEQUENCE transactions_id_seq RESTART"},
				"bank_accounts": {"ALTER SEQUENCE bank_accounts_id_seq RESTART"},
			},
		},
	}

	return &dbm
}

func initDBMCleaner(dbm *dbdog.Manager) *dbdogcleaner.ManagerCleaner {
	return &dbdogcleaner.ManagerCleaner{
		Manager: dbm,
	}
}

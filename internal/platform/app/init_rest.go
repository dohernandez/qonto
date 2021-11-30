package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/internal/platform/config"
	"github.com/dohernandez/kit-template/internal/platform/service"
	grpcRest "github.com/dohernandez/kit-template/pkg/grpc/rest"
	"github.com/dohernandez/kit-template/pkg/must"
	"github.com/dohernandez/kit-template/resources/swagger"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	mux "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v3 "github.com/swaggest/swgui/v3"
	"google.golang.org/grpc"
)

// NewRESTService creates an instance of REST service based on the GRPC service.
func NewRESTService(
	ctx context.Context,
	cfg *config.Config,
	locator *Locator,
	srv *service.KitTemplateService,
	interceptors []grpc.UnaryServerInterceptor,
) (*grpcRest.Server, error) {
	restServiceRegister := service.NewRESTServiceRegister(srv)

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	if err != nil {
		return nil, err
	}

	restServiceRegister.WithUnaryServerInterceptor(
		grpcMiddleware.ChainUnaryServer(interceptors...),
	)

	opts := []grpcRest.Option{
		grpcRest.WithListener(restTListener, true),
		// use to registering point service using the point service registerer
		grpcRest.WithService(restServiceRegister),

		// handler root path
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				w.Header().Set("content-type", "text/html")

				_, err := w.Write([]byte("Welcome to " + locator.Config.ServiceName +
					`. Please read API <a href="docs">documentation</a>.`))
				if err != nil {
					locator.CtxdLogger().Error(r.Context(), "failed to write response",
						"error", err)
				}
			})
		}),
	}

	swaggerOptions := swaggerHandlersOptions(ctx, locator)
	opts = append(opts, swaggerOptions...)

	return grpcRest.NewServer(opts...)
}

func swaggerHandlersOptions(
	ctx context.Context,
	locator *Locator,
) []grpcRest.Option {
	swh := v3.NewHandler(locator.Config.ServiceName, "/docs/service.swagger.json", "/docs/")

	return []grpcRest.Option{
		// handler docs paths
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/service.swagger.json", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				w.Header().Set("Content-Type", "application/json")

				_, err := w.Write(swagger.SwgJSON)
				must.NotFail(ctxd.WrapError(ctx, err, "failed to load /docs/service.swagger.json file"))
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/swagger-ui-bundle.js", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/swagger-ui-standalone-preset.js", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
		grpcRest.WithHandlerPathOption(func(mux *mux.ServeMux) error {
			return mux.HandlePath(http.MethodGet, "/docs/swagger-ui.css", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			})
		}),
	}
}

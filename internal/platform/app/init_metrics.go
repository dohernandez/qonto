package app

import (
	"context"
	"fmt"
	"net"

	"github.com/dohernandez/kit-template/internal/platform/config"
	grpcMetrics "github.com/dohernandez/kit-template/pkg/grpc/metrics"
)

// NewMetricsService creates an instance of metrics service.
func NewMetricsService(
	_ context.Context,
	cfg *config.Config,
) (*grpcMetrics.Server, error) {
	metricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppMetricsPort))
	if err != nil {
		return nil, err
	}

	opts := []grpcMetrics.Option{
		grpcMetrics.WithListener(metricsListener, true),
	}

	return grpcMetrics.NewServer(opts...), nil
}

package infra

import (
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/apiserver_std"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
	"log/slog"
	"time"
)

func Start(deployment string, overrides func(cfg *app.Config), loaded chan struct{}) {
	zapLogger, _ := zap.NewDevelopment()
	logger := slog.New(slogzap.Option{
		AddSource: true,
		Logger:    zapLogger,
	}.NewZapHandler())
	ctx := katapp.StartContext(logger, deployment)

	cfg := app.LoadConfig(deployment)
	if overrides != nil {
		overrides(cfg)
	}
	di := WireDependencies(ctx, cfg)
	defer di.Close()

	uc := usecase.NewUseCases(cfg, di.Ports)

	// Pick which server implementation you want to use
	// Uncomment only one of the following lines:
	// -- Option 1: Echo framework
	//server := apiserver_echo.Start(ctx, uc)
	// -- Option 2: Standard net/http
	server := apiserver_std.Start(ctx, uc)
	// -- Option 3: Chi router
	//server := apiserver_chi.Start(ctx, uc)

	// needed for integration tests only
	if loaded != nil {
		loaded <- struct{}{}
	}

	// Make sure to use the corresponding WaitForInterruptSignal function for your chosen server implementation
	// WaitForInterruptSignal gracefully shuts down the server when interrupt signal (such as Ctrl+C or
	// SIGTERM coming from your instance) is received

	// -- Option 1: Echo framework
	//apiserver_echo.WaitForInterruptSignal(ctx, server, 3*time.Second)
	// -- Option 2: Standard net/http
	apiserver_std.WaitForInterruptSignal(ctx, server, 3*time.Second)
	// -- Option 3: Chi router
	//apiserver_chi.WaitForInterruptSignal(ctx, server, 3*time.Second)
}

package infra

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
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
	validateMandatoryConfig(cfg)

	if overrides != nil {
		overrides(cfg)
	}
	di := WireDependencies(ctx, cfg)
	defer di.Close()

	uc := usecase.NewUseCases(cfg, di.Ports)

	server := apiserver.Start(ctx, uc)

	// needed for integration tests only
	if loaded != nil {
		loaded <- struct{}{}
	}

	// Make sure to use the corresponding WaitForInterruptSignal function for your chosen server implementation
	// WaitForInterruptSignal gracefully shuts down the server when interrupt signal (such as Ctrl+C or
	// SIGTERM coming from your instance) is received

	// -- Option 1: Echo framework
	apiserver.WaitForInterruptSignal(ctx, server, 3*time.Second)
}

func validateMandatoryConfig(cfg *app.Config) {
	if cfg.Credentials.Secret == "_" || cfg.Credentials.Secret == "" {
		panic("credentials.secret is not set")
	}
}

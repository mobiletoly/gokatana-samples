package infra

import (
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/apiserver"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
	"log/slog"
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
	apiserver.Start(ctx, uc, loaded)
}

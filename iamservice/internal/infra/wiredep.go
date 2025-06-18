package infra

import (
	"context"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/persist"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/katpg"
)

type Dependencies struct {
	Close func()
	Ports *outport.Ports
}

func WireDependencies(ctx context.Context, cfg *app.Config) *Dependencies {
	katapp.Logger(ctx).Info("Initialize DI objects")

	db := katpg.MustConnect(ctx, &cfg.Database)
	if cfg.Deployment != "test" {
		db.MustDoMigration(ctx)
	}

	return &Dependencies{
		Close: func() {
			katapp.Logger(ctx).Info("performing cleanup of all dependency objects")
			db.Close()
		},
		Ports: &outport.Ports{
			AuthPersist: persist.NewAuthUserAdapter(db),
		},
	}
}

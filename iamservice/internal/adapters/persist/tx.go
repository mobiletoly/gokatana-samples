package persist

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana/katpg"
)

type TxAdapter struct {
	db *katpg.DBLink
}

// NewTxAdapter creates a new DbAdapter
func NewTxAdapter(db *katpg.DBLink) outport.TxPort {
	return &TxAdapter{
		db: db,
	}
}

func (d TxAdapter) Run(ctx context.Context, f func(tx pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, d.db.Pool, func(tx pgx.Tx) error {
		return f(tx)
	})
}

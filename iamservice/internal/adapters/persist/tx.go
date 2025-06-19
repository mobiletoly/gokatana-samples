package persist

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana/katpg"
)

type TransactionAdapter struct {
	db *katpg.DBLink
}

// NewTransactionAdapter creates a new DbAdapter
func NewTransactionAdapter(db *katpg.DBLink) outport.Transaction {
	return &TransactionAdapter{
		db: db,
	}
}

func (d TransactionAdapter) Run(ctx context.Context, f func() error) error {
	return pgx.BeginFunc(ctx, d.db, func(tx pgx.Tx) error {
		return f()
	})
}

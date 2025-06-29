package outport

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type TxPort interface {
	Run(ctx context.Context, f func(tx pgx.Tx) error) error
}

// TxWithResult wraps a database transaction that returns a result
func TxWithResult[T any](ctx context.Context, tx TxPort, f func(tx pgx.Tx) (T, error)) (T, error) {
	var result T
	err := tx.Run(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = f(tx)
		return err
	})
	return result, err
}

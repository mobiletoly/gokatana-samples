package outport

import "context"

type Transaction interface {
	Run(ctx context.Context, f func() error) error
}

func TxWithResult[T any](ctx context.Context, tx Transaction, f func() (T, error)) (T, error) {
	var result T
	err := tx.Run(ctx, func() error {
		var err error
		result, err = f()
		return err
	})
	return result, err
}

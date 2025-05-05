package outport

import (
	"context"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/model"
)

type Contact interface {
	LoadByID(ctx context.Context, ID string) (*model.Contact, error)
	Add(ctx context.Context, addContact *model.AddContact) (*model.Contact, error)
	LoadAll(ctx context.Context) ([]*model.Contact, error)
}

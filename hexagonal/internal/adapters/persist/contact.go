package persist

import (
	"context"
	"fmt"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/persist/internal/mapper"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/persist/internal/repo"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/outport"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/katpg"
)

type contactAdapter struct {
	repo *repo.ContactRepo
}

func (a *contactAdapter) LoadAll(ctx context.Context) ([]*model.Contact, error) {
	contacts, err := a.repo.SelectAllContacts(ctx)
	if err != nil {
		katapp.Logger(ctx).Error("failed to select all contacts", "error", err)
		return nil, katpg.ToAppError(err, "database select failed")
	}
	var ms = make([]*model.Contact, len(contacts))
	for i, c := range contacts {
		ms[i] = mapper.ContactEntityToModel(&c)
	}
	return ms, nil
}

func (a *contactAdapter) Add(ctx context.Context, add *model.AddContact) (*model.Contact, error) {
	addContactModel := mapper.AddContactModelToEntity(add)
	nID, err := a.repo.InsertContact(ctx, addContactModel)
	if err != nil {
		katapp.Logger(ctx).Error("failed to insert contact", "error", err)
		appErr := katpg.ToAppError(err, "database insert failed")
		if appErr.Scope == katapp.ErrDuplicate {
			return nil, katapp.NewErr(katapp.ErrDuplicate, "contact with the same first and last name already exists")
		} else {
			return nil, appErr
		}
	}
	existingContact, err := a.repo.SelectContactByID(ctx, nID)
	if err != nil || existingContact == nil {
		katapp.Logger(ctx).Error("failed to select existing contact", "ID", nID, "error", err)
		return nil, katpg.ToAppError(err, fmt.Sprintf("failed to select existing contact (id=%d)", nID))
	}
	m := mapper.ContactEntityToModel(existingContact)
	return m, nil
}

func (a *contactAdapter) LoadByID(ctx context.Context, ID string) (*model.Contact, error) {
	nID, err := mapper.ModelIdToRepoId(ID)
	if err != nil {
		katapp.Logger(ctx).Error("invalid record id", "error", err)
		return nil, err
	}
	c, err := a.repo.SelectContactByID(ctx, nID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to select contact", "ID", nID, "error", err)
		return nil, katpg.ToAppError(err, fmt.Sprintf("failed to select contact (id=%d)", nID))
	}
	if c == nil {
		katapp.Logger(ctx).Error("contact not found", "ID", nID)
		return nil, katapp.NewErr(katapp.ErrNotFound, fmt.Sprintf("contact not found (id=%d)", nID))
	}
	m := mapper.ContactEntityToModel(c)
	return m, nil
}

func NewContactAdapter(db *katpg.DBLink) outport.Contact {
	return &contactAdapter{
		repo: repo.NewContactRepo(db.Pool),
	}
}

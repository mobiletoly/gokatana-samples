package usecase

import (
	"context"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/outport"
	"github.com/mobiletoly/gokatana/katapp"
)

type Contact struct {
	config      *app.Config
	contactPort outport.Contact
}

func (uc *Contact) LoadContactByID(ctx context.Context, ID string) (*model.Contact, error) {
	ctx = katapp.ContextWithAddedGroup(ctx, "usecase.Contact.LoadContactByID")
	katapp.Logger(ctx).InfoContext(ctx, "Load contact", "ID", ID)
	contact, err := uc.contactPort.LoadByID(ctx, ID)
	if err != nil {
		katapp.Logger(ctx).WarnContext(ctx, "failed to load contact", "ID", ID, "error", err)
		return nil, err
	}
	katapp.Logger(ctx).DebugContext(ctx, "successfully loaded contact", "entity", contact)
	return contact, nil
}

func (uc *Contact) AddContact(ctx context.Context, add *model.AddContact) (*model.Contact, error) {
	katapp.Logger(ctx).DebugContext(ctx, "Add contact", "entity", add)
	if add.FirstName == nil || add.LastName == nil || *add.FirstName == "" || *add.LastName == "" {
		katapp.Logger(ctx).WarnContext(ctx, "first name and last name are required")
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "first name and last name are required")
	}
	contact, err := uc.contactPort.Add(ctx, add)
	if err != nil {
		katapp.Logger(ctx).WarnContext(ctx, "failed to add contact", "error", err)
		return nil, err
	}
	katapp.Logger(ctx).DebugContext(ctx, "successfully added contact", "entity", contact)
	return contact, nil
}

func (uc *Contact) LoadAllContacts(ctx context.Context) ([]*model.Contact, error) {
	katapp.Logger(ctx).DebugContext(ctx, "Load all contacts")
	contacts, err := uc.contactPort.LoadAll(ctx)
	if err != nil {
		katapp.Logger(ctx).WarnContext(ctx, "failed to load all contacts", "error", err)
		return nil, err
	}
	katapp.Logger(ctx).DebugContext(ctx, "successfully loaded all contacts", "entities", contacts)
	return contacts, nil
}

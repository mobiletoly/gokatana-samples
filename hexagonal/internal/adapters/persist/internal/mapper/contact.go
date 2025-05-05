package mapper

import (
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/persist/internal/repo"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/model"
)

func ContactEntityToModel(c *repo.ContactEntity) *model.Contact {
	return model.NewContactBuilder().
		ID(RepoIdToModelId(*c.ID)).
		FirstName(&c.FirstName).
		LastName(&c.LastName).
		Build()
}

func AddContactModelToEntity(c *model.AddContact) *repo.ContactEntity {
	return repo.NewContactEntityBuilder().
		ID(nil).
		FirstName(*c.FirstName).
		LastName(*c.LastName).
		Build()
}

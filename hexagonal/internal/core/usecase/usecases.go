package usecase

import (
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/outport"
)

type UseCases struct {
	Config  *app.Config
	Contact *Contact
}

func NewUseCases(cfg *app.Config, ports *outport.Ports) *UseCases {
	return &UseCases{
		Config: cfg,
		Contact: &Contact{
			config:      cfg,
			contactPort: ports.Contact,
		},
	}
}

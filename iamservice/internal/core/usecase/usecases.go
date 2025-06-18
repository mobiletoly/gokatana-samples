package usecase

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
)

type UseCases struct {
	Config  *app.Config
	Auth    *AuthUser
	UserMgm *UserMgm
}

func NewUseCases(cfg *app.Config, ports *outport.Ports) *UseCases {
	return &UseCases{
		Config:  cfg,
		Auth:    NewAuthUser(ports.AuthPersist, cfg.Credentials.Secret),
		UserMgm: NewUserMgm(ports.AuthPersist),
	}
}

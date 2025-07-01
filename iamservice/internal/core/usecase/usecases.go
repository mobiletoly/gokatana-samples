package usecase

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
)

type UseCases struct {
	Config         *app.Config
	Auth           *AuthMgm
	UserMgm        *UserMgm
	UserProfileMgm *UserProfileMgm
}

func NewUseCases(cfg *app.Config, ports *outport.Ports) *UseCases {
	return &UseCases{
		Config: cfg,
		Auth: NewAuthUser(
			&cfg.Server, ports.AuthUserPersist, ports.Tx, ports.Mailer, cfg.Credentials.JwtSecret,
		),
		UserMgm:        NewUserMgm(ports.AuthUserPersist, ports.Tx),
		UserProfileMgm: NewUserProfileMgm(ports),
	}
}

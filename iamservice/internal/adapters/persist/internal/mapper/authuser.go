package mapper

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"time"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/persist/internal/repo"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
)

// SwaggerSignupRequestToAuthUserEntity converts swagger.SignupRequest to repo.AuthUserEntity
func SwaggerSignupRequestToAuthUserEntity(req *swagger.SignupRequest, userID string, hashedPassword string) *repo.AuthUserEntity {
	now := time.Now()

	return repo.NewAuthUserEntityBuilder().
		ID(&userID).
		Email(string(*req.Email)).
		PasswordHash(hashedPassword).
		FirstName(*req.FirstName).
		LastName(*req.LastName).
		IsActive(true).
		EmailVerified(false).
		CreatedAt(&now).
		UpdatedAt(&now).
		Build()
}

func AuthUserEntityToAuthUserModel(entity *repo.AuthUserEntity) *model.AuthUser {
	return model.NewAuthUserBuilder().
		ID(*entity.ID).
		Email(entity.Email).
		PasswordHash(entity.PasswordHash).
		FirstName(entity.FirstName).
		LastName(entity.LastName).
		IsActive(entity.IsActive).
		EmailVerified(entity.EmailVerified).
		CreatedAt(entity.CreatedAt.String()).
		UpdatedAt(entity.UpdatedAt.String()).
		Build()
}

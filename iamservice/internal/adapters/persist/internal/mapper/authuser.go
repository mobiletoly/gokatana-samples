package mapper

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/oapi-codegen/runtime/types"
	"github.com/samber/lo"
	"time"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/persist/internal/repo"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
)

// SwaggerSignupRequestToAuthUserEntity converts swagger.SignupRequest to repo.AuthUserEntity
func SwaggerSignupRequestToAuthUserEntity(req *swagger.SignupRequest, userID string, hashedPassword string, tenantID string) *repo.AuthUserEntity {
	now := time.Now()

	return repo.NewAuthUserEntityBuilder().
		ID(&userID).
		Email(string(req.Email)).
		PasswordHash(hashedPassword).
		FirstName(req.FirstName).
		LastName(req.LastName).
		TenantID(tenantID).
		IsActive(true).
		EmailVerified(false).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
}

func AuthUserEntityToAuthUserModel(entity *repo.AuthUserEntity) *model.AuthUser {
	return model.NewAuthUserBuilder().
		ID(*entity.ID).
		Email(entity.Email).
		PasswordHash(entity.PasswordHash).
		FirstName(entity.FirstName).
		LastName(entity.LastName).
		TenantID(entity.TenantID).
		IsActive(entity.IsActive).
		EmailVerified(entity.EmailVerified).
		CreatedAt(entity.CreatedAt).
		UpdatedAt(entity.UpdatedAt).
		Build()
}

// TenantEntityToTenantModel converts repo.TenantEntity to model.Tenant
func TenantEntityToTenantModel(entity *repo.TenantEntity) *model.Tenant {
	return model.NewTenantBuilder().
		ID(entity.ID).
		Name(entity.Name).
		Description(entity.Description).
		CreatedAt(entity.CreatedAt).
		UpdatedAt(entity.UpdatedAt).
		Build()
}

// TenantModelToTenantResponse converts model.Tenant to swagger.TenantResponse
func TenantModelToTenantResponse(tenant *model.Tenant) *swagger.TenantResponse {
	return swagger.NewTenantResponseBuilder().
		CreatedAt(tenant.CreatedAt).
		Description(tenant.Description).
		Id(tenant.ID).
		Name(tenant.Name).
		UpdatedAt(tenant.UpdatedAt).
		Build()
}

// TenantCreateRequestToTenantEntity converts swagger.TenantCreateRequest to repo.TenantEntity
func TenantCreateRequestToTenantEntity(req *swagger.CreateTenantRequest) *repo.TenantEntity {
	now := time.Now()
	return repo.NewTenantEntityBuilder().
		ID(req.Id).
		Name(req.Name).
		Description(req.Description).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
}

// TenantUpdateRequestToTenantEntity converts swagger.TenantUpdateRequest to repo.TenantEntity
func TenantUpdateRequestToTenantEntity(tenantID string, req *swagger.UpdateTenantRequest) *repo.TenantEntity {
	return &repo.TenantEntity{
		ID:          tenantID,
		Name:        req.Name,
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}
}

// UserProfileEntityToSwagger converts repo.UserProfileEntity to model.UserProfile
func UserProfileEntityToSwagger(entity *repo.UserProfileEntity) *swagger.UserProfileResponse {
	var birthDate *types.Date
	if entity.BirthDate != nil {
		birthDate = &types.Date{Time: *entity.BirthDate}
	}

	var gender *swagger.UserProfileGender
	if entity.Gender != nil {
		gender = lo.ToPtr(swagger.UserProfileGender(*entity.Gender))
	}

	return swagger.NewUserProfileResponseBuilder().
		BirthDate(birthDate).
		CreatedAt(entity.CreatedAt).
		Gender(gender).
		Height(entity.Height).
		Id(*entity.ID).
		IsMetric(entity.IsMetric).
		UpdatedAt(entity.UpdatedAt).
		UserId(entity.UserID).
		Weight(entity.Weight).
		Build()
}

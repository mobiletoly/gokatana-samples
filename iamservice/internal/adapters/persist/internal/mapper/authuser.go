package mapper

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"time"

	"github.com/go-openapi/strfmt"
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
		CreatedAt(strfmt.DateTime(tenant.CreatedAt)).
		Description(tenant.Description).
		ID(tenant.ID).
		Name(tenant.Name).
		UpdatedAt(strfmt.DateTime(tenant.UpdatedAt)).
		Build()
}

// TenantCreateRequestToTenantEntity converts swagger.TenantCreateRequest to repo.TenantEntity
func TenantCreateRequestToTenantEntity(req *swagger.TenantCreateRequest) *repo.TenantEntity {
	now := time.Now()
	return repo.NewTenantEntityBuilder().
		ID(req.ID).
		Name(req.Name).
		Description(req.Description).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
}

// TenantUpdateRequestToTenantEntity converts swagger.TenantUpdateRequest to repo.TenantEntity
func TenantUpdateRequestToTenantEntity(tenantID string, req *swagger.TenantUpdateRequest) *repo.TenantEntity {
	return &repo.TenantEntity{
		ID:          tenantID,
		Name:        req.Name,
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}
}

// UserProfileEntityToUserProfileModel converts repo.UserProfileEntity to model.UserProfile
func UserProfileEntityToUserProfileModel(entity *repo.UserProfileEntity) *model.UserProfile {
	var birthDate *string
	if entity.BirthDate != nil {
		// Convert time.Time to string in YYYY-MM-DD format
		dateStr := entity.BirthDate.Format("2006-01-02")
		birthDate = &dateStr
	}

	return model.NewUserProfileBuilder().
		ID(*entity.ID).
		UserID(entity.UserID).
		Height(entity.Height).
		Weight(entity.Weight).
		Gender(entity.Gender).
		BirthDate(birthDate).
		IsMetric(entity.IsMetric).
		CreatedAt(entity.CreatedAt).
		UpdatedAt(entity.UpdatedAt).
		Build()
}

// UserProfileModelToUserProfileResponse converts model.UserProfile to swagger.UserProfile
func UserProfileModelToUserProfileResponse(profile *model.UserProfile) *swagger.UserProfile {
	var birthDate *strfmt.Date
	if profile.BirthDate != nil {
		// Parse the date string and convert to strfmt.Date
		if parsedTime, err := time.Parse("2006-01-02", *profile.BirthDate); err == nil {
			date := strfmt.Date(parsedTime)
			birthDate = &date
		}
	}

	return swagger.NewUserProfileBuilder().
		BirthDate(birthDate).
		CreatedAt(strfmt.DateTime(profile.CreatedAt)).
		Gender(profile.Gender).
		Height(convertIntPtrToInt64Ptr(profile.Height)).
		ID(int64(profile.ID)).
		IsMetric(profile.IsMetric).
		UpdatedAt(strfmt.DateTime(profile.UpdatedAt)).
		UserID(profile.UserID).
		Weight(convertIntPtrToInt64Ptr(profile.Weight)).
		Build()
}

// Helper function to convert *int to *int64
func convertIntPtrToInt64Ptr(intPtr *int) *int64 {
	if intPtr == nil {
		return nil
	}
	val := int64(*intPtr)
	return &val
}

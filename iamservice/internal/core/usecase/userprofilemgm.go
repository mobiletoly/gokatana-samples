package usecase

import (
	"context"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase/internal"
	"github.com/mobiletoly/gokatana/katapp"
)

// UserProfileMgm handles user profile management operations
type UserProfileMgm struct {
	ports *outport.Ports
}

func NewUserProfileMgm(ports *outport.Ports) *UserProfileMgm {
	return &UserProfileMgm{
		ports: ports,
	}
}

// GetUserProfileByUserID returns a user profile by user ID (admin only)
func (u *UserProfileMgm) GetUserProfileByUserID(
	ctx context.Context, principal *UserPrincipal, userID string,
) (*swagger.UserProfile, error) {
	katapp.Logger(ctx).Info("getting user profile by user ID", "userID", userID, "principal", principal.UserID)

	type UserWithProfile struct {
		user    *model.AuthUser
		profile *model.UserProfile
	}

	userWithProfile, err := outport.TxWithResult(ctx, u.ports.Tx, func(tx pgx.Tx) (*UserWithProfile, error) {
		user, err := internal.GetExistingUserById(ctx, u.ports.AuthUserPersist, tx, userID)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
		}
		if user == nil {
			return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
		}
		profile, err := u.ports.UserProfilePersist.GetUserProfileByUserID(ctx, tx, userID)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user profile")
		}
		if profile == nil {
			return nil, katapp.NewErr(katapp.ErrNotFound, "user profile not found")
		}
		return &UserWithProfile{user: user, profile: profile}, nil
	})
	if err != nil {
		return nil, err
	}
	if !principal.CanFetchUser(userID, userWithProfile.user.TenantID) {
		msg := "insufficient permissions to fetch user profile"
		katapp.Logger(ctx).Warn(msg,
			"principal", principal.String(),
			"targetUserID", userID,
		)
		return nil, katapp.NewErr(katapp.ErrNoPermissions, msg)
	}

	return userProfileModelToUserProfileResponse(userWithProfile.profile), nil
}

// UpdateUserProfileByUserID updates a user profile by user ID (admin only)
func (u *UserProfileMgm) UpdateUserProfileByUserID(
	ctx context.Context, principal *UserPrincipal, userID string, req *swagger.UserProfileUpdateRequest,
) (*swagger.UserProfile, error) {
	katapp.Logger(ctx).Info("updating user profile by user ID", "userID", userID, "principal", principal.UserID)

	// Get user and profile in a transaction
	type UserWithProfile struct {
		user    *model.AuthUser
		profile *model.UserProfile
	}

	userWithProfile, err := outport.TxWithResult(ctx, u.ports.Tx, func(tx pgx.Tx) (*UserWithProfile, error) {
		// First, get the target user to check tenant access
		user, err := u.ports.AuthUserPersist.GetUserByID(ctx, tx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
		}

		// Check authorization - users can update their own profile, admin/sysadmin can update any profile in their tenant
		if !principal.CanUpdateUserDetails(userID, user.TenantID) {
			return nil, katapp.NewErr(katapp.ErrNoPermissions, "insufficient permissions to update user profile")
		}

		// Update the user profile
		profile, err := u.ports.UserProfilePersist.UpdateUserProfile(ctx, tx, userID, req)
		if err != nil {
			return nil, err
		}

		return &UserWithProfile{user: user, profile: profile}, nil
	})

	if err != nil {
		return nil, err
	}

	return userProfileModelToUserProfileResponse(userWithProfile.profile), nil
}

// userProfileModelToUserProfileResponse converts model.UserProfile to swagger.UserProfile
func userProfileModelToUserProfileResponse(profile *model.UserProfile) *swagger.UserProfile {
	var birthDate *openapi_types.Date
	if profile.BirthDate != nil {
		if parsedTime, err := time.Parse("2006-01-02", *profile.BirthDate); err == nil {
			birthDate = &openapi_types.Date{Time: parsedTime}
		}
	}

	var upGender *swagger.UserProfileGender
	if profile.Gender != nil {
		gender := swagger.UserProfileGender(*profile.Gender)
		upGender = &gender
	}

	return swagger.NewUserProfileBuilder().
		BirthDate(birthDate).
		CreatedAt(profile.CreatedAt).
		Gender(upGender).
		Height(profile.Height).
		Id(profile.ID).
		IsMetric(profile.IsMetric).
		UpdatedAt(profile.UpdatedAt).
		UserId(profile.UserID).
		Weight(profile.Weight).
		Build()
}

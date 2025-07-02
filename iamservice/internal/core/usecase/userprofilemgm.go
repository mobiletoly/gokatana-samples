package usecase

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase/internal"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/samber/lo"
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
) (*swagger.UserProfileResponse, error) {
	katapp.Logger(ctx).Info("getting user profile by user ID", "userID", userID, "principal", principal.UserID)

	userWithProfile, err := outport.TxWithResult(
		ctx, u.ports.Tx,
		func(tx pgx.Tx) (lo.Tuple2[*model.AuthUser, *swagger.UserProfileResponse], error) {
			var t lo.Tuple2[*model.AuthUser, *swagger.UserProfileResponse]
			user, err := internal.GetExistingUserById(ctx, u.ports.AuthUserPersist, tx, userID)
			if err != nil {
				return t, katapp.NewErr(katapp.ErrInternal, "failed to get user")
			}
			if user == nil {
				return t, katapp.NewErr(katapp.ErrNotFound, "user not found")
			}
			profile, err := u.ports.UserProfilePersist.GetUserProfileByUserID(ctx, tx, userID)
			if err != nil {
				return t, katapp.NewErr(katapp.ErrInternal, "failed to get user profile")
			}
			if profile == nil {
				return t, katapp.NewErr(katapp.ErrNotFound, "user profile not found")
			}
			t.A, t.B = user, profile
			return t, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if !principal.CanFetchUser(userID, userWithProfile.A.TenantID) {
		msg := "insufficient permissions to fetch user profile"
		katapp.Logger(ctx).Warn(msg,
			"principal", principal.String(),
			"targetUserID", userID,
		)
		return nil, katapp.NewErr(katapp.ErrNoPermissions, msg)
	}

	return userWithProfile.B, nil
}

// UpdateUserProfileByUserID updates a user profile by user ID (admin only)
func (u *UserProfileMgm) UpdateUserProfileByUserID(
	ctx context.Context, principal *UserPrincipal, userID string, req *swagger.UpdateUserProfileRequest,
) (*swagger.UserProfileResponse, error) {
	katapp.Logger(ctx).Info("updating user profile by user ID", "userID", userID, "principal", principal.UserID)

	userWithProfile, err := outport.TxWithResult(
		ctx, u.ports.Tx,
		func(tx pgx.Tx) (lo.Tuple2[*model.AuthUser, *swagger.UserProfileResponse], error) {
			var t lo.Tuple2[*model.AuthUser, *swagger.UserProfileResponse]
			// First, get the target user to check tenant access
			user, err := u.ports.AuthUserPersist.GetUserByID(ctx, tx, userID)
			if err != nil {
				return t, err
			}
			if user == nil {
				return t, katapp.NewErr(katapp.ErrNotFound, "user not found")
			}

			// Check authorization - users can update their own profile, admin/sysadmin can update any profile in their tenant
			if !principal.CanUpdateUserDetails(userID, user.TenantID) {
				return t, katapp.NewErr(katapp.ErrNoPermissions, "insufficient permissions to update user profile")
			}
			// Update the user profile
			profile, err := u.ports.UserProfilePersist.UpdateUserProfile(ctx, tx, userID, req)
			if err != nil {
				return t, err
			}
			t.A, t.B = user, profile
			return t, nil
		})

	if err != nil {
		return nil, err
	}

	return userWithProfile.B, nil
}

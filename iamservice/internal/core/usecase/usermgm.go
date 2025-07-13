package usecase

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase/internal"
	"github.com/oapi-codegen/runtime/types"
	"time"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/katapp"
)

// UserMgm handles user management use cases
type UserMgm struct {
	authUserPort outport.AuthUserPersist
	txPort       outport.TxPort
}

// NewUserMgm creates a new UserMgm use case
func NewUserMgm(authUserPort outport.AuthUserPersist, databasePort outport.TxPort) *UserMgm {
	return &UserMgm{
		authUserPort: authUserPort,
		txPort:       databasePort,
	}
}

// LoadUserByID returns a user by ID (admin only)
func (u *UserMgm) LoadUserByID(
	ctx context.Context, principal *UserPrincipal, userID string,
) (*swagger.AuthUserResponse, error) {
	katapp.Logger(ctx).Info("loading user by ID", "userID", userID)
	if userID == "" {
		msg := "user id cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID)
		return nil, katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	user, err := outport.TxWithResult(ctx, u.txPort, func(tx pgx.Tx) (*model.AuthUser, error) {
		return u.authUserPort.GetUserByID(ctx, tx, userID)
	})
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}
	if !principal.CanFetchUser(userID, user.TenantID) {
		msg := "insufficient permissions to fetch user profile"
		katapp.Logger(ctx).Warn(msg,
			"principal", principal.String(),
			"targetUserID", userID,
		)
		return nil, katapp.NewErr(katapp.ErrNoPermissions, msg)
	}
	return authUserToAuthUserResponse(user), nil
}

// ListAllUsersByTenant returns a paginated list of users within principal's tenant (admin role only)
// if userPrincipal is user role only, then only the user's own profile is returned
func (u *UserMgm) ListAllUsersByTenant(
	ctx context.Context, userPrincipal *UserPrincipal, tenantID string, page, limit int,
) (*swagger.AuthUsersResponse, error) {
	katapp.Logger(ctx).Info("listing users by tenant",
		"principal", userPrincipal.String(),
		"tenantID", tenantID,
		"page", page,
		"limit", limit,
	)

	if tenantID == "" {
		msg := "tenant id cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", userPrincipal.String(), "tenantID", tenantID)
		return nil, katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 100
	}

	users, err := outport.TxWithResult(ctx, u.txPort, func(tx pgx.Tx) ([]*model.AuthUser, error) {
		if userPrincipal.CanListUsersForTenant(tenantID) {
			return u.authUserPort.GetAllUsersByTenantID(ctx, tx, tenantID)
		} else {
			var user *model.AuthUser
			user, err := u.authUserPort.GetUserByID(ctx, tx, userPrincipal.UserID)
			if err != nil {
				return nil, err
			}
			return []*model.AuthUser{user}, nil
		}
	})
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get users")
	}

	// TODO For now, we'll implement a simple version without pagination in the persist layer
	// In a real system, you'd add pagination support to the persist layer

	// Convert to swagger models
	userResponses := make([]swagger.AuthUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *authUserToAuthUserResponse(user)
	}

	paginatedUsers, pagination := internal.Paginate(userResponses, page, limit)
	return swagger.NewAuthUsersResponseBuilder().
		Items(paginatedUsers).
		Pagination(*pagination).
		Build(), nil
}

// ListAllUsers returns a paginated list of all users in the system (sysadmin role only)
func (u *UserMgm) ListAllUsers(
	ctx context.Context, principal *UserPrincipal, page, limit int,
) (*swagger.AuthUsersResponse, error) {
	katapp.Logger(ctx).Info("listing all users",
		"principal", principal.String(),
		"page", page,
		"limit", limit)

	if !principal.IsSysAdmin() {
		msg := "insufficient permissions to list all users"
		katapp.Logger(ctx).Warn(msg, "principal", principal.String())
		return nil, katapp.NewErr(katapp.ErrNoPermissions, msg)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 100
	}

	users, err := outport.TxWithResult(ctx, u.txPort, func(tx pgx.Tx) ([]*model.AuthUser, error) {
		return u.authUserPort.GetAllUsers(ctx, tx)
	})
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get users")
	}

	// Convert to swagger models
	userResponses := make([]swagger.AuthUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *authUserToAuthUserResponse(user)
	}

	paginatedUsers, pagination := internal.Paginate(userResponses, page, limit)
	return swagger.NewAuthUsersResponseBuilder().
		Items(paginatedUsers).
		Pagination(*pagination).
		Build(), nil
}

// GetUserRoles returns the roles assigned to a user
func (u *UserMgm) GetUserRoles(
	ctx context.Context, principal *UserPrincipal, userID string,
) (*swagger.UserRolesResponse, error) {
	katapp.Logger(ctx).Info("getting user roles",
		"principal", principal.String(),
		"userID", userID,
	)
	if userID == "" {
		msg := "userID cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID)
		return nil, katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	roles, err := outport.TxWithResult(ctx, u.txPort, func(tx pgx.Tx) ([]string, error) {
		user, err := internal.GetExistingUserById(ctx, u.authUserPort, tx, userID)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
		}

		if !principal.CanFetchUser(userID, user.TenantID) {
			msg := "insufficient permissions to get user roles"
			katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "targetUserID", userID)
			return nil, katapp.NewErr(katapp.ErrNoPermissions, msg)
		}

		roles, err := u.authUserPort.GetUserRoles(ctx, tx, userID)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user roles")
		}
		return roles, nil
	})
	if err != nil {
		return nil, err
	}

	return swagger.NewUserRolesResponseBuilder().
		Roles(roles).
		UserId(userID).
		Build(), nil
}

// AssignUserRole assigns a role to a user (admin only)
func (u *UserMgm) AssignUserRole(ctx context.Context, principal *UserPrincipal, userID string, roleName string) error {
	katapp.Logger(ctx).Info("assigning role to user",
		"principal", principal.String(),
		"userID", userID,
		"roleName", roleName,
	)
	if userID == "" {
		msg := "user id cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID, "roleName", roleName)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}
	if roleName == "" {
		msg := "role name cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID, "roleName", roleName)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	return u.txPort.Run(ctx, func(tx pgx.Tx) error {
		user, err := internal.GetExistingUserById(ctx, u.authUserPort, tx, userID)
		if err != nil {
			return err
		}
		if !principal.CanManageUser(user.TenantID) {
			msg := "insufficient permissions to assign roles"
			katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "userID", userID)
			return katapp.NewErr(katapp.ErrNoPermissions, msg)
		}

		if roleName == "sysadmin" {
			msg := "cannot assign sysadmin role"
			katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "userID", userID)
			return katapp.NewErr(katapp.ErrNoPermissions, msg)
		}

		err = u.authUserPort.AssignUserRole(ctx, tx, userID, roleName)
		if err != nil {
			var appErr *katapp.Err
			if errors.As(err, &appErr) && (appErr.Scope == katapp.ErrDuplicate || appErr.Scope == katapp.ErrNotFound) {
				return err
			}
			return katapp.NewErr(katapp.ErrInternal, "failed to assign role")
		}
		return nil
	})
}

// DeleteUserRole removes a role from a user (admin only)
func (u *UserMgm) DeleteUserRole(ctx context.Context, principal *UserPrincipal, userID string, roleName string) error {
	katapp.Logger(ctx).Info("removing role from user",
		"principal", principal.String(),
		"userID", userID,
		"roleName", roleName,
	)

	if userID == "" {
		msg := "user id cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID, "roleName", roleName)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}
	if roleName == "" {
		msg := "role name cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID, "roleName", roleName)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	return u.txPort.Run(ctx, func(tx pgx.Tx) error {
		user, err := internal.GetExistingUserById(ctx, u.authUserPort, tx, userID)
		if err != nil {
			return err
		}
		if !principal.CanManageUser(user.TenantID) {
			msg := "insufficient permissions to assign roles"
			katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "userID", userID)
			return katapp.NewErr(katapp.ErrNoPermissions, msg)
		}

		err = u.authUserPort.DeleteUserRole(ctx, tx, userID, roleName)
		if err != nil {
			var appErr *katapp.Err
			if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
				return katapp.NewErr(katapp.ErrNotFound, "invalid role name or user doesn't have this role")
			}
			return katapp.NewErr(katapp.ErrInternal, "failed to remove role")
		}

		return nil
	})
}

// DeleteUser deletes a user from the system (admin only)
func (u *UserMgm) DeleteUser(
	ctx context.Context, principal *UserPrincipal, userID string,
) error {
	katapp.Logger(ctx).Info("deleting user",
		"principal", principal.String(),
		"userID", userID,
	)
	if userID == "" {
		msg := "user id cannot be empty"
		katapp.Logger(ctx).Error(msg, "userID", userID)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	err := u.txPort.Run(ctx, func(tx pgx.Tx) error {
		user, err := internal.GetExistingUserById(ctx, u.authUserPort, tx, userID)
		if err != nil {
			return err
		}

		if !principal.CanManageUser(user.TenantID) {
			msg := "insufficient permissions to delete user"
			katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "userID", userID)
			return katapp.NewErr(katapp.ErrNoPermissions, msg)
		}

		err = u.authUserPort.DeleteUser(ctx, tx, userID)
		if err != nil {
			return katapp.NewErr(katapp.ErrInternal, "failed to delete user")
		}
		return nil
	})
	return err
}

// UpdateUserDetails updates user's details
func (u *UserMgm) UpdateUserDetails(
	ctx context.Context, principal *UserPrincipal, userID string, firstName, lastName string,
) error {
	katapp.Logger(ctx).Info("updating user details",
		"principal", principal.String(),
		"userID", userID,
		"firstName", firstName,
		"lastName", lastName,
	)

	if userID == "" {
		msg := "user id cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	if firstName == "" || lastName == "" {
		msg := "all fields are required"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	err := u.txPort.Run(ctx, func(tx pgx.Tx) error {
		// Check if user exists and get their details for authorization
		user, err := internal.GetExistingUserById(ctx, u.authUserPort, tx, userID)
		if err != nil {
			return err
		}

		// Check if the principal can manage this user
		if !principal.CanUpdateUserDetails(userID, user.TenantID) {
			msg := "insufficient permissions to update user details"
			katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "userID", userID)
			return katapp.NewErr(katapp.ErrNoPermissions, msg)
		}

		// Prepare updates map
		updates := map[string]interface{}{
			"first_name": firstName,
			"last_name":  lastName,
			"updated_at": time.Now(),
		}

		_, err = u.authUserPort.UpdateUser(ctx, tx, userID, updates)
		if err != nil {
			return katapp.NewErr(katapp.ErrInternal, "failed to update user details")
		}
		return nil
	})

	return err
}

// ChangeUserPassword changes a user's password (admin only)
func (u *UserMgm) ChangeUserPassword(
	ctx context.Context, principal *UserPrincipal, userID string, newPassword string,
) error {
	katapp.Logger(ctx).Info("changing user password",
		"principal", principal.String(),
		"userID", userID)

	if userID == "" {
		msg := "user id cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	if newPassword == "" {
		msg := "new password cannot be empty"
		katapp.Logger(ctx).Error(msg, "principal", principal.String(), "userID", userID)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	err := u.txPort.Run(ctx, func(tx pgx.Tx) error {
		user, err := internal.GetExistingUserById(ctx, u.authUserPort, tx, userID)
		if err != nil {
			return err
		}

		if !principal.CanUpdateUserDetails(userID, user.TenantID) {
			msg := "insufficient permissions to change user password"
			katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "userID", userID)
			return katapp.NewErr(katapp.ErrNoPermissions, msg)
		}

		// Hash the new password
		hashedPassword, err := internal.HashPassword(newPassword)
		if err != nil {
			return katapp.NewErr(katapp.ErrInternal, "failed to hash password")
		}

		// Prepare updates map
		updates := map[string]interface{}{
			"password_hash": hashedPassword,
			"updated_at":    time.Now(),
		}

		_, err = u.authUserPort.UpdateUser(ctx, tx, userID, updates)
		if err != nil {
			return katapp.NewErr(katapp.ErrInternal, "failed to update password")
		}
		return nil
	})

	return err
}

// Helper methods

// authUserToAuthUserResponse converts model.AuthUser to swagger.AuthUserResponse
func authUserToAuthUserResponse(user *model.AuthUser) *swagger.AuthUserResponse {
	createdAt := user.CreatedAt
	updatedAt := user.UpdatedAt

	return swagger.NewAuthUserResponseBuilder().
		CreatedAt(createdAt).
		Email(types.Email(user.Email)).
		FirstName(user.FirstName).
		Id(user.ID).
		LastName(user.LastName).
		TenantId(user.TenantID).
		UpdatedAt(updatedAt).
		Build()
}

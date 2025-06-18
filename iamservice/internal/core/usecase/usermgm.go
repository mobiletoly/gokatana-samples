package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/katapp"
)

// UserMgm handles user management use cases
type UserMgm struct {
	authUserPort outport.AuthUserPersist
}

// NewUserMgm creates a new UserMgm use case
func NewUserMgm(authUserPort outport.AuthUserPersist) *UserMgm {
	return &UserMgm{
		authUserPort: authUserPort,
	}
}

// GetCurrentUserProfile returns the profile of the authenticated user
func (u *UserMgm) GetCurrentUserProfile(ctx context.Context, userID string) (*swagger.UserProfile, error) {
	// userID should never be empty - this is a programming error if it happens
	if userID == "" {
		panic("userID cannot be empty - this should be validated at the HTTP handler level")
	}

	user, err := u.authUserPort.GetUserByID(ctx, userID)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, err
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user profile")
	}

	return u.authUserToUserProfile(user), nil
}

// GetUserByID returns a user by ID (admin only)
func (u *UserMgm) GetUserByID(ctx context.Context, userID string) (*swagger.UserProfile, error) {
	if userID == "" {
		panic("userID cannot be empty - this should be validated at the HTTP handler level")
	}

	user, err := u.authUserPort.GetUserByID(ctx, userID)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, err
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}

	return u.authUserToUserProfile(user), nil
}

// ListUsers returns a paginated list of users (admin only)
func (u *UserMgm) ListUsers(ctx context.Context, page, limit int) (*swagger.UserListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// For now, we'll implement a simple version without pagination in the persist layer
	// In a real system, you'd add pagination support to the persist layer
	users, err := u.authUserPort.GetAllUsers(ctx)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get users")
	}

	// Convert to swagger models
	userProfiles := make([]*swagger.UserProfile, len(users))
	for i, user := range users {
		userProfiles[i] = u.authUserToUserProfile(user)
	}

	// Simple pagination logic (in production, this should be done in the database)
	total := len(userProfiles)
	totalPages := (total + limit - 1) / limit

	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	paginatedUsers := userProfiles[start:end]

	// Convert to int64 for swagger models
	pageInt64 := int64(page)
	limitInt64 := int64(limit)
	totalInt64 := int64(total)
	totalPagesInt64 := int64(totalPages)

	pagination := swagger.NewPaginationInfoBuilder().
		Page(&pageInt64).
		Limit(&limitInt64).
		Total(&totalInt64).
		TotalPages(&totalPagesInt64).
		Build()

	return swagger.NewUserListResponseBuilder().
		Users(paginatedUsers).
		Pagination(pagination).
		Build(), nil
}

// GetUserRoles returns the roles assigned to a user (admin only)
func (u *UserMgm) GetUserRoles(ctx context.Context, userID string) (*swagger.UserRolesResponse, error) {
	// userID should never be empty - this is a programming error if it happens
	if userID == "" {
		panic("userID cannot be empty - this should be validated at the HTTP handler level")
	}

	// Check if user exists
	_, err := u.authUserPort.GetUserByID(ctx, userID)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, err
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}

	roles, err := u.authUserPort.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user roles")
	}

	return swagger.NewUserRolesResponseBuilder().
		UserID(&userID).
		Roles(roles).
		Build(), nil
}

// AssignUserRole assigns a role to a user (admin only)
func (u *UserMgm) AssignUserRole(ctx context.Context, userID string, roleName string, requestingUserID string) (*swagger.MessageResponse, error) {
	// These should never be empty - programming errors if they happen
	if userID == "" {
		panic("userID cannot be empty - this should be validated at the HTTP handler level")
	}
	if roleName == "" {
		panic("roleName cannot be empty - this should be validated at the HTTP handler level")
	}

	// Check if user exists
	_, err := u.authUserPort.GetUserByID(ctx, userID)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, err
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}

	err = u.authUserPort.AssignUserRole(ctx, userID, roleName, &requestingUserID)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrDuplicate {
			return nil, err
		}
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, err
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to assign role")
	}

	message := "Role assigned successfully"
	return swagger.NewMessageResponseBuilder().
		Message(&message).
		Build(), nil
}

// DeleteUserRole removes a role from a user (admin only)
func (u *UserMgm) DeleteUserRole(ctx context.Context, userID string, roleName string) (*swagger.MessageResponse, error) {
	// These should never be empty - programming errors if they happen
	if userID == "" {
		panic("userID cannot be empty - this should be validated at the HTTP handler level")
	}
	if roleName == "" {
		panic("roleName cannot be empty - this should be validated at the HTTP handler level")
	}

	// Check if user exists
	_, err := u.authUserPort.GetUserByID(ctx, userID)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, err
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}

	err = u.authUserPort.DeleteUserRole(ctx, userID, roleName)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, katapp.NewErr(katapp.ErrNotFound, "invalid role name or user doesn't have this role")
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to remove role")
	}

	message := "Role removed successfully"
	return swagger.NewMessageResponseBuilder().
		Message(&message).
		Build(), nil
}

// DeleteUser deletes a user from the system (admin only)
func (u *UserMgm) DeleteUser(ctx context.Context, userID string) (*swagger.MessageResponse, error) {
	// userID should never be empty - programming error if it happens
	if userID == "" {
		panic("userID cannot be empty - this should be validated at the HTTP handler level")
	}

	// Check if user exists
	_, err := u.authUserPort.GetUserByID(ctx, userID)
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, err
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}

	err = u.authUserPort.DeleteUser(ctx, userID)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to delete user")
	}

	message := "User deleted successfully"
	return swagger.NewMessageResponseBuilder().
		Message(&message).
		Build(), nil
}

// Helper methods

// authUserToUserProfile converts model.AuthUser to swagger.UserProfile
func (u *UserMgm) authUserToUserProfile(user *model.AuthUser) *swagger.UserProfile {
	// Convert types to match swagger expectations
	email := strfmt.Email(user.Email)

	// Parse time strings to proper DateTime format
	createdAtTime, err := time.Parse(time.RFC3339, user.CreatedAt)
	if err != nil {
		// Fallback to current time if parsing fails
		createdAtTime = time.Now()
	}
	updatedAtTime, err := time.Parse(time.RFC3339, user.UpdatedAt)
	if err != nil {
		// Fallback to current time if parsing fails
		updatedAtTime = time.Now()
	}

	createdAt := strfmt.DateTime(createdAtTime)
	updatedAt := strfmt.DateTime(updatedAtTime)

	return swagger.NewUserProfileBuilder().
		ID(&user.ID).
		Email(&email).
		FirstName(&user.FirstName).
		LastName(&user.LastName).
		CreatedAt(&createdAt).
		UpdatedAt(&updatedAt).
		Build()
}

package persist

import (
	"context"
	"github.com/google/uuid"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/persist/internal/mapper"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/persist/internal/repo"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/katpg"
)

// AuthUserAdapter implements the AuthUserPersist outport interface
type AuthUserAdapter struct {
	authRepo *repo.AuthUserRepo
}

// NewAuthUserAdapter creates a new AuthUserAdapter
func NewAuthUserAdapter(db *katpg.DBLink) outport.AuthUserPersist {
	return &AuthUserAdapter{
		authRepo: repo.NewAuthRepo(db.Pool),
	}
}

// CreateUser creates a new user in the auth_user table
func (a *AuthUserAdapter) CreateUser(ctx context.Context, req *swagger.SignupRequest) (*model.AuthUser, error) {
	// Generate UUID for user ID
	userID := uuid.NewString()

	userEntity := mapper.SwaggerSignupRequestToAuthUserEntity(req, userID, *req.Password)

	err := a.authRepo.InsertUser(ctx, userEntity)
	if err != nil {
		katapp.Logger(ctx).Error("failed to create user", "error", err)
		appErr := katpg.ToAppError(err, "database insert failed")
		if appErr.Scope == katapp.ErrDuplicate {
			return nil, katapp.NewErr(katapp.ErrDuplicate, "user with this email already exists")
		}
		return nil, appErr
	}

	return mapper.AuthUserEntityToAuthUserModel(userEntity), nil
}

// GetUserByEmail retrieves a user by email
func (a *AuthUserAdapter) GetUserByEmail(ctx context.Context, email string) (*model.AuthUser, error) {
	userEntity, err := a.authRepo.SelectUserByEmail(ctx, email)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get user by email", "email", email, "error", err)
		appErr := katpg.ToAppError(err, "failed to select user by email")
		return nil, appErr
	}

	if userEntity == nil {
		return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
	}

	return mapper.AuthUserEntityToAuthUserModel(userEntity), nil
}

// GetUserByID retrieves a user by ID
func (a *AuthUserAdapter) GetUserByID(ctx context.Context, userID string) (*model.AuthUser, error) {
	userEntity, err := a.authRepo.SelectUserByID(ctx, userID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get user by ID", "userID", userID, "error", err)
		appErr := katpg.ToAppError(err, "failed to select user by ID")
		return nil, appErr
	}

	if userEntity == nil {
		return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
	}

	return mapper.AuthUserEntityToAuthUserModel(userEntity), nil
}

// UpdateUser updates user information
func (a *AuthUserAdapter) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) (*model.AuthUser, error) {
	err := a.authRepo.UpdateUser(ctx, userID, updates)
	if err != nil {
		katapp.Logger(ctx).Error("failed to update user", "userID", userID, "error", err)
		appErr := katpg.ToAppError(err, "failed to update user")
		return nil, appErr
	}

	return a.GetUserByID(ctx, userID)
}

// DeleteUser deletes a user from the database
func (a *AuthUserAdapter) DeleteUser(ctx context.Context, userID string) error {
	err := a.authRepo.DeleteUser(ctx, userID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to delete user", "userID", userID, "error", err)
		appErr := katpg.ToAppError(err, "failed to delete user")
		return appErr
	}

	return nil
}

// GetUserWithPasswordByEmail retrieves a user by email including the password hash
func (a *AuthUserAdapter) GetUserWithPasswordByEmail(ctx context.Context, email string) (*model.AuthUser, error) {
	userEntity, err := a.authRepo.SelectUserWithPasswordByEmail(ctx, email)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get user with password by email", "email", email, "error", err)
		appErr := katpg.ToAppError(err, "failed to select user with password by email")
		return nil, appErr
	}

	if userEntity == nil {
		return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
	}

	return mapper.AuthUserEntityToAuthUserModel(userEntity), nil
}

// GetAllUsers retrieves all users from the database
func (a *AuthUserAdapter) GetAllUsers(ctx context.Context) ([]*model.AuthUser, error) {
	userEntities, err := a.authRepo.SelectAllUsers(ctx)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get all users", "error", err)
		appErr := katpg.ToAppError(err, "failed to select all users")
		return nil, appErr
	}

	users := make([]*model.AuthUser, len(userEntities))
	for i, entity := range userEntities {
		users[i] = mapper.AuthUserEntityToAuthUserModel(&entity)
	}

	return users, nil
}

// Role management methods

// GetUserRoles retrieves all role names for a user
func (a *AuthUserAdapter) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	roleEntities, err := a.authRepo.SelectUserRoles(ctx, userID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get user roles", "userID", userID, "error", err)
		appErr := katpg.ToAppError(err, "failed to select user roles")
		return nil, appErr
	}

	roles := make([]string, len(roleEntities))
	for i, role := range roleEntities {
		roles[i] = role.Name
	}

	return roles, nil
}

// AssignUserRole assigns a role to a user
func (a *AuthUserAdapter) AssignUserRole(ctx context.Context, userID string, roleName string, assignedBy *string) error {
	// First get the role ID by name
	roleEntity, err := a.authRepo.SelectRoleByName(ctx, roleName)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get role by name", "roleName", roleName, "error", err)
		appErr := katpg.ToAppError(err, "failed to select role by name")
		return appErr
	}

	if roleEntity == nil {
		return katapp.NewErr(katapp.ErrNotFound, "role not found")
	}

	// Assign the role to the user
	err = a.authRepo.InsertUserRole(ctx, userID, *roleEntity.ID, assignedBy)
	if err != nil {
		katapp.Logger(ctx).Error("failed to assign user role", "userID", userID, "roleID", *roleEntity.ID, "error", err)
		appErr := katpg.ToAppError(err, "failed to assign user role")
		if appErr.Scope == katapp.ErrDuplicate {
			return katapp.NewErr(katapp.ErrDuplicate, "user already has this role")
		}
		return appErr
	}

	return nil
}

// DeleteUserRole removes a role from a user
func (a *AuthUserAdapter) DeleteUserRole(ctx context.Context, userID string, roleName string) error {
	// First get the role ID by name
	roleEntity, err := a.authRepo.SelectRoleByName(ctx, roleName)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get role by name", "roleName", roleName, "error", err)
		appErr := katpg.ToAppError(err, "failed to select role by name")
		return appErr
	}

	if roleEntity == nil {
		return katapp.NewErr(katapp.ErrNotFound, "role not found")
	}

	// Remove the role from the user
	err = a.authRepo.DeleteUserRole(ctx, userID, *roleEntity.ID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to remove user role", "userID", userID, "roleID", *roleEntity.ID, "error", err)
		appErr := katpg.ToAppError(err, "failed to remove user role")
		return appErr
	}

	return nil
}

package outport

import (
	"context"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
)

// AuthUserPersist defines the outport interface for authentication operations
type AuthUserPersist interface {
	CreateUser(ctx context.Context, user *swagger.SignupRequest) (*model.AuthUser, error)
	GetUserByEmail(ctx context.Context, email string) (*model.AuthUser, error)
	GetUserByID(ctx context.Context, userID string) (*model.AuthUser, error)
	UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) (*model.AuthUser, error)
	DeleteUser(ctx context.Context, userID string) error

	GetUserWithPasswordByEmail(ctx context.Context, email string) (*model.AuthUser, error)
	GetAllUsers(ctx context.Context) ([]*model.AuthUser, error)

	GetUserRoles(ctx context.Context, userID string) ([]string, error)
	AssignUserRole(ctx context.Context, userID string, roleName string, assignedBy *string) error
	DeleteUserRole(ctx context.Context, userID string, roleName string) error
}

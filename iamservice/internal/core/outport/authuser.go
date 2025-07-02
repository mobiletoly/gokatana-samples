package outport

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"time"
)

// AuthUserPersist defines the outport interface for authentication operations
type AuthUserPersist interface {
	CreateUser(ctx context.Context, tx pgx.Tx, user *swagger.SignupRequest, tenantID string) (*model.AuthUser, error)
	GetUserByEmail(ctx context.Context, tx pgx.Tx, email string, tenantID string) (*model.AuthUser, error)
	GetUserByID(ctx context.Context, tx pgx.Tx, userID string) (*model.AuthUser, error)
	UpdateUser(ctx context.Context, tx pgx.Tx, userID string, updates map[string]interface{}) (*model.AuthUser, error)
	DeleteUser(ctx context.Context, tx pgx.Tx, userID string) error

	GetUserWithPasswordByEmail(ctx context.Context, tx pgx.Tx, email string, tenantID string) (*model.AuthUser, error)
	GetAllUsersByTenantID(ctx context.Context, tx pgx.Tx, tenantID string) ([]*model.AuthUser, error)
	GetAllUsers(ctx context.Context, tx pgx.Tx) ([]*model.AuthUser, error)

	GetUserRoles(ctx context.Context, tx pgx.Tx, userID string) ([]string, error)
	AssignUserRole(ctx context.Context, tx pgx.Tx, userID string, roleName string) error
	DeleteUserRole(ctx context.Context, tx pgx.Tx, userID string, roleName string) error

	// Tenant operations
	GetTenantByID(ctx context.Context, tx pgx.Tx, tenantID string) (*model.Tenant, error)
	GetAllTenants(ctx context.Context, tx pgx.Tx) ([]*model.Tenant, error)
	CreateTenant(ctx context.Context, tx pgx.Tx, tenant *swagger.CreateTenantRequest) (*model.Tenant, error)
	UpdateTenant(ctx context.Context, tx pgx.Tx, tenantID string, tenant *swagger.UpdateTenantRequest) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, tx pgx.Tx, tenantID string) error

	// Email confirmation
	CreateEmailConfirmationToken(ctx context.Context, tx pgx.Tx, userID string, email string, tokenHash string, source string, expiresAt time.Time) (*model.EmailConfirmationToken, error)
	GetEmailConfirmationTokenByUserIDAndHash(ctx context.Context, tx pgx.Tx, userID string, tokenHash string) (*model.EmailConfirmationToken, error)
	MarkEmailConfirmationTokenAsUsed(ctx context.Context, tx pgx.Tx, tokenID string) error
	SetUserEmailVerified(ctx context.Context, tx pgx.Tx, userID string, verified bool) error
}

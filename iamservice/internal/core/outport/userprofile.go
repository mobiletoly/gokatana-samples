package outport

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
)

// UserProfilePersist defines the outport interface for user profile operations
type UserProfilePersist interface {
	GetUserProfileByUserID(ctx context.Context, tx pgx.Tx, userID string) (*swagger.UserProfileResponse, error)
	CreateUserProfile(ctx context.Context, tx pgx.Tx, userID string) (*swagger.UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, tx pgx.Tx, userID string, req *swagger.UpdateUserProfileRequest) (*swagger.UserProfileResponse, error)
}

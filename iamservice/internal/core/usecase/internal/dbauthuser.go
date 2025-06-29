package internal

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana/katapp"
	"golang.org/x/crypto/bcrypt"
)

func GetExistingUserById(
	ctx context.Context, authUserPort outport.AuthUserPersist, tx pgx.Tx, userID string,
) (*model.AuthUser, error) {
	user, err := authUserPort.GetUserByID(ctx, tx, userID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get existing user", "userID", userID, "error", err)
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}
	if user == nil {
		return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
	}
	return user, nil
}

func EnsureUserExistsById(
	ctx context.Context, authUserPort outport.AuthUserPersist, tx pgx.Tx, userID string,
) error {
	_, err := GetExistingUserById(ctx, authUserPort, tx, userID)
	return err
}

func GetExistingTenantById(
	ctx context.Context, authUserPort outport.AuthUserPersist, tx pgx.Tx, tenantID string,
) (*model.Tenant, error) {
	tenant, err := authUserPort.GetTenantByID(ctx, tx, tenantID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to get existing tenant", "tenantID", tenantID, "error", err)
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get tenant")
	}
	if tenant == nil {
		return nil, katapp.NewErr(katapp.ErrNotFound, "tenant not found")
	}
	return tenant, nil
}

func EnsureTenantExistsById(
	ctx context.Context, authUserPort outport.AuthUserPersist, tx pgx.Tx, tenantID string,
) error {
	_, err := GetExistingTenantById(ctx, authUserPort, tx, tenantID)
	return err
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

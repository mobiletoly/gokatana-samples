package usecase

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase/internal"
	"github.com/mobiletoly/gokatana/katapp"
)

// SignIn authenticates a user and returns tokens
func (a *AuthMgm) SignIn(ctx context.Context, req *swagger.SigninRequest) (*swagger.AuthResponse, error) {
	if err := a.validateSigninRequest(req); err != nil {
		return nil, err
	}
	tenantID := req.TenantId

	user, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (*model.AuthUser, error) {
		// Check if tenant exists
		if err := internal.EnsureTenantExistsById(ctx, a.authUserPersist, tx, tenantID); err != nil {
			return nil, err
		}

		// Get user with password hash
		user, err := a.authUserPersist.GetUserWithPasswordByEmail(ctx, tx, string(req.Email), tenantID)
		if err != nil {
			var appErr *katapp.Err
			if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
				return nil, katapp.NewErr(katapp.ErrUnauthorized, "invalid credentials")
			}
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
		}
		return user, err
	})
	if err != nil {
		return nil, err
	}

	// Verify password
	if err := a.verifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, katapp.NewErr(katapp.ErrUnauthorized, "invalid credentials")
	}

	// Check if email is verified
	if !user.EmailVerified {
		return nil, katapp.NewErr(katapp.ErrUnauthorized, "email address not verified. Please check your email for confirmation instructions")
	}

	// Generate tokens with roles
	accessToken, refreshToken, expiresIn, err := a.generateJWTTokenForUser(ctx, user)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to generate tokens")
	}

	// Build response
	tokenType := "Bearer"
	return swagger.NewAuthResponseBuilder().
		AccessToken(accessToken).
		ExpiresIn(expiresIn).
		RefreshToken(refreshToken).
		TokenType(tokenType).
		UserId(user.ID).
		Build(), nil
}

func (a *AuthMgm) validateSigninRequest(req *swagger.SigninRequest) error {
	if req.Email == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "email is required")
	}
	if req.Password == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "password is required")
	}
	if req.TenantId == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "tenant ID is required")
	}

	return nil
}

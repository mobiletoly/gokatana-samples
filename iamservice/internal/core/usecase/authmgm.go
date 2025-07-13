package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase/internal"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/katapp"
	"golang.org/x/crypto/bcrypt"
)

// AuthMgm provides authentication use cases
type AuthMgm struct {
	serverConfig    *katapp.ServerConfig
	authUserPersist outport.AuthUserPersist
	txPort          outport.TxPort
	mailer          outport.Mailer
	jwtSecret       []byte
}

// NewAuthUser creates a new AuthMgm use case
func NewAuthUser(
	serverConfig *katapp.ServerConfig, authUserPort outport.AuthUserPersist, databasePort outport.TxPort,
	mailer outport.Mailer, jwtSecret string,
) *AuthMgm {
	return &AuthMgm{
		serverConfig:    serverConfig,
		authUserPersist: authUserPort,
		txPort:          databasePort,
		mailer:          mailer,
		jwtSecret:       []byte(jwtSecret),
	}
}

// RefreshToken generates new tokens using a refresh token with rotation
func (a *AuthMgm) RefreshToken(ctx context.Context, req *swagger.TokenRefreshRequest) (*swagger.SignInResponse, error) {
	if req.RefreshToken == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "refresh token is required")
	}

	// Hash the provided refresh token for database lookup
	tokenHash := a.hashRefreshToken(req.RefreshToken)

	// Validate refresh token and get user in a transaction
	result, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (*swagger.SignInResponse, error) {
		// Get refresh token from database
		refreshTokenRecord, err := a.authUserPersist.GetRefreshTokenByHash(ctx, tx, tokenHash)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to validate refresh token")
		}
		if refreshTokenRecord == nil || !refreshTokenRecord.IsValid() {
			return nil, katapp.NewErr(katapp.ErrUnauthorized, "invalid or expired refresh token")
		}

		// Get user
		user, err := a.authUserPersist.GetUserByID(ctx, tx, refreshTokenRecord.UserID)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
		}
		if user == nil {
			return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
		}

		// Revoke the old refresh token immediately (rotation)
		err = a.authUserPersist.RevokeRefreshToken(ctx, tx, tokenHash)
		if err != nil {
			katapp.Logger(ctx).Error("failed to revoke old refresh token", "error", err)
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to revoke old refresh token")
		}

		// Generate new tokens with roles
		accessToken, newRefreshToken, expiresIn, err := a.generateJWTTokenForUserWithTx(ctx, tx, user)
		if err != nil {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to generate tokens")
		}

		// Build response
		tokenType := "Bearer"
		return swagger.NewSignInResponseBuilder().
				AccessToken(accessToken).
				ExpiresIn(expiresIn).
				RefreshToken(newRefreshToken).
				TokenType(tokenType).
				UserId(user.ID).
				Build(),
			nil
	})

	return result, err
}

// ValidateAccessToken validates an access token and returns the user ID
func (a *AuthMgm) ValidateAccessToken(token string) (string, error) {
	if token == "" {
		return "", katapp.NewErr(katapp.ErrUnauthorized, "access token is required")
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	userID, err := a.getUserIDFromAccessToken(token)
	if err != nil {
		return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid or expired access token")
	}

	return userID, nil
}

// ValidateUserPasswordMatches validates a user's password
func (a *AuthMgm) ValidateUserPasswordMatches(
	ctx context.Context, userID string, currentPassword string,
) error {
	katapp.Logger(ctx).Info("validating user password", "userID", userID)

	if userID == "" {
		msg := "user id cannot be empty"
		katapp.Logger(ctx).Error(msg, "userID", userID)
		return katapp.NewErr(katapp.ErrInvalidInput, msg)
	}

	err := a.txPort.Run(ctx, func(tx pgx.Tx) error {
		user, err := internal.GetExistingUserById(ctx, a.authUserPersist, tx, userID)
		if err != nil {
			return err
		}

		if err := a.verifyPassword(user.PasswordHash, currentPassword); err != nil {
			return katapp.NewErr(katapp.ErrUnauthorized, "current password is incorrect")
		}
		return nil
	})

	return err
}

// JWT helper methods

// generateJWTTokenForUser generates JWT access and refresh tokens including user roles and tenant ID
func (a *AuthMgm) generateJWTTokenForUser(
	ctx context.Context, user *model.AuthUser,
) (accessToken string, refreshToken string, expiresIn int64, err error) {
	result, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (struct {
		AccessToken  string
		RefreshToken string
		ExpiresIn    int64
	}, error) {
		accessToken, refreshToken, expiresIn, err := a.generateJWTTokenForUserWithTx(ctx, tx, user)
		return struct {
			AccessToken  string
			RefreshToken string
			ExpiresIn    int64
		}{accessToken, refreshToken, expiresIn}, err
	})
	if err != nil {
		return "", "", 0, err
	}
	return result.AccessToken, result.RefreshToken, result.ExpiresIn, nil
}

// generateJWTTokenForUserWithTx generates JWT tokens within a transaction and persists refresh token
func (a *AuthMgm) generateJWTTokenForUserWithTx(
	ctx context.Context, tx pgx.Tx, user *model.AuthUser,
) (accessToken string, refreshToken string, expiresIn int64, err error) {
	now := time.Now()
	expiresIn = 3600 // 1 hour

	roles, err := a.authUserPersist.GetUserRoles(ctx, tx, user.ID)
	if err != nil {
		katapp.Logger(ctx).Warn("failed to get user roles for token generation", "userID", user.ID, "error", err)
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to get user roles")
	}

	// Generate unique nonces to ensure tokens are always different
	accessNonce := a.generateTokenNonce()
	refreshNonce := a.generateTokenNonce()

	// Generate access token with roles and tenant ID
	accessClaims := jwt.MapClaims{
		"sub":      user.ID,
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(expiresIn) * time.Second).Unix(),
		"type":     "access",
		"roles":    roles,
		"tenantId": user.TenantID,
		"nonce":    accessNonce,
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate access token")
	}

	// Generate refresh token (valid for 30 days) - no roles needed in refresh token
	refreshExpiresAt := now.Add(30 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"sub":   user.ID,
		"iat":   now.Unix(),
		"exp":   refreshExpiresAt.Unix(),
		"type":  "refresh",
		"nonce": refreshNonce,
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate refresh token")
	}

	// Clean up old refresh tokens: remove all revoked tokens and keep only the 2 most recent non-revoked tokens
	rowsAffected, err := a.authUserPersist.CleanupUserRefreshTokens(ctx, tx, user.ID)
	if err != nil {
		katapp.Logger(ctx).Error("failed to cleanup user refresh tokens", "userID", user.ID, "error", err)
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to cleanup user refresh tokens")
	}
	katapp.Logger(ctx).Debug("cleaned up old refresh tokens", "userID", user.ID, "rowsAffected", rowsAffected)

	// Persist refresh token in database
	tokenHash := a.hashRefreshToken(refreshToken)
	_, err = a.authUserPersist.CreateRefreshToken(ctx, tx, user.ID, tokenHash, refreshExpiresAt)
	if err != nil {
		katapp.Logger(ctx).Error("failed to persist refresh token", "userID", user.ID, "error", err)
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to persist refresh token")
	}

	return accessToken, refreshToken, expiresIn, nil
}

// getUserIDFromAccessToken validates an access token and returns the user ID
func (a *AuthMgm) getUserIDFromAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
			return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid token type")
		}

		if userID, ok := claims["sub"].(string); ok {
			return userID, nil
		}
	}

	return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid token claims")
}

// getUserIDFromRefreshToken validates a refresh token and returns the user ID
func (a *AuthMgm) getUserIDFromRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid refresh token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
			return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid token type")
		}

		if userID, ok := claims["sub"].(string); ok {
			return userID, nil
		}
	}

	return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid refresh token claims")
}

// SignOut revokes all refresh tokens for a user
func (a *AuthMgm) SignOut(ctx context.Context, userID string) error {
	return a.txPort.Run(ctx, func(tx pgx.Tx) error {
		err := a.authUserPersist.RevokeAllUserRefreshTokens(ctx, tx, userID)
		if err != nil {
			katapp.Logger(ctx).Error("failed to revoke all user refresh tokens", "userID", userID, "error", err)
			return katapp.NewErr(katapp.ErrInternal, "failed to sign out user")
		}
		return nil
	})
}

// hashRefreshToken creates SHA-256 hash of the refresh token for secure database storage
func (a *AuthMgm) hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// generateTokenNonce generates a random nonce for token uniqueness
func (a *AuthMgm) generateTokenNonce() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Password helper methods

// verifyPassword verifies a password against its hash using bcrypt
func (a *AuthMgm) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// hashToken creates SHA-256 hash of the token/code with user ID for secure database storage
func (a *AuthMgm) hashToken(userID string, token string) string {
	// Combine user ID and token to prevent collisions across users
	combined := userID + ":" + token
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

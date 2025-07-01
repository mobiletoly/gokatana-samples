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

// RefreshToken generates new tokens using a refresh token
func (a *AuthMgm) RefreshToken(ctx context.Context, req *swagger.RefreshRequest) (*swagger.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "refresh token is required")
	}

	// Validate refresh token
	userID, err := a.getUserIDFromRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrUnauthorized, "invalid or expired refresh token")
	}

	user, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (*model.AuthUser, error) {
		return a.authUserPersist.GetUserByID(ctx, tx, userID)
	})
	if err != nil {
		katapp.Logger(ctx).Error("failed to get user", "userID", userID, "error", err)
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}
	if user == nil {
		return nil, katapp.NewErr(katapp.ErrNotFound, "user not found")
	}

	// Generate new tokens with roles
	accessToken, refreshToken, expiresIn, err := a.generateJWTTokenForUser(ctx, user)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to generate tokens")
	}

	// Invalidate old refresh token
	_ = a.invalidateRefreshToken(req.RefreshToken)

	// Build response
	tokenType := "Bearer"
	return swagger.NewAuthResponseBuilder().
			AccessToken(accessToken).
			ExpiresIn(expiresIn).
			RefreshToken(refreshToken).
			TokenType(tokenType).
			UserId(user.ID).
			Build(),
		nil
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
	now := time.Now()
	expiresIn = 3600 // 1 hour

	roles, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) ([]string, error) {
		return a.authUserPersist.GetUserRoles(ctx, tx, user.ID)
	})
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

	// Generate refresh token (valid for 7 days) - no roles needed in refresh token
	refreshClaims := jwt.MapClaims{
		"sub":   user.ID,
		"iat":   now.Unix(),
		"exp":   now.Add(7 * 24 * time.Hour).Unix(),
		"type":  "refresh",
		"nonce": refreshNonce,
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate refresh token")
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

// invalidateRefreshToken invalidates a refresh token (placeholder implementation)
func (a *AuthMgm) invalidateRefreshToken(token string) error {
	// TODO In a production system, you would store invalidated tokens in a blacklist
	// For now, this is a no-op since JWT tokens are stateless
	return nil
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

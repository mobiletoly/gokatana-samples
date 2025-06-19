package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/katapp"
	"golang.org/x/crypto/bcrypt"
)

// AuthUser provides authentication use cases
type AuthUser struct {
	authUserPersist outport.AuthUserPersist
	tx              outport.Transaction
	jwtSecret       []byte
}

// NewAuthUser creates a new AuthUser use case
func NewAuthUser(
	authUserPersist outport.AuthUserPersist, tx outport.Transaction, jwtSecret string,
) *AuthUser {
	return &AuthUser{
		authUserPersist: authUserPersist,
		tx:              tx,
		jwtSecret:       []byte(jwtSecret),
	}
}

// SignUp creates a new user account
func (a *AuthUser) SignUp(ctx context.Context, req *swagger.SignupRequest) (*swagger.AuthResponse, error) {
	// Validate input
	if err := a.validateSignupRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := a.authUserPersist.GetUserByEmail(ctx, string(*req.Email))
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			// User not found, which is what we want for signup
		} else {
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to check existing user")
		}
	}
	if existingUser != nil {
		return nil, katapp.NewErr(katapp.ErrDuplicate, "user with this email already exists")
	}

	// Hash password
	hashedPassword, err := a.hashPassword(*req.Password)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to hash password")
	}

	// Create signup request with hashed password
	signupReq := &swagger.SignupRequest{
		Email:     req.Email,
		Password:  &hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Create user and assign role in a transaction
	user, err := outport.TxWithResult(ctx, a.tx, func() (*model.AuthUser, error) {
		// Create user
		user, err := a.authUserPersist.CreateUser(ctx, signupReq)
		if err != nil {
			return nil, err // Return the original error to preserve error type (e.g., ErrDuplicate)
		}

		// Assign default 'user' role to new user
		err = a.authUserPersist.AssignUserRole(ctx, user.ID, "user", nil)
		if err != nil {
			katapp.Logger(ctx).Warn("failed to assign default role to user", "userID", user.ID, "error", err)
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to assign default role")
		}

		return user, nil
	})

	if err != nil {
		return nil, err // Return the transaction error immediately
	}

	// Generate tokens with roles
	accessToken, refreshToken, expiresIn, err := a.generateTokensWithRoles(ctx, user.ID)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to generate tokens")
	}

	// Build response
	tokenType := "Bearer"

	return swagger.NewAuthResponseBuilder().
		AccessToken(&accessToken).
		RefreshToken(&refreshToken).
		TokenType(&tokenType).
		ExpiresIn(&expiresIn).
		Build(), nil
}

// SignIn authenticates a user and returns tokens
func (a *AuthUser) SignIn(ctx context.Context, req *swagger.SigninRequest) (*swagger.AuthResponse, error) {
	// Validate input
	if err := a.validateSigninRequest(req); err != nil {
		return nil, err
	}

	// Get user with password hash
	user, err := a.authUserPersist.GetUserWithPasswordByEmail(ctx, string(*req.Email))
	if err != nil {
		var appErr *katapp.Err
		if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
			return nil, katapp.NewErr(katapp.ErrUnauthorized, "invalid credentials")
		}
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to get user")
	}

	// Verify password
	if err := a.verifyPassword(user.PasswordHash, *req.Password); err != nil {
		return nil, katapp.NewErr(katapp.ErrUnauthorized, "invalid credentials")
	}

	// Generate tokens with roles
	accessToken, refreshToken, expiresIn, err := a.generateTokensWithRoles(ctx, user.ID)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to generate tokens")
	}

	// Build response
	tokenType := "Bearer"
	return swagger.NewAuthResponseBuilder().
		AccessToken(&accessToken).
		RefreshToken(&refreshToken).
		TokenType(&tokenType).
		ExpiresIn(&expiresIn).
		Build(), nil
}

// SignOut invalidates the refresh token
func (a *AuthUser) SignOut(ctx context.Context, refreshToken string) (*swagger.MessageResponse, error) {
	if refreshToken == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "refresh token is required")
	}

	err := a.invalidateRefreshToken(refreshToken)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to invalidate token")
	}

	message := "Successfully signed out"
	return swagger.NewMessageResponseBuilder().
		Message(&message).
		Build(), nil
}

// RefreshToken generates new tokens using a refresh token
func (a *AuthUser) RefreshToken(ctx context.Context, req *swagger.RefreshRequest) (*swagger.AuthResponse, error) {
	if req.RefreshToken == nil || *req.RefreshToken == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "refresh token is required")
	}

	// Validate refresh token
	userID, err := a.validateRefreshToken(*req.RefreshToken)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrUnauthorized, "invalid or expired refresh token")
	}

	// Generate new tokens with roles
	accessToken, refreshToken, expiresIn, err := a.generateTokensWithRoles(ctx, userID)
	if err != nil {
		return nil, katapp.NewErr(katapp.ErrInternal, "failed to generate tokens")
	}

	// Invalidate old refresh token
	_ = a.invalidateRefreshToken(*req.RefreshToken)

	// Build response
	tokenType := "Bearer"
	return swagger.NewAuthResponseBuilder().
		AccessToken(&accessToken).
		RefreshToken(&refreshToken).
		TokenType(&tokenType).
		ExpiresIn(&expiresIn).
		Build(), nil
}

// ValidateAccessToken validates an access token and returns the user ID
func (a *AuthUser) ValidateAccessToken(token string) (string, error) {
	if token == "" {
		return "", katapp.NewErr(katapp.ErrUnauthorized, "access token is required")
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	userID, err := a.validateAccessToken(token)
	if err != nil {
		return "", katapp.NewErr(katapp.ErrUnauthorized, "invalid or expired access token")
	}

	return userID, nil
}

// validateSignupRequest validates the signup request
func (a *AuthUser) validateSignupRequest(req *swagger.SignupRequest) error {
	if req.Email == nil || *req.Email == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "email is required")
	}
	if req.Password == nil || *req.Password == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "password is required")
	}
	if len(*req.Password) < 8 {
		return katapp.NewErr(katapp.ErrInvalidInput, "password must be at least 8 characters")
	}
	if req.FirstName == nil || *req.FirstName == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "first name is required")
	}
	if req.LastName == nil || *req.LastName == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "last name is required")
	}

	// Basic email validation
	if !strings.Contains(string(*req.Email), "@") {
		return katapp.NewErr(katapp.ErrInvalidInput, "invalid email format")
	}

	return nil
}

// validateSigninRequest validates the signin request
func (a *AuthUser) validateSigninRequest(req *swagger.SigninRequest) error {
	if req.Email == nil || *req.Email == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "email is required")
	}
	if req.Password == nil || *req.Password == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "password is required")
	}

	return nil
}

// JWT helper methods

// generateTokens generates JWT access and refresh tokens
func (a *AuthUser) generateTokens(userID string) (accessToken string, refreshToken string, expiresIn int64, err error) {
	now := time.Now()
	expiresIn = 3600 // 1 hour

	// Generate unique nonces to ensure tokens are always different
	accessNonce, err := a.generateTokenNonce()
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate access token nonce")
	}

	refreshNonce, err := a.generateTokenNonce()
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate refresh token nonce")
	}

	// Generate access token
	accessClaims := jwt.MapClaims{
		"sub":   userID,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(expiresIn) * time.Second).Unix(),
		"type":  "access",
		"nonce": accessNonce,
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate access token")
	}

	// Generate refresh token (valid for 7 days)
	refreshClaims := jwt.MapClaims{
		"sub":   userID,
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

// generateTokensWithRoles generates JWT access and refresh tokens including user roles
func (a *AuthUser) generateTokensWithRoles(ctx context.Context, userID string) (accessToken string, refreshToken string, expiresIn int64, err error) {
	now := time.Now()
	expiresIn = 3600 // 1 hour

	// Get user roles
	roles, err := a.authUserPersist.GetUserRoles(ctx, userID)
	if err != nil {
		katapp.Logger(ctx).Warn("failed to get user roles for token generation", "userID", userID, "error", err)
		// Continue without roles if we can't fetch them
		roles = []string{}
	}

	// Generate unique nonces to ensure tokens are always different
	accessNonce, err := a.generateTokenNonce()
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate access token nonce")
	}

	refreshNonce, err := a.generateTokenNonce()
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate refresh token nonce")
	}

	// Generate access token with roles
	accessClaims := jwt.MapClaims{
		"sub":   userID,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(expiresIn) * time.Second).Unix(),
		"type":  "access",
		"roles": roles,
		"nonce": accessNonce,
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", 0, katapp.NewErr(katapp.ErrInternal, "failed to generate access token")
	}

	// Generate refresh token (valid for 7 days) - no roles needed in refresh token
	refreshClaims := jwt.MapClaims{
		"sub":   userID,
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

// GetJWTSecret returns the JWT secret for token parsing (used by middleware)
func (a *AuthUser) GetJWTSecret() []byte {
	return a.jwtSecret
}

// validateAccessToken validates an access token and returns the user ID
func (a *AuthUser) validateAccessToken(tokenString string) (string, error) {
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

// validateRefreshToken validates a refresh token and returns the user ID
func (a *AuthUser) validateRefreshToken(tokenString string) (string, error) {
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
func (a *AuthUser) invalidateRefreshToken(token string) error {
	// TODO In a production system, you would store invalidated tokens in a blacklist
	// For now, this is a no-op since JWT tokens are stateless
	return nil
}

// generateTokenNonce generates a random nonce for token uniqueness
func (a *AuthUser) generateTokenNonce() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// Password helper methods

// hashPassword hashes a password using bcrypt
func (a *AuthUser) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against its hash using bcrypt
func (a *AuthUser) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

package intgr_test

import (
	"testing"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createAndConfirmUser creates a user and confirms their email, returning the user ID
func createAndConfirmUser(t *testing.T, env *TestEnvironment, email string, password string, firstName string, lastName string) string {
	ctx := env.Context
	appConfig := env.AppConfig

	// Clear mock emails
	clearMockEmails()

	// Create user
	signupReq := &swagger.SignupRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		TenantId:  "default-tenant",
		Source:    "web", // Use web for simplicity in auth tests
	}

	signupResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
		ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
	require.NoError(t, err)

	// Wait for email and get confirmation URL
	err = waitForMockEmail(1, 5)
	require.NoError(t, err)

	lastEmail, err := getLastMockEmail()
	require.NoError(t, err)

	confirmationURL := extractConfirmationURL(lastEmail.Body)
	require.NotEmpty(t, confirmationURL)

	// Confirm email
	_, _, err = kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
		ctx, &appConfig.Server, confirmationURL, nil)
	require.NoError(t, err)

	return signupResp.UserId
}

// validateAuthResponse validates all required fields in AuthResponse
func validateAuthResponse(t *testing.T, authResp *swagger.AuthResponse) {
	assert.NotNil(t, authResp)
	assert.NotEmpty(t, authResp.AccessToken)
	assert.NotEmpty(t, authResp.RefreshToken)
	assert.Equal(t, "Bearer", authResp.TokenType)
	assert.Greater(t, authResp.ExpiresIn, int64(0))
	assert.NotEmpty(t, authResp.UserId)
}

// runAuthenticationTests runs all authentication-related tests
func runAuthenticationTests(t *testing.T, env *TestEnvironment) {
	ctx := env.Context
	appConfig := env.AppConfig

	t.Run("GET /version must succeed", func(t *testing.T) {
		resp, _, err := kathttpc.LocalHttpJsonGetRequest[kathttp.Version](
			ctx, &appConfig.Server, "api/v1/version", nil)
		assert.NoError(t, err)
		service, version, _ := GetAPIServerInfo()
		assert.Equal(t, service, resp.Service)
		assert.Equal(t, true, resp.Healthy)
		assert.Equal(t, version, resp.Version)
	})

	t.Run("POST /auth/signup", func(t *testing.T) {
		signupReq := &swagger.SignupRequest{
			Email:     "test@example.com",
			Password:  "qazwsxedc",
			FirstName: "Test",
			LastName:  "User",
			TenantId:  "default-tenant",
			Source:    "web",
		}
		t.Run("must succeed and return SignupResponse", func(t *testing.T) {
			clearMockEmails()
			signupResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
			assert.NoError(t, err)
			assert.NotNil(t, signupResp)
			assert.Equal(t, "test@example.com", string(signupResp.Email))
			assert.NotEmpty(t, signupResp.UserId)
			assert.Contains(t, signupResp.Message, "check your email")

			// Verify email was sent
			err = waitForMockEmail(1, 5)
			assert.NoError(t, err)
		})
		t.Run("duplicate email with verified user must fail with 409 Conflict", func(t *testing.T) {
			// Create and confirm a user first
			createAndConfirmUser(t, env, "verified-duplicate@example.com", "qazwsxedc", "Verified", "User")

			// Try to signup with same email - should fail
			duplicateReq := &swagger.SignupRequest{
				Email:     "verified-duplicate@example.com",
				Password:  "qazwsxedc",
				FirstName: "Duplicate",
				LastName:  "User",
				TenantId:  "default-tenant",
				Source:    "web",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, duplicateReq)
			kathttpc.AssertStatusConflict(t, err)
		})
		t.Run("invalid email format must fail with 400 Bad Request", func(t *testing.T) {
			invalidEmailReq := &swagger.SignupRequest{
				Email:     "invalid-email",
				Password:  "qazwsxedc",
				FirstName: "Test",
				LastName:  "User",
				TenantId:  "default-tenant",
				Source:    "web",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, invalidEmailReq)
			kathttpc.AssertStatusBadRequest(t, err)
		})
		t.Run("missing required fields must fail with 400 Bad Request", func(t *testing.T) {
			incompleteReq := &swagger.SignupRequest{
				Email:    "incomplete@example.com",
				Password: "qazwsxedc",
				TenantId: "default-tenant",
				Source:   "web",
				// Missing FirstName and LastName
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, incompleteReq)
			kathttpc.AssertStatusBadRequest(t, err)
		})
		t.Run("non-existing tenant must fail with 404 Not Found", func(t *testing.T) {
			nonExistentTenantReq := &swagger.SignupRequest{
				Email:     "tenant-test@example.com",
				Password:  "qazwsxedc",
				FirstName: "Tenant",
				LastName:  "Test",
				TenantId:  "non-existent-tenant",
				Source:    "web",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, nonExistentTenantReq)
			kathttpc.AssertStatusNotFound(t, err)
		})

		t.Run("missing source must fail with 400 Bad Request", func(t *testing.T) {
			noSourceReq := &swagger.SignupRequest{
				Email:     "nosource@example.com",
				Password:  "qazwsxedc",
				FirstName: "No",
				LastName:  "Source",
				TenantId:  "default-tenant",
				// Source field missing
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, noSourceReq)
			kathttpc.AssertStatusBadRequest(t, err)
		})
	})

	t.Run("POST /auth/signin", func(t *testing.T) {
		// First create and confirm a user
		createAndConfirmUser(t, env, "signin@example.com", "qazwsxedc", "Signin", "User")

		signinReq := &swagger.SigninRequest{
			Email:    "signin@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		t.Run("valid credentials must succeed", func(t *testing.T) {
			authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			assert.NoError(t, err)
			validateAuthResponse(t, authResp)
		})
		t.Run("invalid credentials must fail with 401 Unauthorized", func(t *testing.T) {
			invalidReq := &swagger.SigninRequest{
				Email:    "signin@example.com",
				Password: "wrongpassword",
				TenantId: "default-tenant",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, invalidReq)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
		t.Run("non-existent user must fail with 401 Unauthorized", func(t *testing.T) {
			nonExistentReq := &swagger.SigninRequest{
				Email:    "nonexistent@example.com",
				Password: "qazwsxedc",
				TenantId: "default-tenant",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, nonExistentReq)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
		t.Run("non-existing tenant must fail with 401 Unauthorized", func(t *testing.T) {
			nonExistentTenantReq := &swagger.SigninRequest{
				Email:    "signin@example.com",
				Password: "qazwsxedc",
				TenantId: "non-existent-tenant",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, nonExistentTenantReq)
			kathttpc.AssertStatusNotFound(t, err)
		})
	})

	t.Run("POST /auth/refresh", func(t *testing.T) {
		// First create and confirm a user, then sign in to get tokens
		createAndConfirmUser(t, env, "refresh@example.com", "qazwsxedc", "Refresh", "User")

		signinReq := &swagger.SigninRequest{
			Email:    "refresh@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
		assert.NoError(t, err)
		validateAuthResponse(t, authResp)

		refreshReq := &swagger.RefreshRequest{
			RefreshToken: authResp.RefreshToken,
		}
		t.Run("with valid refresh token must succeed", func(t *testing.T) {
			newAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.RefreshRequest, swagger.AuthResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			assert.NoError(t, err)
			validateAuthResponse(t, newAuthResp)
			// New tokens should be different from original
			assert.NotEqual(t, authResp.AccessToken, newAuthResp.AccessToken)
			assert.NotEqual(t, authResp.RefreshToken, newAuthResp.RefreshToken)
			// UserID should remain the same
			assert.Equal(t, authResp.UserId, newAuthResp.UserId)
		})
		t.Run("with invalid refresh token must fail with 401 Unauthorized", func(t *testing.T) {
			invalidRefreshReq := &swagger.RefreshRequest{
				RefreshToken: "invalid-token",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.RefreshRequest, swagger.AuthResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, invalidRefreshReq)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
	})

	t.Run("POST /auth/signout", func(t *testing.T) {
		// First create and confirm a user, then sign in to get tokens
		createAndConfirmUser(t, env, "signout@example.com", "qazwsxedc", "Signout", "User")

		signinReq := &swagger.SigninRequest{
			Email:    "signout@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
		assert.NoError(t, err)
		validateAuthResponse(t, authResp)

		t.Run("must succeed", func(t *testing.T) {
			headers := map[string][]string{
				"Authorization": {"Bearer " + authResp.AccessToken},
			}
			signoutReq := map[string]string{
				"refreshToken": authResp.RefreshToken,
			}
			msgResp, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, any](
				ctx, &appConfig.Server, "api/v1/auth/signout", headers, &signoutReq)
			assert.NoError(t, err)
			assert.NotNil(t, msgResp)
		})
	})

	t.Run("JWT Token Security", func(t *testing.T) {
		t.Run("malformed Authorization header must fail", func(t *testing.T) {
			malformedHeaders := map[string][]string{
				"Authorization": {"InvalidToken"},
			}
			_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
				ctx, &appConfig.Server, "api/v1/users/me", malformedHeaders)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
		t.Run("invalid JWT token must fail", func(t *testing.T) {
			invalidHeaders := map[string][]string{
				"Authorization": {"Bearer invalid.jwt.token"},
			}
			_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
				ctx, &appConfig.Server, "api/v1/users/me", invalidHeaders)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
	})
}

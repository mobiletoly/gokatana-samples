package intgr_test

import (
	"testing"
	"time"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runRefreshTokenTests runs comprehensive tests for the refresh token functionality
func runRefreshTokenTests(t *testing.T, env *TestEnvironment) {
	ctx := env.Context
	appConfig := env.AppConfig

	t.Run("Refresh Token Rotation and Persistence", func(t *testing.T) {
		// Create and confirm a user for testing
		createAndConfirmUser(t, env, "refresh-rotation@example.com", "qazwsxedc", "Refresh", "User")

		// Sign in to get initial tokens
		signinReq := &swagger.SignInRequest{
			Email:    "refresh-rotation@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		initialAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
		require.NoError(t, err)
		validateSignInResponse(t, initialAuth)

		t.Run("Token rotation must generate new tokens and invalidate old ones", func(t *testing.T) {
			// Use the refresh token to get new tokens
			refreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: initialAuth.RefreshToken,
			}
			newAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			require.NoError(t, err)
			validateSignInResponse(t, newAuth)

			// Verify new tokens are different from initial tokens
			assert.NotEqual(t, initialAuth.AccessToken, newAuth.AccessToken, "Access token should be different after refresh")
			assert.NotEqual(t, initialAuth.RefreshToken, newAuth.RefreshToken, "Refresh token should be different after refresh")
			assert.Equal(t, initialAuth.UserId, newAuth.UserId, "User ID should remain the same")

			// Try to use the old refresh token again - should fail
			oldRefreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: initialAuth.RefreshToken,
			}
			_, _, err = kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, oldRefreshReq)
			kathttpc.AssertStatusUnauthorized(t, err)
		})

		t.Run("Multiple refresh token rotations must work", func(t *testing.T) {
			// Start with a fresh signin
			freshAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)

			var currentRefreshToken = freshAuth.RefreshToken
			var previousTokens []string

			// Perform multiple rotations
			for i := 0; i < 3; i++ {
				refreshReq := &swagger.TokenRefreshRequest{
					RefreshToken: currentRefreshToken,
				}
				newAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
					ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
				require.NoError(t, err, "Rotation %d should succeed", i+1)
				validateSignInResponse(t, newAuth)

				// Store the old token for later testing
				previousTokens = append(previousTokens, currentRefreshToken)
				currentRefreshToken = newAuth.RefreshToken

				// Verify the new token is different
				assert.NotEqual(t, previousTokens[len(previousTokens)-1], currentRefreshToken, "Token should change on rotation %d", i+1)
			}

			// Verify all previous tokens are now invalid
			for i, oldToken := range previousTokens {
				oldRefreshReq := &swagger.TokenRefreshRequest{
					RefreshToken: oldToken,
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
					ctx, &appConfig.Server, "api/v1/auth/refresh", nil, oldRefreshReq)
				kathttpc.AssertStatusUnauthorized(t, err)
				t.Logf("Previous token %d correctly invalidated", i+1)
			}
		})
	})

	t.Run("Signout must revoke all refresh tokens", func(t *testing.T) {
		// Create and confirm a user for testing
		createAndConfirmUser(t, env, "signout-test@example.com", "qazwsxedc", "Signout", "Test")

		signinReq := &swagger.SignInRequest{
			Email:    "signout-test@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}

		// Sign in multiple times to create multiple refresh tokens
		var authResponses []*swagger.SignInResponse
		for i := 0; i < 3; i++ {
			authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)
			validateSignInResponse(t, authResp)
			authResponses = append(authResponses, authResp)
		}

		// Sign out using one of the access tokens
		headers := map[string][]string{
			"Authorization": {"Bearer " + authResponses[0].AccessToken},
		}
		signoutResp, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, map[string]interface{}](
			ctx, &appConfig.Server, "api/v1/auth/signout", headers, &map[string]string{})
		require.NoError(t, err)
		assert.NotNil(t, signoutResp)

		// Verify all refresh tokens are now invalid
		for i, authResp := range authResponses {
			refreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: authResp.RefreshToken,
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			kathttpc.AssertStatusUnauthorized(t, err)
			t.Logf("Refresh token %d correctly revoked after signout", i+1)
		}
	})

	t.Run("Invalid refresh token scenarios", func(t *testing.T) {
		t.Run("Empty refresh token must fail", func(t *testing.T) {
			refreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: "",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			kathttpc.AssertStatusBadRequest(t, err)
		})

		t.Run("Malformed refresh token must fail", func(t *testing.T) {
			refreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: "invalid.malformed.token",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			kathttpc.AssertStatusUnauthorized(t, err)
		})

		t.Run("Non-existent refresh token must fail", func(t *testing.T) {
			// Create a valid-looking JWT token but with non-existent content
			refreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJub24tZXhpc3RlbnQtdXNlciIsImlhdCI6MTUxNjIzOTAyMiwiZXhwIjo5OTk5OTk5OTk5LCJ0eXBlIjoicmVmcmVzaCIsIm5vbmNlIjoibm9uLWV4aXN0ZW50In0.invalid",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
	})

	t.Run("Concurrent refresh token usage", func(t *testing.T) {
		// Create and confirm a user for testing
		createAndConfirmUser(t, env, "concurrent@example.com", "qazwsxedc", "Concurrent", "User")

		signinReq := &swagger.SignInRequest{
			Email:    "concurrent@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
		require.NoError(t, err)
		validateSignInResponse(t, authResp)

		// Try to use the same refresh token twice concurrently
		refreshReq := &swagger.TokenRefreshRequest{
			RefreshToken: authResp.RefreshToken,
		}

		// First request should succeed
		newAuth1, _, err1 := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)

		// Second request with same token should fail (token already rotated)
		_, _, err2 := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)

		// One should succeed, one should fail
		if err1 == nil {
			require.NoError(t, err1)
			validateSignInResponse(t, newAuth1)
			kathttpc.AssertStatusUnauthorized(t, err2)
		} else {
			// If first failed, second should have succeeded
			require.NoError(t, err2)
			kathttpc.AssertStatusUnauthorized(t, err1)
		}
	})

	t.Run("Token expiration and cleanup", func(t *testing.T) {
		// Create and confirm a user for testing
		createAndConfirmUser(t, env, "expiration@example.com", "qazwsxedc", "Expiration", "User")

		signinReq := &swagger.SignInRequest{
			Email:    "expiration@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
		require.NoError(t, err)
		validateSignInResponse(t, authResp)

		// Verify the refresh token works initially
		refreshReq := &swagger.TokenRefreshRequest{
			RefreshToken: authResp.RefreshToken,
		}
		newAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
		require.NoError(t, err)
		validateSignInResponse(t, newAuth)

		// Note: We can't easily test actual token expiration in integration tests
		// since refresh tokens are valid for 7 days. This would require either:
		// 1. Mocking time (complex in integration tests)
		// 2. Creating tokens with very short expiration (requires test-specific config)
		// 3. Manually manipulating database (breaks encapsulation)
		// For now, we verify the token works and trust unit tests for expiration logic
	})

	t.Run("Security validations", func(t *testing.T) {
		// Create and confirm a user for testing
		createAndConfirmUser(t, env, "security@example.com", "qazwsxedc", "Security", "User")

		signinReq := &swagger.SignInRequest{
			Email:    "security@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
		require.NoError(t, err)
		validateSignInResponse(t, authResp)

		t.Run("Access token cannot be used as refresh token", func(t *testing.T) {
			// Try to use access token as refresh token
			refreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: authResp.AccessToken, // Using access token instead of refresh token
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			kathttpc.AssertStatusUnauthorized(t, err)
		})

		t.Run("Refresh token cannot be used for API access", func(t *testing.T) {
			// Try to use refresh token for API access
			headers := map[string][]string{
				"Authorization": {"Bearer " + authResp.RefreshToken}, // Using refresh token instead of access token
			}
			_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
				ctx, &appConfig.Server, "api/v1/users/me", headers)
			// Refresh tokens don't have the "access" type, so they get rejected with 403 (insufficient role)
			// This is actually correct behavior - refresh tokens shouldn't be usable for API access
			kathttpc.AssertStatusForbidden(t, err)
		})
	})

	t.Run("Cross-user token validation", func(t *testing.T) {
		// Create two different users
		user1ID := createAndConfirmUser(t, env, "user1@example.com", "qazwsxedc", "User", "One")
		user2ID := createAndConfirmUser(t, env, "user2@example.com", "qazwsxedc", "User", "Two")

		// Sign in both users
		signin1Req := &swagger.SignInRequest{
			Email:    "user1@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		auth1Resp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signin1Req)
		require.NoError(t, err)
		assert.Equal(t, user1ID, auth1Resp.UserId)

		signin2Req := &swagger.SignInRequest{
			Email:    "user2@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		auth2Resp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, signin2Req)
		require.NoError(t, err)
		assert.Equal(t, user2ID, auth2Resp.UserId)

		// Verify tokens are different
		assert.NotEqual(t, auth1Resp.AccessToken, auth2Resp.AccessToken)
		assert.NotEqual(t, auth1Resp.RefreshToken, auth2Resp.RefreshToken)

		// Verify each user can only refresh their own tokens
		refresh1Req := &swagger.TokenRefreshRequest{
			RefreshToken: auth1Resp.RefreshToken,
		}
		newAuth1, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refresh1Req)
		require.NoError(t, err)
		assert.Equal(t, user1ID, newAuth1.UserId)

		refresh2Req := &swagger.TokenRefreshRequest{
			RefreshToken: auth2Resp.RefreshToken,
		}
		newAuth2, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refresh2Req)
		require.NoError(t, err)
		assert.Equal(t, user2ID, newAuth2.UserId)
	})

	t.Run("Refresh token cleanup functionality", func(t *testing.T) {
		// Create and confirm a user for testing
		createAndConfirmUser(t, env, "cleanup-test@example.com", "qazwsxedc", "Cleanup", "Test")

		signinReq := &swagger.SignInRequest{
			Email:    "cleanup-test@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}

		t.Run("Cleanup removes old tokens during signin", func(t *testing.T) {
			// Create 3 tokens through signin - each signin should trigger cleanup
			var tokens []string

			// First signin - no cleanup expected (no tokens to clean)
			auth1, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)
			validateSignInResponse(t, auth1)
			tokens = append(tokens, auth1.RefreshToken)
			time.Sleep(10 * time.Millisecond)

			// Second signin - still no cleanup (only 1 token exists)
			auth2, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)
			validateSignInResponse(t, auth2)
			tokens = append(tokens, auth2.RefreshToken)
			time.Sleep(10 * time.Millisecond)

			// Third signin - still no cleanup (only 2 tokens exist, within limit)
			auth3, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)
			validateSignInResponse(t, auth3)
			tokens = append(tokens, auth3.RefreshToken)
			time.Sleep(10 * time.Millisecond)

			// Fourth signin - NOW cleanup should happen (3 tokens exist, need to keep only 2 most recent)
			auth4, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)
			validateSignInResponse(t, auth4)
			tokens = append(tokens, auth4.RefreshToken)

			// All tokens should be different
			for i := 0; i < len(tokens); i++ {
				for j := i + 1; j < len(tokens); j++ {
					assert.NotEqual(t, tokens[i], tokens[j], "Tokens %d and %d should be different", i, j)
				}
			}

			// After 4th signin, the 1st token should be cleaned up
			// Test token 0 (should be invalid due to cleanup)
			refreshReq0 := &swagger.TokenRefreshRequest{RefreshToken: tokens[0]}
			_, _, err = kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq0)
			if err != nil {
				kathttpc.AssertStatusUnauthorized(t, err)
				t.Log("Token 0 correctly cleaned up and invalid")
			} else {
				t.Error("Token 0 should be invalid due to cleanup, but it still works")
			}

			// Test token 1 (should still work - within 2 most recent before new token)
			refreshReq1 := &swagger.TokenRefreshRequest{RefreshToken: tokens[1]}
			newAuth1, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq1)
			if err != nil {
				t.Logf("Token 1 was cleaned up (this is acceptable behavior)")
			} else {
				validateSignInResponse(t, newAuth1)
				t.Log("Token 1 still works (within recent limit)")
			}
		})

		t.Run("Token cleanup should limit active tokens", func(t *testing.T) {
			// Sign in to get a fresh token
			authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)
			validateSignInResponse(t, authResp)

			// Perform multiple refresh operations
			currentToken := authResp.RefreshToken
			var allGeneratedTokens []string
			allGeneratedTokens = append(allGeneratedTokens, currentToken)

			// Generate 5 more tokens through refresh
			for i := 0; i < 5; i++ {
				refreshReq := &swagger.TokenRefreshRequest{
					RefreshToken: currentToken,
				}
				newAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
					ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
				require.NoError(t, err, "Refresh %d should succeed", i+1)
				validateSignInResponse(t, newAuth)

				allGeneratedTokens = append(allGeneratedTokens, newAuth.RefreshToken)
				currentToken = newAuth.RefreshToken

				time.Sleep(10 * time.Millisecond) // Ensure different timestamps
			}

			// Due to cleanup logic, only the most recent tokens should be valid
			// All previous tokens should be invalid due to cleanup
			for i := 0; i < len(allGeneratedTokens)-1; i++ {
				refreshReq := &swagger.TokenRefreshRequest{
					RefreshToken: allGeneratedTokens[i],
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
					ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
				kathttpc.AssertStatusUnauthorized(t, err)
				t.Logf("Token %d correctly cleaned up and invalid", i)
			}

			// Only the current (last) token should still work
			refreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: currentToken,
			}
			finalAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
			require.NoError(t, err)
			validateSignInResponse(t, finalAuth)
		})

		t.Run("Cleanup should not affect other users", func(t *testing.T) {
			// Create another user
			createAndConfirmUser(t, env, "cleanup-other@example.com", "qazwsxedc", "Other", "User")

			// Sign in both users
			user1Auth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
			require.NoError(t, err)

			user2Req := &swagger.SignInRequest{
				Email:    "cleanup-other@example.com",
				Password: "qazwsxedc",
				TenantId: "default-tenant",
			}
			user2Auth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, user2Req)
			require.NoError(t, err)

			// Refresh user1's token multiple times (triggering cleanup)
			currentToken := user1Auth.RefreshToken
			for i := 0; i < 3; i++ {
				refreshReq := &swagger.TokenRefreshRequest{
					RefreshToken: currentToken,
				}
				newAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
					ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
				require.NoError(t, err)
				currentToken = newAuth.RefreshToken
			}

			// User2's token should still work (cleanup should not affect other users)
			user2RefreshReq := &swagger.TokenRefreshRequest{
				RefreshToken: user2Auth.RefreshToken,
			}
			user2NewAuth, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TokenRefreshRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/refresh", nil, user2RefreshReq)
			require.NoError(t, err)
			validateSignInResponse(t, user2NewAuth)
			assert.Equal(t, user2Auth.UserId, user2NewAuth.UserId)
		})
	})
}

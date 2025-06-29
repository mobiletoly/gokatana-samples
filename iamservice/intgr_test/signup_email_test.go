package intgr_test

import (
	"fmt"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runSignupEmailTests runs tests for signup with email confirmation using mock emails
func runSignupEmailTests(t *testing.T, env *TestEnvironment) {
	ctx := env.Context
	appConfig := env.AppConfig

	t.Run("Signup with Mock Email Confirmation", func(t *testing.T) {
		// Clear any existing mock emails
		clearMockEmails()

		t.Run("Web signup must send confirmation email", func(t *testing.T) {
			signupReq := &swagger.SignupRequest{
				Email:     strfmt.Email("web-test@example.com"),
				Password:  "qazwsxedc",
				FirstName: "Web",
				LastName:  "User",
				TenantID:  "default-tenant",
				Source:    "web",
			}

			// Perform signup
			signupResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
			require.NoError(t, err)
			assert.NotNil(t, signupResp)
			assert.Equal(t, "web-test@example.com", string(signupResp.Email))
			assert.NotEmpty(t, signupResp.UserID)
			assert.Contains(t, signupResp.Message, "check your email")

			// Wait for email to be saved
			err = waitForMockEmail(1, 5)
			require.NoError(t, err)

			// Verify email was saved
			email, err := getLastMockEmail()
			require.NoError(t, err)
			assert.Equal(t, "web-test@example.com", email.To)
			assert.Contains(t, email.Subject, "Confirm Your Email Address")
			assert.Equal(t, "text/html", email.ContentType)
			assert.Contains(t, email.Body, "Web") // User's first name
			assert.Contains(t, email.Body, "IAMService")

			// Extract and validate confirmation URL
			confirmationURL := extractConfirmationURL(email.Body)
			assert.NotEmpty(t, confirmationURL)
			assert.Contains(t, confirmationURL, signupResp.UserID)

			// Test the confirmation URL
			confirmResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
				ctx, &appConfig.Server, confirmationURL, nil)
			require.NoError(t, err)
			assert.Contains(t, confirmResp.Message, "confirmed successfully")
		})

		t.Run("Android signup must send confirmation email with 6-digit code", func(t *testing.T) {
			clearMockEmails()

			signupReq := &swagger.SignupRequest{
				Email:     strfmt.Email("android-test@example.com"),
				Password:  "qazwsxedc",
				FirstName: "Android",
				LastName:  "User",
				TenantID:  "default-tenant",
				Source:    "android",
			}

			// Perform signup
			signupResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
			require.NoError(t, err)

			// Wait for email to be saved
			err = waitForMockEmail(1, 5)
			require.NoError(t, err)

			// Verify email was saved
			email, err := getLastMockEmail()
			require.NoError(t, err)
			assert.Equal(t, "android-test@example.com", email.To)
			assert.Contains(t, email.Subject, "Your Confirmation Code")
			assert.Contains(t, email.Subject, "Android")
			assert.Contains(t, email.Body, "Android") // User's first name
			assert.Contains(t, email.Body, "android") // Platform

			// Extract and validate 6-digit code
			confirmationCode := extractSixDigitCode(email.Body)
			assert.Len(t, confirmationCode, 6)
			assert.Regexp(t, `^\d{6}$`, confirmationCode)

			// Test the confirmation with code
			confirmURL := fmt.Sprintf("api/v1/auth/confirm-email?userId=%s&code=%s", signupResp.UserID, confirmationCode)
			confirmResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
				ctx, &appConfig.Server, confirmURL, nil)
			require.NoError(t, err)
			assert.Contains(t, confirmResp.Message, "confirmed successfully")
		})

		t.Run("iOS signup must send confirmation email with 6-digit code", func(t *testing.T) {
			clearMockEmails()

			signupReq := &swagger.SignupRequest{
				Email:     strfmt.Email("ios-test@example.com"),
				Password:  "qazwsxedc",
				FirstName: "iOS",
				LastName:  "User",
				TenantID:  "default-tenant",
				Source:    "ios",
			}

			// Perform signup
			signupResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
			require.NoError(t, err)

			// Wait for email to be saved
			err = waitForMockEmail(1, 5)
			require.NoError(t, err)

			// Verify email was saved
			email, err := getLastMockEmail()
			require.NoError(t, err)
			assert.True(t, validateMobileEmailContent(email, "ios-test@example.com", "iOS", "iOS"))

			// Extract and test confirmation code
			confirmationCode := extractSixDigitCode(email.Body)
			confirmURL := fmt.Sprintf("api/v1/auth/confirm-email?userId=%s&code=%s", signupResp.UserID, confirmationCode)
			confirmResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
				ctx, &appConfig.Server, confirmURL, nil)
			require.NoError(t, err)
			assert.Contains(t, confirmResp.Message, "confirmed successfully")
		})

		t.Run("Re-signup with unverified email must work", func(t *testing.T) {
			clearMockEmails()
			email := "resignup-test@example.com"

			// First signup
			signupReq1 := &swagger.SignupRequest{
				Email:     strfmt.Email(email),
				Password:  "password1",
				FirstName: "First",
				LastName:  "Attempt",
				TenantID:  "default-tenant",
				Source:    "web",
			}

			signupResp1, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq1)
			require.NoError(t, err)

			// Wait for first email
			err = waitForMockEmail(1, 5)
			require.NoError(t, err)

			// Second signup (should replace first)
			signupReq2 := &swagger.SignupRequest{
				Email:     strfmt.Email(email),
				Password:  "password2",
				FirstName: "Second",
				LastName:  "Attempt",
				TenantID:  "default-tenant",
				Source:    "android",
			}

			signupResp2, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq2)
			require.NoError(t, err)

			// Should have different user ID
			assert.NotEqual(t, signupResp1.UserID, signupResp2.UserID)

			// Wait for second email
			err = waitForMockEmail(2, 5)
			require.NoError(t, err)

			// Get all emails for this address
			emails, err := getMockEmailsTo(email)
			require.NoError(t, err)
			assert.Len(t, emails, 2)

			// First email should be for web
			assert.Contains(t, emails[0].Subject, "Confirm Your Email Address")
			// Second email should be for android
			assert.Contains(t, emails[1].Subject, "Your Confirmation Code")
			assert.Contains(t, emails[1].Subject, "Android")

			// Only the second confirmation should work
			secondCode := extractSixDigitCode(emails[1].Body)
			confirmURL := fmt.Sprintf("api/v1/auth/confirm-email?userId=%s&code=%s", signupResp2.UserID, secondCode)
			confirmResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
				ctx, &appConfig.Server, confirmURL, nil)
			require.NoError(t, err)
			assert.Contains(t, confirmResp.Message, "confirmed successfully")
		})

		t.Run("Invalid confirmation scenarios", func(t *testing.T) {
			clearMockEmails()

			// Create a user first
			signupReq := &swagger.SignupRequest{
				Email:     strfmt.Email("invalid-test@example.com"),
				Password:  "qazwsxedc",
				FirstName: "Invalid",
				LastName:  "Test",
				TenantID:  "default-tenant",
				Source:    "android",
			}

			signupResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.SignupResponse](
				ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
			require.NoError(t, err)

			t.Run("Invalid user ID must fail", func(t *testing.T) {
				confirmURL := "api/v1/auth/confirm-email?userId=invalid-user-id&code=123456"
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
					ctx, &appConfig.Server, confirmURL, nil)
				kathttpc.AssertStatusNotFound(t, err)
			})

			t.Run("Invalid confirmation code must fail", func(t *testing.T) {
				confirmURL := fmt.Sprintf("api/v1/auth/confirm-email?userId=%s&code=999999", signupResp.UserID)
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
					ctx, &appConfig.Server, confirmURL, nil)
				kathttpc.AssertStatusNotFound(t, err)
			})

			t.Run("Missing parameters must fail", func(t *testing.T) {
				testCases := []string{
					"api/v1/auth/confirm-email",
					"api/v1/auth/confirm-email?userId=test-id",
					"api/v1/auth/confirm-email?code=123456",
				}

				for _, confirmURL := range testCases {
					_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.EmailConfirmationResponse](
						ctx, &appConfig.Server, confirmURL, nil)
					kathttpc.AssertStatusBadRequest(t, err)
				}
			})
		})
	})
}

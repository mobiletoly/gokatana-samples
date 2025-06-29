package intgr_test

import (
	"testing"

	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/stretchr/testify/assert"
)

// TestAPIRoutes is the single entry point for all integration tests
// This prevents parallel execution issues with TestContainers
func TestAPIRoutes(t *testing.T) {
	env := SetupTestEnvironment(t)
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

	// Run authentication tests
	t.Run("Authentication API", func(t *testing.T) {
		runAuthenticationTests(t, env)
	})

	// Run signup and email confirmation tests with mock emails
	t.Run("Signup with Email Confirmation", func(t *testing.T) {
		runSignupEmailTests(t, env)
	})

	// Run user management tests
	t.Run("User Management API", func(t *testing.T) {
		runUserManagementTests(t, env)
	})

	// Run tenant management tests
	t.Run("Tenant Management API", func(t *testing.T) {
		runTenantManagementTests(t, env)
	})
}

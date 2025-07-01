package intgr_test

import (
	"testing"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/stretchr/testify/assert"
)

// runTenantManagementTests runs all tenant management-related tests
func runTenantManagementTests(t *testing.T, env *TestEnvironment) {
	ctx := env.Context
	appConfig := env.AppConfig

	t.Run("Tenant Management API (Sysadmin-only routes)", func(t *testing.T) {
		// Create sysadmin user using sample data
		sysadminSigninReq := &swagger.SigninRequest{
			Email:    "john.doe.sysadmin@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		sysadminAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, sysadminSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, sysadminAuthResp)
		assert.NotNil(t, sysadminAuthResp.AccessToken)
		sysadminHeaders := map[string][]string{
			"Authorization": {"Bearer " + sysadminAuthResp.AccessToken},
		}

		// Create admin user using sample data (should not have access to tenant management)
		adminSigninReq := &swagger.SigninRequest{
			Email:    "testadmin@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		adminAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, adminSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, adminAuthResp)
		assert.NotNil(t, adminAuthResp.AccessToken)
		adminHeaders := map[string][]string{
			"Authorization": {"Bearer " + adminAuthResp.AccessToken},
		}

		// Create regular user using sample data (should not have access to tenant management)
		userSigninReq := &swagger.SigninRequest{
			Email:    "testuser@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		userAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, userSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, userAuthResp)
		assert.NotNil(t, userAuthResp.AccessToken)
		userHeaders := map[string][]string{
			"Authorization": {"Bearer " + userAuthResp.AccessToken},
		}

		t.Run("GET /tenants", func(t *testing.T) {
			t.Run("sysadmin user must succeed", func(t *testing.T) {
				tenantListResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantListResponse](
					ctx, &appConfig.Server, "api/v1/tenants", sysadminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, tenantListResp)
				assert.NotNil(t, tenantListResp.Tenants)
				assert.Greater(t, len(tenantListResp.Tenants), 0)
				// Should contain at least default-tenant and test-tenant
				assert.GreaterOrEqual(t, len(tenantListResp.Tenants), 2)
			})
			t.Run("admin user must return its own single tenant", func(t *testing.T) {
				tenantListResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantListResponse](
					ctx, &appConfig.Server, "api/v1/tenants", adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, tenantListResp)
				assert.NotNil(t, tenantListResp.Tenants)
				assert.Equal(t, 1, len(tenantListResp.Tenants))
				assert.Equal(t, "default-tenant", tenantListResp.Tenants[0].Id)
			})
			t.Run("regular user must return its own single tenant", func(t *testing.T) {
				tenantListResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantListResponse](
					ctx, &appConfig.Server, "api/v1/tenants", userHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, tenantListResp)
				assert.NotNil(t, tenantListResp.Tenants)
				assert.Equal(t, 1, len(tenantListResp.Tenants))
				assert.Equal(t, "default-tenant", tenantListResp.Tenants[0].Id)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantListResponse](
					ctx, &appConfig.Server, "api/v1/tenants", nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("POST /tenants", func(t *testing.T) {
			t.Run("sysadmin user must succeed in creating tenant", func(t *testing.T) {
				createTenantReq := &swagger.TenantCreateRequest{
					Id:          "integration-test-tenant",
					Name:        "Integration Test Tenant",
					Description: "Tenant created during integration tests",
				}
				tenantResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants", sysadminHeaders, createTenantReq)
				assert.NoError(t, err)
				assert.NotNil(t, tenantResp)
				assert.Equal(t, "integration-test-tenant", tenantResp.Id)
				assert.Equal(t, "Integration Test Tenant", tenantResp.Name)
				assert.Equal(t, "Tenant created during integration tests", tenantResp.Description)
				assert.NotNil(t, tenantResp.CreatedAt)
				assert.NotNil(t, tenantResp.UpdatedAt)
			})
			t.Run("sysadmin user must fail with duplicate tenant ID", func(t *testing.T) {
				duplicateTenantReq := &swagger.TenantCreateRequest{
					Id:          "integration-test-tenant", // Same ID as above
					Name:        "Duplicate Tenant",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants", sysadminHeaders, duplicateTenantReq)
				kathttpc.AssertStatusConflict(t, err)
			})
			t.Run("sysadmin user must fail with invalid tenant ID", func(t *testing.T) {
				invalidTenantReq := &swagger.TenantCreateRequest{
					Id:          "ab", // Too short (less than 3 characters)
					Name:        "Invalid Tenant",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants", sysadminHeaders, invalidTenantReq)
				kathttpc.AssertStatusBadRequest(t, err)
			})
			t.Run("sysadmin user must fail with missing required fields", func(t *testing.T) {
				incompleteTenantReq := &swagger.TenantCreateRequest{
					Id: "incomplete-tenant",
					// Missing Name field
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants", sysadminHeaders, incompleteTenantReq)
				kathttpc.AssertStatusBadRequest(t, err)
			})
			t.Run("admin user must fail with 403 Forbidden", func(t *testing.T) {
				createTenantReq := &swagger.TenantCreateRequest{
					Id:          "admin-test-tenant",
					Name:        "Admin Test Tenant",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants", adminHeaders, createTenantReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				createTenantReq := &swagger.TenantCreateRequest{
					Id:          "user-test-tenant",
					Name:        "User Test Tenant",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants", userHeaders, createTenantReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				createTenantReq := &swagger.TenantCreateRequest{
					Id:          "unauth-test-tenant",
					Name:        "Unauth Test Tenant",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants", nil, createTenantReq)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("GET /tenants/{tenantId}", func(t *testing.T) {
			tenantID := "default-tenant"
			t.Run("sysadmin user must succeed", func(t *testing.T) {
				tenantResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, sysadminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, tenantResp)
				assert.Equal(t, tenantID, tenantResp.Id)
				assert.Equal(t, "Default Tenant", tenantResp.Name)
				assert.NotNil(t, tenantResp.CreatedAt)
				assert.NotNil(t, tenantResp.UpdatedAt)
			})
			t.Run("sysadmin user must fail with non-existent tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/non-existent-tenant", sysadminHeaders)
				kathttpc.AssertStatusNotFound(t, err)
			})
			t.Run("admin user must succeed for own tenant", func(t *testing.T) {
				tenantResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, tenantResp)
				assert.Equal(t, "default-tenant", tenantResp.Id)
				assert.Equal(t, "Default Tenant", tenantResp.Name)
				assert.NotNil(t, tenantResp.CreatedAt)
				assert.NotNil(t, tenantResp.UpdatedAt)
			})
			t.Run("admin user must fail with 403 Forbidden for other tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/test-tenant", adminHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must succeed for own tenant", func(t *testing.T) {
				tenantResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, userHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, tenantResp)
				assert.Equal(t, "default-tenant", tenantResp.Id)
				assert.Equal(t, "Default Tenant", tenantResp.Name)
				assert.NotNil(t, tenantResp.CreatedAt)
				assert.NotNil(t, tenantResp.UpdatedAt)
			})
			t.Run("regular user must fail with 403 Forbidden for other tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/test-tenant", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("PUT /tenants/{tenantId}", func(t *testing.T) {
			tenantID := "integration-test-tenant" // Created in POST test above
			t.Run("sysadmin user must succeed in updating tenant", func(t *testing.T) {
				updateTenantReq := &swagger.TenantUpdateRequest{
					Name:        "Updated Integration Test Tenant",
					Description: "Updated description for integration test tenant",
				}
				tenantResp, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.TenantUpdateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, sysadminHeaders, updateTenantReq)
				assert.NoError(t, err)
				assert.NotNil(t, tenantResp)
				assert.Equal(t, tenantID, tenantResp.Id)
				assert.Equal(t, "Updated Integration Test Tenant", tenantResp.Name)
				assert.Equal(t, "Updated description for integration test tenant", tenantResp.Description)
				assert.NotNil(t, tenantResp.UpdatedAt)
			})
			t.Run("sysadmin user must fail with non-existent tenant", func(t *testing.T) {
				updateTenantReq := &swagger.TenantUpdateRequest{
					Name:        "Non-existent Tenant",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.TenantUpdateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/non-existent-tenant", sysadminHeaders, updateTenantReq)
				kathttpc.AssertStatusNotFound(t, err)
			})
			t.Run("sysadmin user must fail with missing required fields", func(t *testing.T) {
				incompleteUpdateReq := &swagger.TenantUpdateRequest{
					// Missing Name field
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.TenantUpdateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, sysadminHeaders, incompleteUpdateReq)
				kathttpc.AssertStatusBadRequest(t, err)
			})
			t.Run("admin user must fail with 403 Forbidden", func(t *testing.T) {
				updateTenantReq := &swagger.TenantUpdateRequest{
					Name:        "Admin Update",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.TenantUpdateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, adminHeaders, updateTenantReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				updateTenantReq := &swagger.TenantUpdateRequest{
					Name:        "User Update",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.TenantUpdateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, userHeaders, updateTenantReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				updateTenantReq := &swagger.TenantUpdateRequest{
					Name:        "Unauth Update",
					Description: "This should fail",
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.TenantUpdateRequest, swagger.TenantResponse](
					ctx, &appConfig.Server, "api/v1/tenants/"+tenantID, nil, updateTenantReq)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("DELETE /tenants/{tenantId}", func(t *testing.T) {
			// First create a tenant specifically for deletion testing
			createTenantReq := &swagger.TenantCreateRequest{
				Id:          "delete-test-tenant",
				Name:        "Delete Test Tenant",
				Description: "Tenant created for deletion testing",
			}
			_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.TenantCreateRequest, swagger.TenantResponse](
				ctx, &appConfig.Server, "api/v1/tenants", sysadminHeaders, createTenantReq)
			assert.NoError(t, err)

			t.Run("sysadmin user must succeed in deleting tenant", func(t *testing.T) {
				msgResp, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/tenants/delete-test-tenant", sysadminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, msgResp)
			})
			t.Run("sysadmin user must fail with non-existent tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/tenants/non-existent-tenant", sysadminHeaders)
				kathttpc.AssertStatusNotFound(t, err)
			})
			t.Run("sysadmin user must fail when deleting default tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/tenants/default-tenant", sysadminHeaders)
				kathttpc.AssertStatusBadRequest(t, err)
			})
			t.Run("sysadmin user must fail when deleting tenant with existing users", func(t *testing.T) {
				// test-tenant has existing users, so deletion should fail
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/tenants/test-tenant", sysadminHeaders)
				kathttpc.AssertStatusBadRequest(t, err)
			})
			t.Run("admin user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/tenants/integration-test-tenant", adminHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/tenants/integration-test-tenant", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/tenants/integration-test-tenant", nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})
	})
}

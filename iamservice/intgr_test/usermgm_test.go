package intgr_test

import (
	"github.com/oapi-codegen/runtime/types"
	"github.com/samber/lo"
	"time"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/stretchr/testify/assert"
	"testing"
)

// runUserManagementTests runs all user management-related tests
func runUserManagementTests(t *testing.T, env *TestEnvironment) {
	ctx := env.Context
	appConfig := env.AppConfig

	t.Run("User Management API (Admin-only routes)", func(t *testing.T) {
		sysadminSigninReq := &swagger.SignInRequest{
			Email:    "john.doe.sysadmin@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		sysadminAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, sysadminSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, sysadminAuthResp)
		assert.NotNil(t, sysadminAuthResp.AccessToken)
		sysadminHeaders := map[string][]string{
			"Authorization": {"Bearer " + sysadminAuthResp.AccessToken},
		}

		// Create admin user using sample data
		adminSigninReq := &swagger.SignInRequest{
			Email:    "testadmin@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		adminAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, adminSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, adminAuthResp)
		assert.NotNil(t, adminAuthResp.AccessToken)
		adminHeaders := map[string][]string{
			"Authorization": {"Bearer " + adminAuthResp.AccessToken},
		}

		// Create regular user using sample data
		userSigninReq := &swagger.SignInRequest{
			Email:    "testuser@example.com",
			Password: "qazwsxedc",
			TenantId: "default-tenant",
		}
		userAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, userSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, userAuthResp)
		assert.NotNil(t, userAuthResp.AccessToken)
		userHeaders := map[string][]string{
			"Authorization": {"Bearer " + userAuthResp.AccessToken},
		}

		t.Run("GET /users/all", func(t *testing.T) {
			t.Run("sysadmin user must succeed", func(t *testing.T) {
				userListResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUsersResponse](
					ctx, &appConfig.Server, "api/v1/users/all?limit=100", sysadminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userListResp)
				assert.NotNil(t, userListResp.Items)
				// loop through users and check that users are from at least 2 different tenants
				tenants := make(map[string]bool)
				for _, user := range userListResp.Items {
					tenants[user.TenantId] = true
				}
				assert.GreaterOrEqual(t, len(tenants), 2)
			})
			t.Run("admin user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUsersResponse](
					ctx, &appConfig.Server, "api/v1/users/all", adminHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUsersResponse](
					ctx, &appConfig.Server, "api/v1/users/all", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUsersResponse](
					ctx, &appConfig.Server, "api/v1/users/all", nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("GET /users", func(t *testing.T) {
			t.Run("admin user must succeed", func(t *testing.T) {
				userListResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUsersResponse](
					ctx, &appConfig.Server, "api/v1/users", adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userListResp)
				assert.NotNil(t, userListResp.Items)
				assert.Greater(t, len(userListResp.Items), 0)
			})
			t.Run("regular user must return this user only", func(t *testing.T) {
				userListResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUsersResponse](
					ctx, &appConfig.Server, "api/v1/users", userHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userListResp)
				assert.NotNil(t, userListResp.Items)
				assert.Equal(t, 1, len(userListResp.Items))
				assert.Equal(t, "testuser@example.com", string(userListResp.Items[0].Email))
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUsersResponse](
					ctx, &appConfig.Server, "api/v1/users", nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("GET /users/{userId}", func(t *testing.T) {
			userID := "test-user-5"
			t.Run("admin user must succeed", func(t *testing.T) {
				authUserResponse, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID, adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, authUserResponse)
				assert.Equal(t, userID, authUserResponse.Id)
			})
			t.Run("admin user must fail with 403 Forbidden for user from other tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/test-user-1", adminHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must be able to fetch itself", func(t *testing.T) {
				authUserResponse, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID, userHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, authUserResponse)
				assert.Equal(t, userID, authUserResponse.Id)
			})
			t.Run("regular user must fail with 403 Forbidden for non-itself user", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/default-user-2", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
		})

		t.Run("User Profile API (Authenticated routes)", func(t *testing.T) {
			// Create regular user using sample data
			userSigninReq := &swagger.SignInRequest{
				Email:    "testuser@example.com",
				Password: "qazwsxedc",
				TenantId: "default-tenant",
			}
			userAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignInRequest, swagger.SignInResponse](
				ctx, &appConfig.Server, "api/v1/auth/signin", nil, userSigninReq)
			assert.NoError(t, err)
			assert.NotNil(t, userAuthResp)
			assert.NotNil(t, userAuthResp.AccessToken)
			userHeaders := map[string][]string{
				"Authorization": {"Bearer " + userAuthResp.AccessToken},
			}

			t.Run("GET /users/me", func(t *testing.T) {
				t.Run("authenticated user must succeed", func(t *testing.T) {
					authUserResponse, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
						ctx, &appConfig.Server, "api/v1/users/me", userHeaders)
					assert.NoError(t, err)
					assert.NotNil(t, authUserResponse)
					assert.Equal(t, "testuser@example.com", string(authUserResponse.Email))
					assert.Equal(t, "Test", authUserResponse.FirstName)
					assert.Equal(t, "User", authUserResponse.LastName)
				})
				t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
					_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.AuthUserResponse](
						ctx, &appConfig.Server, "api/v1/users/me", nil)
					kathttpc.AssertStatusUnauthorized(t, err)
				})
			})
		})

		t.Run("PUT /users/{userId}", func(t *testing.T) {
			userID := "test-user-5"
			t.Run("admin user must succeed", func(t *testing.T) {
				// Create update request
				updateReq := swagger.UpdateAuthUserRequest{
					FirstName: "UpdatedFirst",
					LastName:  "UpdatedLast",
				}

				// Update user details
				updatedUser, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateAuthUserRequest, swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID, adminHeaders, &updateReq)
				assert.NoError(t, err)
				assert.NotNil(t, updatedUser)
				assert.Equal(t, userID, updatedUser.Id)
				assert.Equal(t, "UpdatedFirst", updatedUser.FirstName)
				assert.Equal(t, "UpdatedLast", updatedUser.LastName)
			})
			t.Run("admin user must fail with 403 Forbidden for user from other tenant", func(t *testing.T) {
				updateReq := swagger.UpdateAuthUserRequest{
					FirstName: "UpdatedFirst",
					LastName:  "UpdatedLast",
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateAuthUserRequest, swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/test-user-1", adminHeaders, &updateReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must be able to update itself", func(t *testing.T) {
				updateReq := swagger.UpdateAuthUserRequest{
					FirstName: "SelfUpdatedFirst",
					LastName:  "SelfUpdatedLast",
				}
				// Use test-user-5 which corresponds to the userHeaders (testuser@example.com)
				updatedUser, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateAuthUserRequest, swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/test-user-5", userHeaders, &updateReq)
				assert.NoError(t, err)
				assert.NotNil(t, updatedUser)
				assert.Equal(t, "test-user-5", updatedUser.Id)
				assert.Equal(t, "SelfUpdatedFirst", updatedUser.FirstName)
				assert.Equal(t, "SelfUpdatedLast", updatedUser.LastName)
			})
			t.Run("regular user must fail with 403 Forbidden for other user", func(t *testing.T) {
				updateReq := swagger.UpdateAuthUserRequest{
					FirstName: "UpdatedFirst",
					LastName:  "UpdatedLast",
				}
				// Try to update a different user (default-user-1) - should fail
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateAuthUserRequest, swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/default-user-1", userHeaders, &updateReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("invalid request body must fail with 400 Bad Request", func(t *testing.T) {
				updateReq := swagger.UpdateAuthUserRequest{
					FirstName: "", // Empty first name should fail
					LastName:  "UpdatedLast",
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateAuthUserRequest, swagger.AuthUserResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID, adminHeaders, &updateReq)
				kathttpc.AssertStatusBadRequest(t, err)
			})
		})

		t.Run("GET /users/{userId}/profile", func(t *testing.T) {
			userID := "test-user-5"
			t.Run("admin user must succeed", func(t *testing.T) {
				userProfile, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/profile", adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userProfile)
				assert.Equal(t, userID, userProfile.UserId)
				assert.NotNil(t, userProfile.CreatedAt)
				assert.NotNil(t, userProfile.UpdatedAt)
				assert.Equal(t, 175, *userProfile.Height)
				assert.Equal(t, 70, *userProfile.Weight)
				assert.Equal(t, "male", string(*userProfile.Gender))
				assert.Equal(t, "1990-01-15", userProfile.BirthDate.String())
			})
			t.Run("admin user must fail with 403 Forbidden for user from other tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/test-user-1/profile", adminHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must succeed for itself", func(t *testing.T) {
				userProfile, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userAuthResp.UserId+"/profile", userHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userProfile)
				assert.Equal(t, userAuthResp.UserId, userProfile.UserId)
				assert.NotNil(t, userProfile.CreatedAt)
				assert.NotNil(t, userProfile.UpdatedAt)
			})
			t.Run("regular user must fail with 403 Forbidden when fetch other user's profile", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/default-user-2/profile", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/profile", nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("PUT /users/{userId}/profile", func(t *testing.T) {
			userID := "test-user-5"
			t.Run("admin user must succeed", func(t *testing.T) {
				// Create update request
				updateReq := swagger.UpdateUserProfileRequest{
					Height:    lo.ToPtr(180),
					Weight:    lo.ToPtr(75),
					Gender:    lo.ToPtr(swagger.UserProfileGender("male")),
					BirthDate: &types.Date{Time: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)},
				}

				updatedProfile, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateUserProfileRequest, swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/profile", adminHeaders, &updateReq)
				assert.NoError(t, err)
				assert.NotNil(t, updatedProfile)
				assert.Equal(t, userID, updatedProfile.UserId)
				assert.Equal(t, 180, *updatedProfile.Height)
				assert.Equal(t, 75, *updatedProfile.Weight)
				assert.Equal(t, "male", string(*updatedProfile.Gender))
				assert.NotNil(t, updatedProfile.BirthDate)
				assert.NotNil(t, updatedProfile.UpdatedAt)
			})
			t.Run("admin user must fail with 403 Forbidden for user from other tenant", func(t *testing.T) {
				updateReq := swagger.UpdateUserProfileRequest{
					Height: lo.ToPtr(170),
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateUserProfileRequest, swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/test-user-1/profile", adminHeaders, &updateReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must succeed for own profile", func(t *testing.T) {
				updateReq := swagger.UpdateUserProfileRequest{
					Height: lo.ToPtr(170),
					Weight: lo.ToPtr(65),
					Gender: lo.ToPtr(swagger.UserProfileGender("female")),
				}
				updatedProfile, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateUserProfileRequest, swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userAuthResp.UserId+"/profile", userHeaders, &updateReq)
				assert.NoError(t, err)
				assert.NotNil(t, updatedProfile)
				assert.Equal(t, userAuthResp.UserId, updatedProfile.UserId)
				assert.Equal(t, 170, *updatedProfile.Height)
				assert.Equal(t, 65, *updatedProfile.Weight)
				assert.Equal(t, "female", string(*updatedProfile.Gender))
			})
			t.Run("regular user must fail with 403 Forbidden for other user's profile", func(t *testing.T) {
				updateReq := swagger.UpdateUserProfileRequest{
					Height: lo.ToPtr(170),
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateUserProfileRequest, swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/default-user-1/profile", userHeaders, &updateReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				updateReq := swagger.UpdateUserProfileRequest{
					Height: lo.ToPtr(170),
				}
				_, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateUserProfileRequest, swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/profile", nil, &updateReq)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
			t.Run("update with null values must succeed", func(t *testing.T) {
				// Clear gender field
				updateReq := swagger.UpdateUserProfileRequest{
					Height:    lo.ToPtr(175),
					Weight:    lo.ToPtr(70),
					Gender:    nil, // This should clear the gender field
					BirthDate: nil, // This should clear the birth date field
				}

				updatedProfile, _, err := kathttpc.LocalHttpJsonPutRequest[swagger.UpdateUserProfileRequest, swagger.UserProfileResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/profile", adminHeaders, &updateReq)
				assert.NoError(t, err)
				assert.NotNil(t, updatedProfile)
				assert.Equal(t, userID, updatedProfile.UserId)
				assert.Equal(t, 175, *updatedProfile.Height)
				assert.Equal(t, 70, *updatedProfile.Weight)
				assert.Nil(t, updatedProfile.Gender)
				assert.Nil(t, updatedProfile.BirthDate)
			})
		})

		t.Run("GET /users/{userId}/roles", func(t *testing.T) {
			userID := "test-admin-5"
			t.Run("sysadmin user must succeed", func(t *testing.T) {
				rolesResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserRolesResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", sysadminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, rolesResp)
				assert.Contains(t, rolesResp.Roles, "admin")
				assert.Contains(t, rolesResp.Roles, "user")
			})
			t.Run("admin user must succeed", func(t *testing.T) {
				rolesResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserRolesResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, rolesResp)
				assert.Contains(t, rolesResp.Roles, "admin")
				assert.Contains(t, rolesResp.Roles, "user")
			})
			t.Run("admin user must fail with 403 Forbidden for user from other tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserRolesResponse](
					ctx, &appConfig.Server, "api/v1/users/test-user-1/roles", adminHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must succeed for itself", func(t *testing.T) {
				rolesResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserRolesResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userAuthResp.UserId+"/roles", userHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, rolesResp)
				assert.Contains(t, rolesResp.Roles, "user")
			})
			t.Run("regular user must fail with 403 Forbidden for non-itself user", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserRolesResponse](
					ctx, &appConfig.Server, "api/v1/users/default-user-2/roles", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
		})

		t.Run("POST /users/{userId}/roles", func(t *testing.T) {
			userID := "test-user-5"
			t.Run("admin user must succeed in assigning role within own tenant", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "admin",
				}
				msgResp, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", adminHeaders, &assignRoleReq)
				assert.NoError(t, err)
				assert.NotNil(t, msgResp)
			})
			t.Run("admin user must fail with 403 Forbidden for user from other tenant", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "admin",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, any](
					ctx, &appConfig.Server, "api/v1/users/test-user-1/roles", adminHeaders, &assignRoleReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("admin user must fail with 403 Forbidden when assign sysadmin role", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "sysadmin",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", adminHeaders, &assignRoleReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "user",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", userHeaders, &assignRoleReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "admin",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", nil, &assignRoleReq)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("DELETE /users/{userId}/roles/{roleName}", func(t *testing.T) {
			userID := "test-user-5"
			roleName := "user"
			t.Run("sysadmin user must succeed in removing role", func(t *testing.T) {
				msgResp, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles/"+roleName, sysadminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, msgResp)
			})
			t.Run("admin user must succeed in removing role within own tenant", func(t *testing.T) {
				msgResp, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles/"+roleName, adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, msgResp)
			})
			t.Run("admin user must fail with 403 Forbidden for user from other tenant", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/users/test-user-1/roles/"+roleName, adminHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles/"+roleName, userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[any](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles/"+roleName, nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})
	})
}

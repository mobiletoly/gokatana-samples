package intgr_test

import (
	// Pick one of the imports below, depending on which framework you use
	// -- Option 1: Echo framework
	apiserver "github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver"
	// -- Option 2: Chi router
	// apiserver "github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver_chi"
	// -- Option 3: Standard net/http
	//apiserver "github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver_std"

	"github.com/go-openapi/strfmt"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/infra"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/mobiletoly/gokatana/katpg"
	"log/slog"
	"os"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAPIRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := katapp.ContextWithAppLogger(logger)
	ctx = katapp.ContextWithRunInTest(ctx, true)

	dbMigrate := "../dbmigrate"
	pc := katpg.RunPostgresTestContainer(ctx, t, &dbMigrate, []string{
		"init/sample_data.sql",
	})
	t.Cleanup(func() {
		pc.Terminate(ctx, t)
	})

	started := make(chan struct{})
	var appConfig *app.Config
	go func() {
		infra.Start("test",
			func(cfg *app.Config) {
				pc.ApplyToConfig(&cfg.Database)
				appConfig = cfg
			},
			started,
		)
	}()
	<-started
	kathttpc.WaitForURLToBecomeReady(ctx, kathttpc.LocalURL(appConfig.Server.Port, "api/v1/version"))

	t.Run("API Routes", func(t *testing.T) {
		t.Run("GET /version must succeed", func(t *testing.T) {
			resp, _, err := kathttpc.LocalHttpJsonGetRequest[kathttp.Version](
				ctx, &appConfig.Server, "api/v1/version", nil)
			assert.NoError(t, err)
			assert.Equal(t, apiserver.HttpVersionResponse.Service, resp.Service)
			assert.Equal(t, true, resp.Healthy)
			assert.Equal(t, apiserver.AppTagVersion, resp.Version)
		})
		t.Run("Authentication API", func(t *testing.T) {
			t.Run("POST /auth/signup", func(t *testing.T) {
				signupReq := &swagger.SignupRequest{
					Email:     (*strfmt.Email)(lo.ToPtr("test@example.com")),
					Password:  lo.ToPtr("password123"),
					FirstName: lo.ToPtr("Test"),
					LastName:  lo.ToPtr("User"),
				}
				t.Run("must succeed", func(t *testing.T) {
					authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
					assert.NoError(t, err)
					assert.NotNil(t, authResp)
					assert.NotEmpty(t, authResp.AccessToken)
					assert.NotEmpty(t, authResp.RefreshToken)
					assert.EqualValues(t, "Bearer", *authResp.TokenType)
					assert.Greater(t, *authResp.ExpiresIn, int64(0))
				})
				t.Run("duplicate email must fail with 409 Conflict", func(t *testing.T) {
					_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
					kathttpc.AssertStatusConflict(t, err)
				})
				t.Run("invalid email format must fail with 400 Bad Request", func(t *testing.T) {
					invalidEmailReq := &swagger.SignupRequest{
						Email:     (*strfmt.Email)(lo.ToPtr("invalid-email")),
						Password:  lo.ToPtr("password123"),
						FirstName: lo.ToPtr("Test"),
						LastName:  lo.ToPtr("User"),
					}
					_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/signup", nil, invalidEmailReq)
					kathttpc.AssertStatusBadRequest(t, err)
				})
				t.Run("missing required fields must fail with 400 Bad Request", func(t *testing.T) {
					incompleteReq := &swagger.SignupRequest{
						Email:    (*strfmt.Email)(lo.ToPtr("incomplete@example.com")),
						Password: lo.ToPtr("password123"),
						// Missing FirstName and LastName
					}
					_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/signup", nil, incompleteReq)
					kathttpc.AssertStatusBadRequest(t, err)
				})
			})

			t.Run("POST /auth/signin", func(t *testing.T) {
				// First create a user
				signupReq := &swagger.SignupRequest{
					Email:     (*strfmt.Email)(lo.ToPtr("signin@example.com")),
					Password:  lo.ToPtr("password123"),
					FirstName: lo.ToPtr("Signin"),
					LastName:  lo.ToPtr("User"),
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.AuthResponse](
					ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
				assert.NoError(t, err)

				signinReq := &swagger.SigninRequest{
					Email:    (*strfmt.Email)(lo.ToPtr("signin@example.com")),
					Password: lo.ToPtr("password123"),
				}
				t.Run("valid credentials must succeed", func(t *testing.T) {
					authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/signin", nil, signinReq)
					assert.NoError(t, err)
					assert.NotNil(t, authResp)
					assert.NotEmpty(t, authResp.AccessToken)
					assert.NotEmpty(t, authResp.RefreshToken)
				})
				t.Run("invalid credentials must fail with 401 Unauthorized", func(t *testing.T) {
					invalidReq := &swagger.SigninRequest{
						Email:    (*strfmt.Email)(lo.ToPtr("signin@example.com")),
						Password: lo.ToPtr("wrongpassword"),
					}
					_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/signin", nil, invalidReq)
					kathttpc.AssertStatusUnauthorized(t, err)
				})
				t.Run("non-existent user must fail with 401 Unauthorized", func(t *testing.T) {
					nonExistentReq := &swagger.SigninRequest{
						Email:    (*strfmt.Email)(lo.ToPtr("nonexistent@example.com")),
						Password: lo.ToPtr("password123"),
					}
					_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/signin", nil, nonExistentReq)
					kathttpc.AssertStatusUnauthorized(t, err)
				})
			})

			t.Run("POST /auth/refresh", func(t *testing.T) {
				// First create a user and get tokens
				signupReq := &swagger.SignupRequest{
					Email:     (*strfmt.Email)(lo.ToPtr("refresh@example.com")),
					Password:  lo.ToPtr("password123"),
					FirstName: lo.ToPtr("Refresh"),
					LastName:  lo.ToPtr("User"),
				}
				authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.AuthResponse](
					ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
				assert.NoError(t, err)

				refreshReq := &swagger.RefreshRequest{
					RefreshToken: authResp.RefreshToken,
				}
				t.Run("with valid refresh token must succeed", func(t *testing.T) {
					newAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.RefreshRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/refresh", nil, refreshReq)
					assert.NoError(t, err)
					assert.NotNil(t, newAuthResp)
					assert.NotEmpty(t, newAuthResp.AccessToken)
					assert.NotEmpty(t, newAuthResp.RefreshToken)
					// New tokens should be different from original
					assert.NotEqual(t, *authResp.AccessToken, *newAuthResp.AccessToken)
					assert.NotEqual(t, *authResp.RefreshToken, *newAuthResp.RefreshToken)
				})
				t.Run("with invalid refresh token must fail with 401 Unauthorized", func(t *testing.T) {
					invalidRefreshReq := &swagger.RefreshRequest{
						RefreshToken: lo.ToPtr("invalid-token"),
					}
					_, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.RefreshRequest, swagger.AuthResponse](
						ctx, &appConfig.Server, "api/v1/auth/refresh", nil, invalidRefreshReq)
					kathttpc.AssertStatusUnauthorized(t, err)
				})
			})

			t.Run("POST /auth/signout", func(t *testing.T) {
				// First create a user and get tokens
				signupReq := &swagger.SignupRequest{
					Email:     (*strfmt.Email)(lo.ToPtr("signout@example.com")),
					Password:  lo.ToPtr("password123"),
					FirstName: lo.ToPtr("Signout"),
					LastName:  lo.ToPtr("User"),
				}
				authResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SignupRequest, swagger.AuthResponse](
					ctx, &appConfig.Server, "api/v1/auth/signup", nil, signupReq)
				assert.NoError(t, err)

				t.Run("must succeed", func(t *testing.T) {
					headers := map[string][]string{
						"Authorization": {"Bearer " + *authResp.AccessToken},
					}
					signoutReq := map[string]string{
						"refreshToken": *authResp.RefreshToken,
					}
					msgResp, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, swagger.MessageResponse](
						ctx, &appConfig.Server, "api/v1/auth/signout", headers, &signoutReq)
					assert.NoError(t, err)
					assert.NotNil(t, msgResp)
					assert.Contains(t, *msgResp.Message, "signed out")
				})
			})
		})
	})

	t.Run("User Management API (Admin-only routes)", func(t *testing.T) {
		// Create admin user using sample data
		adminSigninReq := &swagger.SigninRequest{
			Email:    (*strfmt.Email)(lo.ToPtr("testadmin@example.com")),
			Password: lo.ToPtr("password123"),
		}
		adminAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, adminSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, adminAuthResp)
		assert.NotNil(t, adminAuthResp.AccessToken)
		adminHeaders := map[string][]string{
			"Authorization": {"Bearer " + *adminAuthResp.AccessToken},
		}

		// Create regular user using sample data
		userSigninReq := &swagger.SigninRequest{
			Email:    (*strfmt.Email)(lo.ToPtr("testuser@example.com")),
			Password: lo.ToPtr("password123"),
		}
		userAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, userSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, userAuthResp)
		assert.NotNil(t, userAuthResp.AccessToken)
		userHeaders := map[string][]string{
			"Authorization": {"Bearer " + *userAuthResp.AccessToken},
		}

		t.Run("GET /users", func(t *testing.T) {
			t.Run("admin user must succeed", func(t *testing.T) {
				userListResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserListResponse](
					ctx, &appConfig.Server, "api/v1/users", adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userListResp)
				assert.NotNil(t, userListResp.Users)
				assert.Greater(t, len(userListResp.Users), 0)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserListResponse](
					ctx, &appConfig.Server, "api/v1/users", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserListResponse](
					ctx, &appConfig.Server, "api/v1/users", nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("GET /users/{userId}", func(t *testing.T) {
			userID := "test-user-1"
			t.Run("admin user must succeed", func(t *testing.T) {
				userProfile, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfile](
					ctx, &appConfig.Server, "api/v1/users/"+userID, adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userProfile)
				assert.Equal(t, userID, *userProfile.ID)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfile](
					ctx, &appConfig.Server, "api/v1/users/"+userID, userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
		})

		t.Run("GET /users/{userId}/roles", func(t *testing.T) {
			userID := "test-admin-1"
			t.Run("admin user must succeed", func(t *testing.T) {
				rolesResp, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserRolesResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, rolesResp)
				assert.Contains(t, rolesResp.Roles, "admin")
				assert.Contains(t, rolesResp.Roles, "user")
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserRolesResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
		})

		t.Run("POST /users/{userId}/roles", func(t *testing.T) {
			userID := "test-user-1"
			t.Run("admin user must succeed in assigning role", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "moderator",
				}
				msgResp, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, swagger.MessageResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", adminHeaders, &assignRoleReq)
				assert.NoError(t, err)
				assert.NotNil(t, msgResp)
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "moderator",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, swagger.MessageResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", userHeaders, &assignRoleReq)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				assignRoleReq := map[string]string{
					"roleName": "moderator",
				}
				_, _, err := kathttpc.LocalHttpJsonPostRequest[map[string]string, swagger.MessageResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles", nil, &assignRoleReq)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})

		t.Run("DELETE /users/{userId}/roles/{roleName}", func(t *testing.T) {
			userID := "test-user-1"
			roleName := "moderator"
			t.Run("admin user must succeed in removing role", func(t *testing.T) {
				msgResp, _, err := kathttpc.LocalHttpJsonDeleteRequest[swagger.MessageResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles/"+roleName, adminHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, msgResp)
				assert.Contains(t, *msgResp.Message, "Role removed")
			})
			t.Run("regular user must fail with 403 Forbidden", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[swagger.MessageResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles/"+roleName, userHeaders)
				kathttpc.AssertStatusForbidden(t, err)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonDeleteRequest[swagger.MessageResponse](
					ctx, &appConfig.Server, "api/v1/users/"+userID+"/roles/"+roleName, nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})
	})

	t.Run("JWT Token Security", func(t *testing.T) {
		t.Run("malformed Authorization header must fail", func(t *testing.T) {
			malformedHeaders := map[string][]string{
				"Authorization": {"InvalidToken"},
			}
			_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfile](
				ctx, &appConfig.Server, "api/v1/me/profile", malformedHeaders)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
		t.Run("invalid JWT token must fail", func(t *testing.T) {
			invalidHeaders := map[string][]string{
				"Authorization": {"Bearer invalid.jwt.token"},
			}
			_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfile](
				ctx, &appConfig.Server, "api/v1/me/profile", invalidHeaders)
			kathttpc.AssertStatusUnauthorized(t, err)
		})
	})

	t.Run("User Profile API (Authenticated routes)", func(t *testing.T) {
		// Create regular user using sample data
		userSigninReq := &swagger.SigninRequest{
			Email:    (*strfmt.Email)(lo.ToPtr("testuser@example.com")),
			Password: lo.ToPtr("password123"),
		}
		userAuthResp, _, err := kathttpc.LocalHttpJsonPostRequest[swagger.SigninRequest, swagger.AuthResponse](
			ctx, &appConfig.Server, "api/v1/auth/signin", nil, userSigninReq)
		assert.NoError(t, err)
		assert.NotNil(t, userAuthResp)
		assert.NotNil(t, userAuthResp.AccessToken)
		userHeaders := map[string][]string{
			"Authorization": {"Bearer " + *userAuthResp.AccessToken},
		}

		t.Run("GET /me/profile", func(t *testing.T) {
			t.Run("authenticated user must succeed", func(t *testing.T) {
				userProfile, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfile](
					ctx, &appConfig.Server, "api/v1/me/profile", userHeaders)
				assert.NoError(t, err)
				assert.NotNil(t, userProfile)
				assert.Equal(t, "testuser@example.com", string(*userProfile.Email))
				assert.Equal(t, "Test", *userProfile.FirstName)
				assert.Equal(t, "User", *userProfile.LastName)
			})
			t.Run("unauthenticated request must fail with 401 Unauthorized", func(t *testing.T) {
				_, _, err := kathttpc.LocalHttpJsonGetRequest[swagger.UserProfile](
					ctx, &appConfig.Server, "api/v1/me/profile", nil)
				kathttpc.AssertStatusUnauthorized(t, err)
			})
		})
	})
}

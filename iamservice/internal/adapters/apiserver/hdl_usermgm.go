package apiserver

import (
	"errors"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/kathttp_echo"
)

// getMyUserHandler handles getting current user profile
func getMyUserHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principle, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		if authUserResponse, err := uc.LoadUserByID(ctx, principle, principle.UserID); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, authUserResponse)
		}
	}
}

// getUserByIdHandler handles getting user by ID (admin only)
func getUserByIdHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		userID := c.Param("userId")
		if authUserResponse, err := uc.LoadUserByID(ctx, principal, userID); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, authUserResponse)
		}
	}
}

// updateAuthUserHandler handles updating user details (first name and last name)
func updateAuthUserHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		userID := c.Param("userId")

		// Parse request body
		var req swagger.UpdateAuthUserRequest
		if err := c.Bind(&req); err != nil {
			return kathttp_echo.ReportBadRequest(errors.New("invalid request body"))
		}

		// Update user details
		if err := uc.UpdateUserDetails(ctx, principal, userID, req.FirstName, req.LastName); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		// Return updated user data
		if authUserResponse, err := uc.LoadUserByID(ctx, principal, userID); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, authUserResponse)
		}
	}
}

// listAllUsersByTenantHandler handles listing all tenant users with pagination
func listAllUsersByTenantHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		// Parse pagination parameters
		page := 1
		if pageStr := c.QueryParam("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		limit := 20
		if limitStr := c.QueryParam("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		if userList, err := uc.ListAllUsersByTenant(ctx, principal, principal.TenantID, page, limit); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, userList)
		}
	}
}

// listAllUsersHandler handles listing all users with pagination (sysadmin only)
func listAllUsersHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		// Parse pagination parameters
		page := 1
		if pageStr := c.QueryParam("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		limit := 20
		if limitStr := c.QueryParam("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		if userList, err := uc.ListAllUsers(ctx, principal, page, limit); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, userList)
		}
	}
}

// getUserRolesHandler handles getting user roles (admin only)
func getUserRolesHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		userID := c.Param("userId")

		if userRoles, err := uc.GetUserRoles(ctx, principal, userID); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, userRoles)
		}
	}
}

// assignUserRoleHandler handles assigning a role to a user (admin only)
func assignUserRoleHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		userID := c.Param("userId")

		// Parse request body
		var req swagger.AssignUserRoleRequest
		if err := c.Bind(&req); err != nil {
			return kathttp_echo.ReportBadRequest(errors.New("invalid request body"))
		}

		if err = uc.AssignUserRole(ctx, principal, userID, string(req.RoleName)); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, struct{}{})
	}
}

// deleteUserRoleHandler handles removing a role from a user (admin only)
func deleteUserRoleHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		userID := c.Param("userId")
		roleName := c.Param("roleName")

		if err = uc.DeleteUserRole(ctx, principal, userID, roleName); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, struct{}{})
	}
}

// getUserProfileHandler handles getting user profile by user ID (admin only)
func getUserProfileHandler(uc *usecase.UserProfileMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		userID := c.Param("userId")

		if userProfile, err := uc.GetUserProfileByUserID(ctx, principal, userID); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, userProfile)
		}
	}
}

// updateUserProfileHandler handles updating user profile by user ID (admin only)
func updateUserProfileHandler(uc *usecase.UserProfileMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		userID := c.Param("userId")

		// Parse request body
		var req swagger.UpdateUserProfileRequest
		if err := c.Bind(&req); err != nil {
			return kathttp_echo.ReportBadRequest(errors.New("invalid request body"))
		}

		if userProfile, err := uc.UpdateUserProfileByUserID(ctx, principal, userID, &req); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		} else {
			return c.JSON(http.StatusOK, userProfile)
		}
	}
}

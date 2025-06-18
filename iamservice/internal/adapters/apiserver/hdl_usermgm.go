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

// getUserProfileHandler handles getting current user profile
func getUserProfileHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Extract user ID from JWT token
		userID, err := serverhelp.UserIDFromValidatedToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		userProfile, err := uc.GetCurrentUserProfile(ctx, userID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, userProfile)
	}
}

// getUserByIdHandler handles getting user by ID (admin only)
func getUserByIdHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Get user ID from path parameter and validate
		userID := c.Param("userId")
		if userID == "" {
			return kathttp_echo.ReportBadRequest(errors.New("user ID is required"))
		}

		userProfile, err := uc.GetUserByID(ctx, userID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, userProfile)
	}
}

// listUsersHandler handles listing all users with pagination (admin only)
func listUsersHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

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

		userList, err := uc.ListUsers(ctx, page, limit)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, userList)
	}
}

// getUserRolesHandler handles getting user roles (admin only)
func getUserRolesHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Get user ID from path parameter and validate
		userID := c.Param("userId")
		if userID == "" {
			return kathttp_echo.ReportBadRequest(errors.New("user ID is required"))
		}

		userRoles, err := uc.GetUserRoles(ctx, userID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, userRoles)
	}
}

// assignUserRoleHandler handles assigning a role to a user (admin only)
func assignUserRoleHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Extract requesting user info from JWT token
		requestingUserID, err := serverhelp.UserIDFromValidatedToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		// Get user ID from path parameter and validate
		userID := c.Param("userId")
		if userID == "" {
			return kathttp_echo.ReportBadRequest(errors.New("user ID is required"))
		}

		// Parse request body
		var req swagger.AssignRoleRequest
		if err := c.Bind(&req); err != nil {
			return kathttp_echo.ReportBadRequest(errors.New("invalid request body"))
		}

		// Validate role name
		if req.RoleName == nil || *req.RoleName == "" {
			return kathttp_echo.ReportBadRequest(errors.New("role name is required"))
		}

		response, err := uc.AssignUserRole(ctx, userID, *req.RoleName, requestingUserID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, response)
	}
}

// deleteUserRoleHandler handles removing a role from a user (admin only)
func deleteUserRoleHandler(uc *usecase.UserMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Get user ID from path parameter and validate
		userID := c.Param("userId")
		if userID == "" {
			return kathttp_echo.ReportBadRequest(errors.New("user ID is required"))
		}

		// Get role name from query parameter and validate
		roleName := c.Param("roleName")
		if roleName == "" {
			return kathttp_echo.ReportBadRequest(errors.New("role name is required"))
		}

		response, err := uc.DeleteUserRole(ctx, userID, roleName)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, response)
	}
}

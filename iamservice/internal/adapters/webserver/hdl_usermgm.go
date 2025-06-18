package webserver

import (
	"errors"
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates"
	"github.com/mobiletoly/gokatana/kathttp_echo"
)

// UserMgmWebHandlers handles user management-related web requests
type UserMgmWebHandlers struct {
	userMgmUC   *usecase.UserMgm
	authUC      *usecase.AuthUser
	authHandler *AuthWebHandlers
}

// NewUserMgmWebHandlers creates a new instance of UserMgmWebHandlers
func NewUserMgmWebHandlers(userMgmUC *usecase.UserMgm, authUC *usecase.AuthUser, authHandler *AuthWebHandlers) *UserMgmWebHandlers {
	return &UserMgmWebHandlers{
		userMgmUC:   userMgmUC,
		authUC:      authUC,
		authHandler: authHandler,
	}
}

// UsersListHandler renders the users list
// Note: Admin role validation is handled by middleware
func (h *UserMgmWebHandlers) UsersListHandler(c echo.Context) error {
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

	// Call the real UserMgm.ListUsers method
	userListResponse, err := h.userMgmUC.ListUsers(ctx, page, limit)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	users := userListResponse.Users
	return renderTemplateComponent(c, "Users", templates.UsersList(users))
}

// UserFormHandler renders the user form
func (h *UserMgmWebHandlers) UserFormHandler(c echo.Context) error {
	return renderTemplateComponent(c, "Add User", templates.UserForm())
}

// UserDetailHandler renders a single user's details
func (h *UserMgmWebHandlers) UserDetailHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")

	user, err := h.userMgmUC.GetUserByID(ctx, userID)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}
	return renderTemplateComponent(c, "User Details", templates.UserDetail(user))
}

// CreateUserHandler handles user creation
func (h *UserMgmWebHandlers) CreateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()

	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))
	firstName := strings.TrimSpace(c.FormValue("firstName"))
	lastName := strings.TrimSpace(c.FormValue("lastName"))

	if email == "" || password == "" || firstName == "" || lastName == "" {
		return kathttp_echo.ReportBadRequest(errors.New("all fields are required"))
	}

	// Convert email to strfmt.Email
	emailFormat := strfmt.Email(email)

	// Create SignupRequest using the builder pattern
	signupReq := swagger.NewSignupRequestBuilder().
		Email(&emailFormat).
		Password(&password).
		FirstName(&firstName).
		LastName(&lastName).
		Build()

	// Call the real AuthUser.SignUp method
	_, err := h.authUC.SignUp(ctx, signupReq)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	// Return success template
	userName := firstName + " " + lastName
	return templates.UserFormSuccess(userName).Render(ctx, c.Response().Writer)
}

// UserRolesHandler renders user roles management
// Note: Admin role validation is handled by middleware
func (h *UserMgmWebHandlers) UserRolesHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")

	// Call the real UserMgm.GetUserRoles method
	userRolesResponse, err := h.userMgmUC.GetUserRoles(ctx, userID)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	roles := userRolesResponse.Roles
	return renderTemplateComponent(c, "User Roles", templates.UserRoles(userID, roles))
}

// AssignRoleHandler handles role assignment
func (h *UserMgmWebHandlers) AssignRoleHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")
	roleName := strings.TrimSpace(c.FormValue("roleName"))

	if roleName == "" {
		return kathttp_echo.ReportBadRequest(errors.New("role name is required"))
	}

	// Get the requesting user ID from authentication context
	requestingUserID, _ := h.authHandler.GetAuthenticatedUser(c)

	// Call the real UserMgm.AssignUserRole method
	_, err := h.userMgmUC.AssignUserRole(ctx, userID, roleName, requestingUserID)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	// Get updated roles and render the template
	userRolesResponse, err := h.userMgmUC.GetUserRoles(ctx, userID)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	roles := userRolesResponse.Roles
	return templates.UserRoles(userID, roles).Render(ctx, c.Response().Writer)
}

// DeleteRoleHandler handles role removal
func (h *UserMgmWebHandlers) DeleteRoleHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")
	roleName := c.Param("roleName")

	if roleName == "" {
		return kathttp_echo.ReportBadRequest(errors.New("role name is required"))
	}

	// Call the real UserMgm.DeleteUserRole method
	_, err := h.userMgmUC.DeleteUserRole(ctx, userID, roleName)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	// Get updated roles and render the template
	userRolesResponse, err := h.userMgmUC.GetUserRoles(ctx, userID)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	roles := userRolesResponse.Roles
	return templates.UserRoles(userID, roles).Render(ctx, c.Response().Writer)
}

// DeleteUserHandler handles user deletion
func (h *UserMgmWebHandlers) DeleteUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")

	if userID == "" {
		return kathttp_echo.ReportBadRequest(errors.New("user ID is required"))
	}

	// Call the real UserMgm.DeleteUser method
	_, err := h.userMgmUC.DeleteUser(ctx, userID)
	if err != nil {
		return kathttp_echo.ReportHTTPError(err)
	}

	// For HTMX, return empty content to remove the element from the DOM
	c.Response().WriteHeader(200)
	return nil
}

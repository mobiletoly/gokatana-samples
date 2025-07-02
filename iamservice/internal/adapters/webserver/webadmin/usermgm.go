package webadmin

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/admin"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
)

// UserMgmWebHandlers handles user management-related web requests
type UserMgmWebHandlers struct {
	userMgm *usecase.UserMgm
	authMgm *usecase.AuthMgm
}

// NewUserMgmWebHandlers creates a new instance of UserMgmWebHandlers
func NewUserMgmWebHandlers(userMgm *usecase.UserMgm, authMgm *usecase.AuthMgm) *UserMgmWebHandlers {
	return &UserMgmWebHandlers{
		userMgm: userMgm,
		authMgm: authMgm,
	}
}

// UsersListLoadHandler renders the users list
func (h *UserMgmWebHandlers) UsersListLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()

	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	var tenantID string

	if principal.IsSysAdmin() {
		// For sysadmin, check if tenant is specified in query param
		tenantID = c.QueryParam("tenant-selector")
		katapp.Logger(ctx).Debug("sysadmin tenant selection", "tenantParam", tenantID)
		if tenantID == "" {
			katapp.Logger(ctx).Debug("using default tenant from token", "tenantID", tenantID)
			tenantID = principal.TenantID
		}
	} else {
		tenantID = principal.TenantID
	}
	canCreateUser := principal.CanManageUser(tenantID)

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

	userListResponse, err := h.userMgm.ListAllUsersByTenant(ctx, principal, tenantID, page, limit)
	if err != nil {
		return err
	}
	users := userListResponse.Items

	// If sysadmin, show tenant selector
	if principal.IsSysAdmin() {
		tenantsListResponse, err := h.authMgm.GetAllTenants(ctx, principal)
		if err != nil {
			return err
		}

		// Check if this is an HTMX request targeting just the users list
		if c.Request().Header.Get("HX-Target") == "users-list" {
			return admin.UsersListContent(users, canCreateUser).Render(ctx, c.Response().Writer)
		}
		return renderTemplateComponent(c, "Users",
			admin.UsersListWithTenantSelector(users, tenantsListResponse.Items, tenantID, true, canCreateUser))
	}
	return renderTemplateComponent(c, "Users", admin.UsersList(users, canCreateUser))
}

// NewUserLoadHandler renders the user form
func (h *UserMgmWebHandlers) NewUserLoadHandler(c echo.Context) error {
	return renderTemplateComponent(c, "Add User", admin.UserForm())
}

// UserDetailLoadHandler renders a single user's details
func (h *UserMgmWebHandlers) UserDetailLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")

	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	authUserResponse, err := h.userMgm.LoadUserByID(ctx, principal, userID)
	if err != nil {
		return err
	}
	userRolesResponse, err := h.userMgm.GetUserRoles(ctx, principal, userID)
	if err != nil {
		return err
	}
	roles := userRolesResponse.Roles

	// Check if the current user can manage users (for showing admin buttons)
	canManageUsers := principal.CanManageUsers()
	return renderTemplateComponent(c, "User Details", admin.UserDetail(authUserResponse, roles, canManageUsers))
}

// CreateUserLoadHandler handles user creation
func (h *UserMgmWebHandlers) CreateUserLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID := strings.TrimSpace(c.FormValue("tenantId"))
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))
	firstName := strings.TrimSpace(c.FormValue("firstName"))
	lastName := strings.TrimSpace(c.FormValue("lastName"))

	// Create SignupRequest using the builder pattern
	signupReq := swagger.NewSignupRequestBuilder().
		Email(email).
		FirstName(firstName).
		LastName(lastName).
		Password(password).
		Source("web").
		TenantId(tenantID).
		Build()

	if _, err := h.authMgm.SignUp(ctx, signupReq); err != nil {
		return err
	}

	userName := firstName + " " + lastName
	return admin.UserFormSuccess(userName).Render(ctx, c.Response().Writer)
}

// UserRolesLoadHandler renders user roles management
func (h *UserMgmWebHandlers) UserRolesLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	userRolesResponse, err := h.userMgm.GetUserRoles(ctx, principal, userID)
	if err != nil {
		return err
	}
	roles := userRolesResponse.Roles
	return renderTemplateComponent(c, "User Roles", admin.UserRoles(userID, roles))
}

// AssignRoleSubmitHandler handles role assignment
func (h *UserMgmWebHandlers) AssignRoleSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")
	roleName := strings.TrimSpace(c.FormValue("roleName"))
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	if err = h.userMgm.AssignUserRole(ctx, principal, userID, roleName); err != nil {
		return err
	}
	userRolesResponse, err := h.userMgm.GetUserRoles(ctx, principal, userID)
	if err != nil {
		return err
	}
	return admin.UserRoles(userID, userRolesResponse.Roles).Render(ctx, c.Response().Writer)
}

// DeleteRoleSubmitHandler handles role removal
func (h *UserMgmWebHandlers) DeleteRoleSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")
	roleName := c.Param("roleName")
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	if err = h.userMgm.DeleteUserRole(ctx, principal, userID, roleName); err != nil {
		return err
	}
	userRolesResponse, err := h.userMgm.GetUserRoles(ctx, principal, userID)
	if err != nil {
		return err
	}
	roles := userRolesResponse.Roles
	return admin.UserRoles(userID, roles).Render(ctx, c.Response().Writer)
}

// DeleteUserSubmitHandler handles user deletion
func (h *UserMgmWebHandlers) DeleteUserSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	if err = h.userMgm.DeleteUser(ctx, principal, userID); err != nil {
		return err
	}
	// For HTMX, return empty content to remove the element from the DOM
	c.Response().WriteHeader(200)
	return nil
}

// UserEditLoadHandler renders the user edit form
func (h *UserMgmWebHandlers) UserEditLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	userID := c.Param("id")
	authUserResponse, err := h.userMgm.LoadUserByID(ctx, principal, userID)
	if err != nil {
		return err
	}
	return renderTemplateComponent(c, "Edit User", admin.UserEditForm(authUserResponse))
}

// UpdateUserSubmitHandler handles user details updates
func (h *UserMgmWebHandlers) UpdateUserSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	userID := c.Param("id")
	firstName := strings.TrimSpace(c.FormValue("firstName"))
	lastName := strings.TrimSpace(c.FormValue("lastName"))

	if err = h.userMgm.UpdateUserDetails(ctx, principal, userID, firstName, lastName); err != nil {
		return err
	}
	userName := firstName + " " + lastName
	return admin.UserEditSuccess(userName).Render(ctx, c.Response().Writer)
}

// UserChangePasswordLoadHandler renders the change password form
func (h *UserMgmWebHandlers) UserChangePasswordLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("id")
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	authUserResponse, err := h.userMgm.LoadUserByID(ctx, principal, userID)
	if err != nil {
		return err
	}
	return renderTemplateComponent(c, "Change Password", admin.UserChangePasswordForm(authUserResponse))
}

// ChangePasswordSubmitHandler handles password changes
func (h *UserMgmWebHandlers) ChangePasswordSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	userID := c.Param("id")
	newPassword := strings.TrimSpace(c.FormValue("newPassword"))
	confirmPassword := strings.TrimSpace(c.FormValue("confirmPassword"))
	if newPassword == "" || confirmPassword == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "Both password fields are required")
	}

	if len(newPassword) < 8 {
		return katapp.NewErr(katapp.ErrInvalidInput, "Password must be at least 8 characters long")
	}

	if newPassword != confirmPassword {
		return katapp.NewErr(katapp.ErrInvalidInput, "Passwords do not match")
	}

	authUserResponse, err := h.userMgm.LoadUserByID(ctx, principal, userID)
	if err != nil {
		return err
	}

	if err = h.userMgm.ChangeUserPassword(ctx, principal, userID, newPassword); err != nil {
		return err
	}
	userName := authUserResponse.FirstName + " " + authUserResponse.LastName
	return admin.UserPasswordChangeSuccess(userName).Render(ctx, c.Response().Writer)
}

package webadmin

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/admin"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
)

// TenantMgmWebHandlers handles tenant management-related web requests
type TenantMgmWebHandlers struct {
	authMgm *usecase.AuthMgm
}

// NewTenantMgmWebHandlers creates a new instance of TenantMgmWebHandlers
func NewTenantMgmWebHandlers(authUC *usecase.AuthMgm) *TenantMgmWebHandlers {
	return &TenantMgmWebHandlers{
		authMgm: authUC,
	}
}

// TenantsListLoadHandler renders the tenants list
// Note: Sysadmin role validation is handled by middleware
func (h *TenantMgmWebHandlers) TenantsListLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	tenantsListResponse, err := h.authMgm.GetAllTenants(ctx, principal)
	if err != nil {
		return err
	}
	return renderTemplateComponent(c, "Tenants", admin.TenantsList(tenantsListResponse))
}

// NewTenantLoadHandler renders the tenant form
func (h *TenantMgmWebHandlers) NewTenantLoadHandler(c echo.Context) error {
	return renderTemplateComponent(c, "Add Tenant", admin.TenantForm())
}

// TenantDetailLoadHandler renders a single tenant's details
func (h *TenantMgmWebHandlers) TenantDetailLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	tenantID := c.Param("id")

	if tenantResponse, err := h.authMgm.GetTenantByID(ctx, principal, tenantID); err != nil {
		return err
	} else {
		return renderTemplateComponent(c, "Tenant Details", admin.TenantDetail(tenantResponse))
	}
}

// TenantEditLoadHandler renders the tenant edit form
func (h *TenantMgmWebHandlers) TenantEditLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	tenantID := c.Param("id")

	if tenantResponse, err := h.authMgm.GetTenantByID(ctx, principal, tenantID); err != nil {
		return err
	} else {
		return renderTemplateComponent(c, "Edit Tenant", admin.TenantEditForm(tenantResponse))
	}
}

// CreateTenantSubmitHandler handles tenant creation
func (h *TenantMgmWebHandlers) CreateTenantSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}

	id := strings.TrimSpace(c.FormValue("id"))
	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))
	createReq := &swagger.TenantCreateRequest{
		ID:          id,
		Name:        name,
		Description: description,
	}

	if tenant, err := h.authMgm.CreateTenant(ctx, principal, createReq); err != nil {
		return err
	} else {
		return admin.TenantFormSuccess(tenant.Name).Render(ctx, c.Response().Writer)
	}
}

// UpdateTenantSubmitHandler handles tenant updates
func (h *TenantMgmWebHandlers) UpdateTenantSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	tenantID := c.Param("id")
	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))
	updateReq := &swagger.TenantUpdateRequest{
		Name:        name,
		Description: description,
	}

	if tenant, err := h.authMgm.UpdateTenant(ctx, principal, tenantID, updateReq); err != nil {
		return err
	} else {
		return admin.TenantUpdateSuccess(tenant.Name).Render(ctx, c.Response().Writer)
	}
}

// DeleteTenantSubmitHandler handles tenant deletion
func (h *TenantMgmWebHandlers) DeleteTenantSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	principal, err := serverhelp.GetUserPrincipalFromToken(c)
	if err != nil {
		return err
	}
	tenantID := c.Param("id")

	err = h.authMgm.DeleteTenant(ctx, principal, tenantID)
	if err != nil {
		return err
	}

	// For HTMX, return empty content to remove the element from the DOM
	c.Response().WriteHeader(200)
	return nil
}

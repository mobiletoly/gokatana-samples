package apiserver

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"net/http"

	"github.com/go-openapi/strfmt"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp_echo"
)

// getAllTenantsHandler handles GET /api/v1/tenants
func getAllTenantsHandler(authMgm *usecase.AuthMgm) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		tenantsListResponse, err := authMgm.GetAllTenants(ctx, principal)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, tenantsListResponse)
	}
}

// getTenantByIdHandler handles GET /api/v1/tenants/{tenantId}
func getTenantByIdHandler(authMgm *usecase.AuthMgm) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		tenantID := c.Param("tenantId")
		if tenantID == "" {
			return kathttp_echo.ReportBadRequest(katapp.NewErr(katapp.ErrInvalidInput, "tenant ID is required"))
		}

		tenantResponse, err := authMgm.GetTenantByID(ctx, principal, tenantID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, tenantResponse)
	}
}

// createTenantHandler handles POST /api/v1/tenants
func createTenantHandler(authMgm *usecase.AuthMgm) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		var req swagger.TenantCreateRequest
		if err := c.Bind(&req); err != nil {
			return kathttp_echo.ReportBadRequest(katapp.NewErr(katapp.ErrInvalidInput, "invalid request body"))
		}
		tenantResponse, err := authMgm.CreateTenant(ctx, principal, &req)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusCreated, tenantResponse)
	}
}

// updateTenantHandler handles PUT /api/v1/tenants/{tenantId}
func updateTenantHandler(authMgm *usecase.AuthMgm) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		tenantID := c.Param("tenantId")

		var req swagger.TenantUpdateRequest
		if err := c.Bind(&req); err != nil {
			return kathttp_echo.ReportBadRequest(katapp.NewErr(katapp.ErrInvalidInput, "invalid request body"))
		}

		tenantResponse, err := authMgm.UpdateTenant(ctx, principal, tenantID, &req)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, tenantResponse)
	}
}

// deleteTenantHandler handles DELETE /api/v1/tenants/{tenantId}
func deleteTenantHandler(authMgm *usecase.AuthMgm) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		principal, err := serverhelp.GetUserPrincipalFromToken(c)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		tenantID := c.Param("tenantId")

		err = authMgm.DeleteTenant(ctx, principal, tenantID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}
		return c.JSON(http.StatusOK, struct{}{})
	}
}

// tenantModelToTenantResponse converts model.Tenant to swagger.TenantResponse
func tenantModelToTenantResponse(tenant *model.Tenant) *swagger.TenantResponse {
	return swagger.NewTenantResponseBuilder().
		CreatedAt(strfmt.DateTime(tenant.CreatedAt)).
		Description(tenant.Description).
		ID(tenant.ID).
		Name(tenant.Name).
		UpdatedAt(strfmt.DateTime(tenant.UpdatedAt)).
		Build()
}

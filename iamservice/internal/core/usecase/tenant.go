package usecase

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/outport"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase/internal"
	"github.com/mobiletoly/gokatana/katapp"
)

// GetAllTenants returns all tenants
func (a *AuthMgm) GetAllTenants(ctx context.Context, principal *UserPrincipal) (*swagger.TenantsResponse, error) {
	katapp.Logger(ctx).Debug("getting all tenants",
		"principal", principal.String(),
	)

	tenants, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) ([]*model.Tenant, error) {
		return a.authUserPersist.GetAllTenants(ctx, tx)
	})
	if err != nil {
		katapp.Logger(ctx).Error("failed to get all tenants", "error", err)
		return nil, err
	}

	filteredTenants := make([]*model.Tenant, 0)
	for _, tenant := range tenants {
		if principal.CanReadTenant(tenant.ID) {
			filteredTenants = append(filteredTenants, tenant)
		}
	}

	tenantResponses := make([]swagger.TenantResponse, len(filteredTenants))
	for i, tenant := range filteredTenants {
		tenantResponses[i] = *tenantModelToTenantResponse(tenant)
	}

	// Pagination (we hardcode for now)
	page := 1
	limit := 20
	tenants, pagination := internal.Paginate(filteredTenants, page, limit)
	response := swagger.NewTenantsResponseBuilder().
		Items(tenantResponses).
		Pagination(*pagination).
		Build()

	return response, nil
}

// GetTenantByID returns a tenant by ID
func (a *AuthMgm) GetTenantByID(
	ctx context.Context, principal *UserPrincipal, tenantID string,
) (*swagger.TenantResponse, error) {
	katapp.Logger(ctx).Debug("getting tenant by ID",
		"principal", principal.String(),
		"tenantID", tenantID,
	)
	if tenantID == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "tenant ID is required")
	}

	if !principal.CanReadTenant(tenantID) {
		msg := "insufficient permissions to get tenant"
		katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "tenantID", tenantID)
		return nil, katapp.NewErr(katapp.ErrNoPermissions, msg)
	}

	tenant, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (*model.Tenant, error) {
		return a.authUserPersist.GetTenantByID(ctx, tx, tenantID)
	})
	if err != nil {
		katapp.Logger(ctx).Error("failed to get tenant by ID", "tenantID", tenantID, "error", err)
		return nil, err
	}
	if tenant == nil {
		return nil, katapp.NewErr(katapp.ErrNotFound, "tenant not found")
	}

	return tenantModelToTenantResponse(tenant), nil
}

// CreateTenant creates a new tenant (sysadmin only)
func (a *AuthMgm) CreateTenant(
	ctx context.Context, principal *UserPrincipal, req *swagger.CreateTenantRequest,
) (*swagger.TenantResponse, error) {
	katapp.Logger(ctx).Info("creating tenant",
		"principal", principal.String(),
		"tenantID", req.Id,
		"name", req.Name)

	if !principal.IsSysAdmin() {
		panic("insufficient permissions to create tenant, must be handled on API level")
	}

	// Validate input
	if req.Id == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "tenant ID is required")
	}
	if len(req.Id) < 3 {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "tenant ID must be at least 3 characters long")
	}
	if req.Name == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "tenant name is required")
	}

	// Create tenant in a transaction
	tenant, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (*model.Tenant, error) {
		// Check if tenant already exists
		existingTenant, err := a.authUserPersist.GetTenantByID(ctx, tx, req.Id)
		if err != nil {
			var appErr *katapp.Err
			if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
				// Tenant doesn't exist, which is what we want
			} else {
				return nil, katapp.NewErr(katapp.ErrInternal, "failed to check tenant existence")
			}
		}
		if existingTenant != nil {
			return nil, katapp.NewErr(katapp.ErrDuplicate, "tenant with this ID already exists")
		}
		return a.authUserPersist.CreateTenant(ctx, tx, req)
	})
	if err != nil {
		return nil, err
	}

	katapp.Logger(ctx).Info("tenant created successfully", "tenantID", tenant.ID, "name", tenant.Name)
	return tenantModelToTenantResponse(tenant), nil
}

// UpdateTenant updates an existing tenant
func (a *AuthMgm) UpdateTenant(
	ctx context.Context, principal *UserPrincipal, tenantID string, req *swagger.UpdateTenantRequest,
) (*swagger.TenantResponse, error) {
	katapp.Logger(ctx).Info("updating tenant",
		"principal", principal.String(),
		"tenantID", tenantID)

	// Validate input
	if tenantID == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "tenant ID is required")
	}
	if req.Name == "" {
		return nil, katapp.NewErr(katapp.ErrInvalidInput, "tenant name is required")
	}

	if !principal.CanManageTenant(tenantID) {
		msg := "insufficient permissions to update tenant"
		katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "tenantID", tenantID)
		return nil, katapp.NewErr(katapp.ErrNoPermissions, msg)
	}

	// Update tenant in a transaction
	tenant, err := outport.TxWithResult(ctx, a.txPort, func(tx pgx.Tx) (*model.Tenant, error) {
		// Check if tenant exists
		existingTenant, err := a.authUserPersist.GetTenantByID(ctx, tx, tenantID)
		if err != nil {
			var appErr *katapp.Err
			if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
				return nil, katapp.NewErr(katapp.ErrNotFound, "tenant not found")
			}
			return nil, katapp.NewErr(katapp.ErrInternal, "failed to check tenant existence")
		}
		if existingTenant == nil {
			return nil, katapp.NewErr(katapp.ErrNotFound, "tenant not found")
		}

		// Update the tenant
		tenant, err := a.authUserPersist.UpdateTenant(ctx, tx, tenantID, req)
		if err != nil {
			katapp.Logger(ctx).Error("failed to update tenant", "tenantID", tenantID, "error", err)
		}
		return tenant, err
	})
	if err != nil {
		return nil, err
	}

	katapp.Logger(ctx).Info("tenant updated successfully", "tenantID", tenant.ID, "name", tenant.Name)
	return tenantModelToTenantResponse(tenant), nil
}

// DeleteTenant deletes a tenant
func (a *AuthMgm) DeleteTenant(ctx context.Context, principal *UserPrincipal, tenantID string) error {
	katapp.Logger(ctx).Info("deleting tenant",
		"principal", principal.String(),
		"tenantID", tenantID,
	)

	// Validate input
	if tenantID == "" {
		return katapp.NewErr(katapp.ErrInvalidInput, "tenant ID is required")
	}

	if !principal.CanManageTenant(tenantID) {
		msg := "insufficient permissions to delete tenant"
		katapp.Logger(ctx).Warn(msg, "principal", principal.String(), "tenantID", tenantID)
		return katapp.NewErr(katapp.ErrNoPermissions, msg)
	}

	// Delete tenant in a transaction
	err := a.txPort.Run(ctx, func(tx pgx.Tx) error {
		// Check if tenant exists
		existingTenant, err := a.authUserPersist.GetTenantByID(ctx, tx, tenantID)
		if err != nil {
			var appErr *katapp.Err
			if errors.As(err, &appErr) && appErr.Scope == katapp.ErrNotFound {
				return katapp.NewErr(katapp.ErrNotFound, "tenant not found")
			}
			return katapp.NewErr(katapp.ErrInternal, "failed to check tenant existence")
		}
		if existingTenant == nil {
			return katapp.NewErr(katapp.ErrNotFound, "tenant not found")
		}

		// Check if tenant has users
		users, err := a.authUserPersist.GetAllUsersByTenantID(ctx, tx, tenantID)
		if err != nil {
			return katapp.NewErr(katapp.ErrInternal, "failed to check tenant users")
		}
		if len(users) > 0 {
			return katapp.NewErr(katapp.ErrInvalidInput, "cannot delete tenant with existing users")
		}

		// Delete the tenant
		return a.authUserPersist.DeleteTenant(ctx, tx, tenantID)
	})
	if err != nil {
		return err
	}

	katapp.Logger(ctx).Info("tenant deleted successfully", "tenantID", tenantID)
	return nil
}

// tenantModelToTenantResponse converts model.Tenant to swagger.TenantResponse
func tenantModelToTenantResponse(tenant *model.Tenant) *swagger.TenantResponse {
	return &swagger.TenantResponse{
		Id:          tenant.ID,
		Name:        tenant.Name,
		Description: tenant.Description,
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
	}
}

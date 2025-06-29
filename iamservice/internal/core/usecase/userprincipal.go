package usecase

import "fmt"

// UserPrincipal represents the authenticated user context for use case operations
type UserPrincipal struct {
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// HasRole checks if the user principal has a specific role
func (up *UserPrincipal) HasRole(role string) bool {
	for _, r := range up.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (up *UserPrincipal) String() string {
	if up == nil {
		return "UserPrincipal{nil}"
	}
	return fmt.Sprintf("UserPrincipal{UserID: %s, TenantID: %s, Email: %s, Roles: %v}",
		up.UserID, up.TenantID, up.Email, up.Roles)
}

// IsSysAdmin checks if the user principal has sysadmin role
func (up *UserPrincipal) IsSysAdmin() bool {
	return up.HasRole("sysadmin")
}

// IsAdmin checks if the user principal has admin role
func (up *UserPrincipal) IsAdmin() bool {
	return up.HasRole("admin") || up.IsSysAdmin()
}

// IsUser checks if the user principal has user role
func (up *UserPrincipal) IsUser() bool {
	return up.HasRole("user")
}

// IsUserOnly checks if the user principal has only user role
func (up *UserPrincipal) IsUserOnly() bool {
	return up.IsUser() && !up.IsAdmin() && !up.IsSysAdmin()
}

func (up *UserPrincipal) CanFetchUser(targetUserID string, targetTenantID string) bool {
	return up.IsSysAdmin() || (up.IsAdmin() && up.TenantID == targetTenantID) ||
		(up.IsUserOnly() && up.UserID == targetUserID)
}

// CanUpdateUserDetails checks if the user principal can update users (update user does not include changing roles)
func (up *UserPrincipal) CanUpdateUserDetails(targetUserID string, targetTenantID string) bool {
	return up.IsSysAdmin() || (up.IsAdmin() && up.TenantID == targetTenantID) ||
		(up.IsUserOnly() && up.UserID == targetUserID)
}

func (up *UserPrincipal) CanListUsersForTenant(targetTenantID string) bool {
	return up.IsSysAdmin() || (up.IsAdmin() && up.TenantID == targetTenantID)
}

// CanManageUser checks if the user principal can assign roles to users
func (up *UserPrincipal) CanManageUser(targetTenantID string) bool {
	if up.IsSysAdmin() {
		return true
	}
	if up.IsAdmin() && up.TenantID == targetTenantID {
		return true
	}
	return false
}

func (up *UserPrincipal) CanManageTenant(targetTenantID string) bool {
	return up.IsSysAdmin() || (up.IsAdmin() && up.TenantID == targetTenantID)
}

func (up *UserPrincipal) CanReadTenant(targetTenantID string) bool {
	return up.IsSysAdmin() || (up.TenantID == targetTenantID)
}

// CanChangePasswords checks if the user can change other users' passwords
func (up *UserPrincipal) CanChangePasswords() bool {
	return up.IsSysAdmin() || up.IsAdmin()
}

// CanCreateUsers checks if the user can create new users
func (up *UserPrincipal) CanCreateUsers() bool {
	return up.IsSysAdmin() || up.IsAdmin()
}

// CanManageUsers checks if the user can manage users (sysadmin or admin)
func (up *UserPrincipal) CanManageUsers() bool {
	return up.IsSysAdmin() || up.IsAdmin()
}

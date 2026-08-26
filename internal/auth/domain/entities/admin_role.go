package entities

// AdminRole defines the granular role assigned to an admin user.
type AdminRole string

const (
	AdminRoleSuperAdmin      AdminRole = "SUPER_ADMIN"
	AdminRoleOperationsAdmin AdminRole = "OPERATIONS_ADMIN"
	AdminRoleFinanceAdmin    AdminRole = "FINANCE_ADMIN"
	AdminRoleSupportAdmin    AdminRole = "SUPPORT_ADMIN"
	AdminRoleSafetyAdmin     AdminRole = "SAFETY_ADMIN"
)

// AllAdminRoles returns the complete set of valid admin roles.
func AllAdminRoles() []AdminRole {
	return []AdminRole{
		AdminRoleSuperAdmin,
		AdminRoleOperationsAdmin,
		AdminRoleFinanceAdmin,
		AdminRoleSupportAdmin,
		AdminRoleSafetyAdmin,
	}
}

// IsValid reports whether the role is a recognized admin role.
func (r AdminRole) IsValid() bool {
	for _, valid := range AllAdminRoles() {
		if r == valid {
			return true
		}
	}
	return false
}

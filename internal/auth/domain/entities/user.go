package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
)

type User struct {
	ID              uuid.UUID
	FirstName       string
	LastName        string
	Email           *string // Nullable if only phone is used
	Phone           *string // Nullable if only email is used
	PasswordHash    string
	Status          domain.UserStatus
	EmailVerifiedAt *time.Time
	PhoneVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Roles           []domain.Role
}

func (u *User) HasRole(role domain.Role) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) IsAdmin() bool {
	return u.HasRole(domain.RoleAdmin) ||
		u.HasRole(domain.RoleSuperAdmin) ||
		u.HasRole(domain.RoleSupportAdmin) ||
		u.HasRole(domain.RoleSafetyAdmin) ||
		u.HasRole(domain.RoleFinanceAdmin)
}

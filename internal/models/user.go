package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string     `json:"email" gorm:"uniqueIndex;not null"`
	Password      string     `json:"-" gorm:"not null"` // "-" means don't include in JSON
	Name          string     `json:"name" gorm:"not null"`
	MonthlyIncome *float64   `json:"monthly_income" gorm:"type:decimal(15,2)"`
	Status        Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	LastLogin     *time.Time `json:"last_login,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// IsActive returns true if the user account is active
func (u *User) IsActive() bool {
	return u.Status.IsActive()
}

// IsAccessible returns true if the user can access the system
func (u *User) IsAccessible() bool {
	return u.Status.IsAccessible()
}

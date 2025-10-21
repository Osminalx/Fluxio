package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	Name            string      `json:"name" gorm:"not null"`
	ExpenseType     ExpenseType `json:"expense_type" gorm:"type:expense_type_enum;not null"` // PostgreSQL enum: needs, wants, savings
	Status          Status      `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time  `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	// Relations
	User     User      `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Expenses []Expense `json:"expenses" gorm:"foreignKey:CategoryID"`
}

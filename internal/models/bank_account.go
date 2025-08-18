package models

import (
	"time"

	"github.com/google/uuid"
)

type BankAccount struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	AccountName     string     `json:"account_name" gorm:"not null"`
	Balance         float64    `json:"balance" gorm:"type:decimal(15,2);not null;default:0.00"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

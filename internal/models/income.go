package models

import (
	"time"

	"github.com/google/uuid"
)

type Income struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Amount          float64    `json:"amount" gorm:"type:decimal(15,2);not null"`
	BankAccountID   uuid.UUID  `json:"bank_account_id" gorm:"type:uuid;not null"`
	Date            time.Time  `json:"date" gorm:"type:date;not null"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relaciones
	User        User        `json:"user" gorm:"foreignKey:UserID;references:ID"`
	BankAccount BankAccount `json:"bank_account" gorm:"foreignKey:BankAccountID;references:ID"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	FromAccountID   uuid.UUID  `json:"from_account_id" gorm:"type:uuid;not null"`
	ToAccountID     uuid.UUID  `json:"to_account_id" gorm:"type:uuid;not null"`
	Amount          float64    `json:"amount" gorm:"type:decimal(15,2);not null"`
	Description     *string    `json:"description"`
	Date            time.Time  `json:"date" gorm:"type:date;not null"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relaciones
	User        User        `json:"user" gorm:"foreignKey:UserID;references:ID"`
	FromAccount BankAccount `json:"from_account" gorm:"foreignKey:FromAccountID;references:ID"`
	ToAccount   BankAccount `json:"to_account" gorm:"foreignKey:ToAccountID;references:ID"`
}

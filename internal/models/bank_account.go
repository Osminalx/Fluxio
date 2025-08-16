package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BankAccount struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	AccountName string         `json:"account_name" gorm:"not null"`
	Balance     float64        `json:"balance" gorm:"type:decimal(15,2);not null;default:0.00"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Goal struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Name        string         `json:"name" gorm:"not null"`
	TotalAmount float64        `json:"total_amount" gorm:"type:decimal(15,2);not null"`
	SavedAmount float64        `json:"saved_amount" gorm:"type:decimal(15,2);not null;default:0.00"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

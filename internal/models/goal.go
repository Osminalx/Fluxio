package models

import (
	"time"

	"github.com/google/uuid"
)

type Goal struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Name            string     `json:"name" gorm:"not null"`
	TotalAmount     float64    `json:"total_amount" gorm:"type:decimal(15,2);not null"`
	SavedAmount     float64    `json:"saved_amount" gorm:"type:decimal(15,2);not null;default:0.00"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

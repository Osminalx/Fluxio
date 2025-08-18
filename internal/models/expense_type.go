package models

import (
	"time"

	"github.com/google/uuid"
)

type ExpenseType struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name            string     `json:"name" gorm:"not null"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relaciones
	Categories []Category `json:"categories" gorm:"foreignKey:ExpenseTypeID"`
}

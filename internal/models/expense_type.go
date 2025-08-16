package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExpenseType struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	Categories []Category `json:"categories" gorm:"foreignKey:ExpenseTypeID"`
}

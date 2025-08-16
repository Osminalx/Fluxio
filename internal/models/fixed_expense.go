package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FixedExpense struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Name      string         `json:"name" gorm:"not null"`
	Amount    float64        `json:"amount" gorm:"type:decimal(15,2);not null"`
	DueDate   time.Time      `json:"due_date" gorm:"type:date;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

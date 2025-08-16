package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Income struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Amount    float64        `json:"amount" gorm:"type:decimal(15,2);not null"`
	Date      time.Time      `json:"date" gorm:"type:date;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

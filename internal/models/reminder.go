package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Reminder struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Title        string         `json:"title" gorm:"not null"`
	Description  *string        `json:"description"`
	DueDate      time.Time      `json:"due_date" gorm:"type:date;not null"`
	IsCompleted  bool           `json:"is_completed" gorm:"default:false"`
	ReminderType string         `json:"reminder_type" gorm:"check:reminder_type IN ('bill', 'goal', 'budget_review')"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type Reminder struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Title           string     `json:"title" gorm:"not null"`
	Description     *string    `json:"description"`
	DueDate         time.Time  `json:"due_date" gorm:"type:date;not null"`
	IsCompleted     bool       `json:"is_completed" gorm:"default:false"`
	ReminderType    string     `json:"reminder_type" gorm:"check:reminder_type IN ('bill', 'goal', 'budget_review')"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relaciones
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

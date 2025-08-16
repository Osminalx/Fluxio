package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BudgetHistory struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BudgetID         uuid.UUID      `json:"budget_id" gorm:"type:uuid;not null"`
	UserID           uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	OldNeedsBudget   *float64       `json:"old_needs_budget" gorm:"type:decimal(15,2)"`
	OldWantsBudget   *float64       `json:"old_wants_budget" gorm:"type:decimal(15,2)"`
	OldSavingsBudget *float64       `json:"old_savings_budget" gorm:"type:decimal(15,2)"`
	NewNeedsBudget   *float64       `json:"new_needs_budget" gorm:"type:decimal(15,2)"`
	NewWantsBudget   *float64       `json:"new_wants_budget" gorm:"type:decimal(15,2)"`
	NewSavingsBudget *float64       `json:"new_savings_budget" gorm:"type:decimal(15,2)"`
	ChangedAt        time.Time      `json:"changed_at"`
	ChangeReason     *string        `json:"change_reason"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	Budget Budget `json:"budget" gorm:"foreignKey:BudgetID;references:ID"`
	User   User   `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

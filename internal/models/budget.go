package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Budget struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	MonthYear     time.Time      `json:"month_year" gorm:"type:date;not null"` // primer día del mes
	NeedsBudget   float64        `json:"needs_budget" gorm:"type:decimal(15,2);not null"`   // 50%
	WantsBudget   float64        `json:"wants_budget" gorm:"type:decimal(15,2);not null"`   // 30%
	SavingsBudget float64        `json:"savings_budget" gorm:"type:decimal(15,2);not null"` // 20%
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	User           User            `json:"user" gorm:"foreignKey:UserID;references:ID"`
	BudgetHistories []BudgetHistory `json:"budget_histories" gorm:"foreignKey:BudgetID"`
}

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name          string         `json:"name" gorm:"not null"`
	ExpenseTypeID uuid.UUID      `json:"expense_type_id" gorm:"type:uuid;not null"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	ExpenseType ExpenseType `json:"expense_type" gorm:"foreignKey:ExpenseTypeID;references:ID"`
	Expenses    []Expense   `json:"expenses" gorm:"foreignKey:CategoryID"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type FixedExpense struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Name            string     `json:"name" gorm:"not null"`
	Amount          float64    `json:"amount" gorm:"type:decimal(15,2);not null"`
	DueDate         time.Time  `json:"due_date" gorm:"type:date;not null"` // Day of month (1-31)
	CategoryID      *uuid.UUID `json:"category_id" gorm:"type:uuid"`       // Optional category to classify as needs/wants/savings
	BankAccountID   uuid.UUID  `json:"bank_account_id" gorm:"type:uuid"`   // Note: nullable for migration, validation in service layer ensures NOT NULL
	IsRecurring     bool       `json:"is_recurring" gorm:"default:true"`   // Whether it repeats monthly
	RecurrenceType  string     `json:"recurrence_type" gorm:"type:varchar(20);default:'monthly'"` // monthly, yearly
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	StatusChangedAt *time.Time `json:"status_changed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastProcessedAt *time.Time `json:"last_processed_at,omitempty"` // Last time it was auto-deducted
	NextDueDate     time.Time  `json:"next_due_date" gorm:"type:date"` // Next scheduled deduction (nullable for migration)

	// Relaciones
	User        User        `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Category    Category    `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
	BankAccount BankAccount `json:"bank_account,omitempty" gorm:"foreignKey:BankAccountID;references:ID"`
}

// GetDueDateForMonth returns the due date for this fixed expense in a specific year/month
// Handles edge cases for months with fewer days (e.g., Feb 30 -> Feb 28)
func (f FixedExpense) GetDueDateForMonth(year int, month time.Month) time.Time {
	day := f.DueDate.Day()
	
	// Get the last day of the target month
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	
	// If day is beyond the last day of target month, use the last day
	if day > lastDay {
		day = lastDay
	}
	
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// ShouldApplyForMonth determines if this fixed expense should be applied for a given month
func (f FixedExpense) ShouldApplyForMonth(year int, month time.Month) bool {
	if !f.IsRecurring || f.Status != StatusActive {
		return false
	}
	
	// If recurrence type is monthly, apply every month
	if f.RecurrenceType == "monthly" {
		return true
	}
	
	// For yearly, only apply on the same month as original due date
	if f.RecurrenceType == "yearly" {
		originalMonth := f.DueDate.Month()
		return originalMonth == month
	}
	
	return true
}

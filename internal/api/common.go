package api

import (
	"strings"
	"time"
)

// Common request structures
type ChangeStatusRequest struct {
	Status string  `json:"status" example:"inactive"`
	Reason *string `json:"reason,omitempty" example:"Error in the record"`
}

// Common response structures  
type BankAccountResponse struct {
	ID          string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AccountName string  `json:"account_name" example:"Main Checking"`
	Balance     float64 `json:"balance" example:"2500.00"`
}

// Common helper functions

// parseDate parses a date in format YYYY-MM-DD
func parseDate(dateStr string) (time.Time, error) {
	const layout = "2006-01-02"
	return time.Parse(layout, dateStr)
}

// extractIDFromPath extracts the ID from the URL
func extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	
	id := strings.TrimPrefix(path, prefix)
	// Remove any additional suffix (like /restore or /status)
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}
	
	return strings.TrimSpace(id)
}

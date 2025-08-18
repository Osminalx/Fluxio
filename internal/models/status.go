package models

import "time"

// Status represents the current state of a record
type Status string

const (
	// StatusActive indicates the record is active and in use
	StatusActive Status = "active"
	
	// StatusDeleted indicates the record has been deleted by the user
	StatusDeleted Status = "deleted"
	
	// StatusSuspended indicates the record is temporarily disabled
	StatusSuspended Status = "suspended"
	
	// StatusArchived indicates the record is kept for historical purposes but not active
	StatusArchived Status = "archived"
	
	// StatusPending indicates the record is waiting for validation or approval
	StatusPending Status = "pending"
	
	// StatusLocked indicates the record is locked due to security or dispute
	StatusLocked Status = "locked"
)

// IsActive returns true if the status indicates an active record
func (s Status) IsActive() bool {
	return s == StatusActive
}

// IsDeleted returns true if the status indicates a deleted record
func (s Status) IsDeleted() bool {
	return s == StatusDeleted
}

// IsAccessible returns true if the status allows normal user access
func (s Status) IsAccessible() bool {
	return s == StatusActive || s == StatusPending
}

// IsVisible returns true if the status allows the record to be shown to users
func (s Status) IsVisible() bool {
	return s == StatusActive || s == StatusPending || s == StatusSuspended
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}

// StatusChange represents a status change event for auditing
type StatusChange struct {
	OldStatus   Status     `json:"old_status"`
	NewStatus   Status     `json:"new_status"`
	ChangedAt   time.Time  `json:"changed_at"`
	Reason      *string    `json:"reason,omitempty"`
	ChangedBy   *string    `json:"changed_by,omitempty"`
}

// ValidateStatus checks if a status is valid
func ValidateStatus(status Status) bool {
	switch status {
	case StatusActive, StatusDeleted, StatusSuspended, StatusArchived, StatusPending, StatusLocked:
		return true
	default:
		return false
	}
}

// GetActiveStatuses returns statuses that should be considered for normal operations
func GetActiveStatuses() []Status {
	return []Status{StatusActive, StatusPending}
}

// GetVisibleStatuses returns statuses that should be visible to users
func GetVisibleStatuses() []Status {
	return []Status{StatusActive, StatusPending, StatusSuspended}
}

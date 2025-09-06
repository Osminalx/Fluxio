package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReminderService struct {
	db *gorm.DB
}

func NewReminderService() *ReminderService {
	return &ReminderService{
		db: db.DB,
	}
}

// CreateReminder creates a new reminder for a user
func (s *ReminderService) CreateReminder(userID uuid.UUID, title string, description *string, dueDate time.Time, reminderType string) (*models.Reminder, error) {
	// Validate reminder type
	validTypes := map[string]bool{
		"bill":          true,
		"goal":          true,
		"budget_review": true,
	}
	if !validTypes[reminderType] {
		return nil, errors.New("invalid reminder type. Must be one of: bill, goal, budget_review")
	}

	reminder := &models.Reminder{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        title,
		Description:  description,
		DueDate:      dueDate,
		IsCompleted:  false,
		ReminderType: reminderType,
		Status:       models.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(reminder).Error; err != nil {
		return nil, err
	}

	return reminder, nil
}

// GetReminderByID retrieves a reminder by ID for a specific user
func (s *ReminderService) GetReminderByID(userID, reminderID uuid.UUID) (*models.Reminder, error) {
	var reminder models.Reminder
	if err := s.db.Where("id = ? AND user_id = ? AND status IN ?", reminderID, userID, models.GetActiveStatuses()).First(&reminder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("reminder not found")
		}
		return nil, err
	}

	return &reminder, nil
}

// GetUserReminders retrieves all reminders for a user with filters
func (s *ReminderService) GetUserReminders(userID uuid.UUID, completed *bool, reminderType *string, limit, offset int) ([]*models.Reminder, error) {
	query := s.db.Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses())

	// Filter by completion status
	if completed != nil {
		query = query.Where("is_completed = ?", *completed)
	}

	// Filter by reminder type
	if reminderType != nil && *reminderType != "" {
		query = query.Where("reminder_type = ?", *reminderType)
	}

	// Order by due date (upcoming first)
	query = query.Order("due_date ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var reminders []*models.Reminder
	if err := query.Find(&reminders).Error; err != nil {
		return nil, err
	}

	return reminders, nil
}

// GetUpcomingReminders retrieves reminders due within the specified number of days
func (s *ReminderService) GetUpcomingReminders(userID uuid.UUID, daysAhead int) ([]*models.Reminder, error) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, daysAhead)

	var reminders []*models.Reminder
	if err := s.db.Where("user_id = ? AND status = ? AND is_completed = ? AND due_date >= ? AND due_date <= ?", 
		userID, models.StatusActive, false, now, futureDate).
		Order("due_date ASC").
		Find(&reminders).Error; err != nil {
		return nil, err
	}

	return reminders, nil
}

// GetOverdueReminders retrieves reminders that are past due and not completed
func (s *ReminderService) GetOverdueReminders(userID uuid.UUID) ([]*models.Reminder, error) {
	now := time.Now()

	var reminders []*models.Reminder
	if err := s.db.Where("user_id = ? AND status = ? AND is_completed = ? AND due_date < ?", 
		userID, models.StatusActive, false, now).
		Order("due_date ASC").
		Find(&reminders).Error; err != nil {
		return nil, err
	}

	return reminders, nil
}

// UpdateReminder updates a reminder's information
func (s *ReminderService) UpdateReminder(userID, reminderID uuid.UUID, updates map[string]interface{}) (*models.Reminder, error) {
	// Verify reminder exists and belongs to user
	reminder, err := s.GetReminderByID(userID, reminderID)
	if err != nil {
		return nil, err
	}

	// Validate reminder type if being updated
	if reminderType, ok := updates["reminder_type"].(string); ok {
		validTypes := map[string]bool{
			"bill":          true,
			"goal":          true,
			"budget_review": true,
		}
		if !validTypes[reminderType] {
			return nil, errors.New("invalid reminder type. Must be one of: bill, goal, budget_review")
		}
	}

	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	// Update reminder
	if err := s.db.Model(reminder).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Return updated reminder
	return s.GetReminderByID(userID, reminderID)
}

// CompleteReminder marks a reminder as completed
func (s *ReminderService) CompleteReminder(userID, reminderID uuid.UUID) (*models.Reminder, error) {
	updates := map[string]interface{}{
		"is_completed": true,
		"updated_at":   time.Now(),
	}

	return s.UpdateReminder(userID, reminderID, updates)
}

// IncompleteReminder marks a reminder as not completed
func (s *ReminderService) IncompleteReminder(userID, reminderID uuid.UUID) (*models.Reminder, error) {
	updates := map[string]interface{}{
		"is_completed": false,
		"updated_at":   time.Now(),
	}

	return s.UpdateReminder(userID, reminderID, updates)
}

// DeleteReminder soft deletes a reminder
func (s *ReminderService) DeleteReminder(userID, reminderID uuid.UUID) error {
	// Verify reminder exists and belongs to user
	_, err := s.GetReminderByID(userID, reminderID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"status":            models.StatusDeleted,
		"status_changed_at": time.Now(),
		"updated_at":        time.Now(),
	}

	return s.db.Model(&models.Reminder{}).Where("id = ? AND user_id = ?", reminderID, userID).Updates(updates).Error
}

// GetRemindersByType retrieves reminders of a specific type
func (s *ReminderService) GetRemindersByType(userID uuid.UUID, reminderType string, completed *bool, limit, offset int) ([]*models.Reminder, error) {
	query := s.db.Where("user_id = ? AND reminder_type = ? AND status IN ?", userID, reminderType, models.GetActiveStatuses())

	if completed != nil {
		query = query.Where("is_completed = ?", *completed)
	}

	query = query.Order("due_date ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var reminders []*models.Reminder
	if err := query.Find(&reminders).Error; err != nil {
		return nil, err
	}

	return reminders, nil
}

// GetReminderStats returns statistics about user's reminders
func (s *ReminderService) GetReminderStats(userID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total active reminders
	var totalCount int64
	s.db.Model(&models.Reminder{}).Where("user_id = ? AND status = ?", userID, models.StatusActive).Count(&totalCount)
	stats["total_reminders"] = totalCount

	// Completed reminders
	var completedCount int64
	s.db.Model(&models.Reminder{}).Where("user_id = ? AND status = ? AND is_completed = ?", userID, models.StatusActive, true).Count(&completedCount)
	stats["completed_reminders"] = completedCount

	// Pending reminders
	var pendingCount int64
	s.db.Model(&models.Reminder{}).Where("user_id = ? AND status = ? AND is_completed = ?", userID, models.StatusActive, false).Count(&pendingCount)
	stats["pending_reminders"] = pendingCount

	// Overdue reminders
	now := time.Now()
	var overdueCount int64
	s.db.Model(&models.Reminder{}).Where("user_id = ? AND status = ? AND is_completed = ? AND due_date < ?", 
		userID, models.StatusActive, false, now).Count(&overdueCount)
	stats["overdue_reminders"] = overdueCount

	// Upcoming reminders (next 7 days)
	futureDate := now.AddDate(0, 0, 7)
	var upcomingCount int64
	s.db.Model(&models.Reminder{}).Where("user_id = ? AND status = ? AND is_completed = ? AND due_date >= ? AND due_date <= ?", 
		userID, models.StatusActive, false, now, futureDate).Count(&upcomingCount)
	stats["upcoming_reminders"] = upcomingCount

	// Count by type
	typeStats := make(map[string]int64)
	types := []string{"bill", "goal", "budget_review"}
	for _, reminderType := range types {
		var count int64
		s.db.Model(&models.Reminder{}).Where("user_id = ? AND status = ? AND reminder_type = ?", 
			userID, models.StatusActive, reminderType).Count(&count)
		typeStats[reminderType] = count
	}
	stats["by_type"] = typeStats

	return stats, nil
}

// BulkCompleteReminders marks multiple reminders as completed
func (s *ReminderService) BulkCompleteReminders(userID uuid.UUID, reminderIDs []uuid.UUID) error {
	if len(reminderIDs) == 0 {
		return errors.New("no reminder IDs provided")
	}

	updates := map[string]interface{}{
		"is_completed": true,
		"updated_at":   time.Now(),
	}

	return s.db.Model(&models.Reminder{}).
		Where("id IN ? AND user_id = ? AND status IN ?", reminderIDs, userID, models.GetActiveStatuses()).
		Updates(updates).Error
}

// SnoozeReminder postpones a reminder by the specified number of days
func (s *ReminderService) SnoozeReminder(userID, reminderID uuid.UUID, days int) (*models.Reminder, error) {
	reminder, err := s.GetReminderByID(userID, reminderID)
	if err != nil {
		return nil, err
	}

	newDueDate := reminder.DueDate.AddDate(0, 0, days)
	updates := map[string]interface{}{
		"due_date":   newDueDate,
		"updated_at": time.Now(),
	}

	return s.UpdateReminder(userID, reminderID, updates)
}

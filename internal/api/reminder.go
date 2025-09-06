package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

// CreateReminderRequest represents the request body for creating a reminder
type CreateReminderRequest struct {
	Title        string    `json:"title" validate:"required,min=1,max=200"`
	Description  *string   `json:"description,omitempty"`
	DueDate      time.Time `json:"due_date" validate:"required"`
	ReminderType string    `json:"reminder_type" validate:"required,oneof=bill goal budget_review"`
}

// UpdateReminderRequest represents the request body for updating a reminder
type UpdateReminderRequest struct {
	Title        *string    `json:"title,omitempty"`
	Description  *string    `json:"description,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	ReminderType *string    `json:"reminder_type,omitempty"`
	IsCompleted  *bool      `json:"is_completed,omitempty"`
}

// CreateReminderHandler godoc
// @Summary Create a new reminder
// @Description Creates a new reminder for the authenticated user
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateReminderRequest true "Reminder data"
// @Success 201 {object} models.Reminder
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders [post]
func CreateReminderHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req CreateReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding reminder request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if req.ReminderType == "" {
		http.Error(w, "Reminder type is required", http.StatusBadRequest)
		return
	}

	if req.DueDate.IsZero() {
		http.Error(w, "Due date is required", http.StatusBadRequest)
		return
	}

	reminderService := services.NewReminderService()
	reminder, err := reminderService.CreateReminder(userID, req.Title, req.Description, req.DueDate, req.ReminderType)
	if err != nil {
		logger.Error("Error creating reminder: %v", err)
		http.Error(w, "Error creating reminder", http.StatusInternalServerError)
		return
	}

	logger.Info("Reminder created successfully: %s", reminder.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reminder)
}

// GetAllRemindersHandler godoc
// @Summary Get all reminders for user
// @Description Retrieves all reminders for the authenticated user with optional filtering
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Param type query string false "Filter by reminder type (bill, goal, budget_review)"
// @Param completed query boolean false "Filter by completion status"
// @Param upcoming query boolean false "Show only upcoming reminders"
// @Success 200 {array} models.Reminder
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders [get]
func GetAllRemindersHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	reminderType := r.URL.Query().Get("type")
	completedStr := r.URL.Query().Get("completed")
	upcomingStr := r.URL.Query().Get("upcoming")

	var completed *bool
	if completedStr != "" {
		c, _ := strconv.ParseBool(completedStr)
		completed = &c
	}

	reminderService := services.NewReminderService()
	
	var reminders []*models.Reminder
	
	if upcomingStr == "true" {
		days := 7 // Default to 7 days
		if daysStr := r.URL.Query().Get("days"); daysStr != "" {
			if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
				days = d
			}
		}
		reminders, err = reminderService.GetUpcomingReminders(userID, days)
	} else {
		var reminderTypePtr *string
		if reminderType != "" {
			reminderTypePtr = &reminderType
		}
		reminders, err = reminderService.GetUserReminders(userID, completed, reminderTypePtr, limit, offset)
	}

	if err != nil {
		logger.Error("Error retrieving reminders: %v", err)
		http.Error(w, "Error retrieving reminders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminders)
}

// GetReminderByIDHandler godoc
// @Summary Get reminder by ID
// @Description Retrieves a specific reminder by ID for the authenticated user
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Reminder ID"
// @Success 200 {object} models.Reminder
// @Failure 400 {string} string "Invalid reminder ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Reminder not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders/{id} [get]
func GetReminderByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract reminder ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	reminderIDStr := pathParts[len(pathParts)-1]
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		logger.Error("Invalid reminder ID format: %v", err)
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	reminderService := services.NewReminderService()
	reminder, err := reminderService.GetReminderByID(reminderID, userID)
	if err != nil {
		if err.Error() == "reminder not found" {
			http.Error(w, "Reminder not found", http.StatusNotFound)
			return
		}
		logger.Error("Error retrieving reminder: %v", err)
		http.Error(w, "Error retrieving reminder", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminder)
}

// UpdateReminderHandler godoc
// @Summary Update reminder
// @Description Updates a reminder for the authenticated user
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Reminder ID"
// @Param request body UpdateReminderRequest true "Update data"
// @Success 200 {object} models.Reminder
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Reminder not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders/{id} [patch]
func UpdateReminderHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract reminder ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	reminderIDStr := pathParts[len(pathParts)-1]
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		logger.Error("Invalid reminder ID format: %v", err)
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	var req UpdateReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding update reminder request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reminderService := services.NewReminderService()
	
	// Build updates map
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}
	if req.ReminderType != nil {
		updates["reminder_type"] = *req.ReminderType
	}
	if req.IsCompleted != nil {
		updates["is_completed"] = *req.IsCompleted
	}
	
	reminder, err := reminderService.UpdateReminder(userID, reminderID, updates)
	if err != nil {
		if err.Error() == "reminder not found" {
			http.Error(w, "Reminder not found", http.StatusNotFound)
			return
		}
		logger.Error("Error updating reminder: %v", err)
		http.Error(w, "Error updating reminder", http.StatusInternalServerError)
		return
	}

	logger.Info("Reminder updated successfully: %s", reminder.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminder)
}

// DeleteReminderHandler godoc
// @Summary Delete reminder
// @Description Soft deletes a reminder for the authenticated user
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Reminder ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid reminder ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Reminder not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders/{id} [delete]
func DeleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract reminder ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	reminderIDStr := pathParts[len(pathParts)-1]
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		logger.Error("Invalid reminder ID format: %v", err)
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	reminderService := services.NewReminderService()
	err = reminderService.DeleteReminder(userID, reminderID)
	if err != nil {
		if err.Error() == "reminder not found" {
			http.Error(w, "Reminder not found", http.StatusNotFound)
			return
		}
		logger.Error("Error deleting reminder: %v", err)
		http.Error(w, "Error deleting reminder", http.StatusInternalServerError)
		return
	}

	logger.Info("Reminder deleted successfully: %s", reminderID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Reminder deleted successfully",
	})
}

// CompleteReminderHandler godoc
// @Summary Mark reminder as completed
// @Description Marks a reminder as completed for the authenticated user
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Reminder ID"
// @Success 200 {object} models.Reminder
// @Failure 400 {string} string "Invalid reminder ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Reminder not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders/{id}/complete [post]
func CompleteReminderHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Extract reminder ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	reminderIDStr := pathParts[len(pathParts)-2] // -2 because last part is "complete"
	reminderID, err := uuid.Parse(reminderIDStr)
	if err != nil {
		logger.Error("Invalid reminder ID format: %v", err)
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	reminderService := services.NewReminderService()
	
	// Mark as completed using UpdateReminder
	updates := map[string]interface{}{
		"is_completed": true,
	}
	reminder, err := reminderService.UpdateReminder(userID, reminderID, updates)
	if err != nil {
		if err.Error() == "reminder not found" {
			http.Error(w, "Reminder not found", http.StatusNotFound)
			return
		}
		logger.Error("Error completing reminder: %v", err)
		http.Error(w, "Error completing reminder", http.StatusInternalServerError)
		return
	}

	logger.Info("Reminder marked as completed: %s", reminder.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminder)
}

// GetOverdueRemindersHandler godoc
// @Summary Get overdue reminders
// @Description Retrieves all overdue reminders for the authenticated user
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} models.Reminder
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders/overdue [get]
func GetOverdueRemindersHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	reminderService := services.NewReminderService()
	reminders, err := reminderService.GetOverdueReminders(userID)
	
	// Apply manual pagination if needed
	if limit > 0 && len(reminders) > limit {
		start := offset
		if start > len(reminders) {
			start = len(reminders)
		}
		end := start + limit
		if end > len(reminders) {
			end = len(reminders)
		}
		if start < end {
			reminders = reminders[start:end]
		} else {
			reminders = []*models.Reminder{}
		}
	}
	if err != nil {
		logger.Error("Error retrieving overdue reminders: %v", err)
		http.Error(w, "Error retrieving overdue reminders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminders)
}

// GetReminderStatsHandler godoc
// @Summary Get reminder statistics
// @Description Retrieves reminder statistics for the authenticated user
// @Tags reminders
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/reminders/stats [get]
func GetReminderStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	reminderService := services.NewReminderService()
	stats, err := reminderService.GetReminderStats(userID)
	if err != nil {
		logger.Error("Error retrieving reminder stats: %v", err)
		http.Error(w, "Error retrieving reminder stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

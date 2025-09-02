package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// Request and response structures
type CreateGoalRequest struct {
	Name        string  `json:"name" example:"Emergency Fund"`
	TotalAmount float64 `json:"total_amount" example:"10000.00"`
	SavedAmount float64 `json:"saved_amount,omitempty" example:"2500.00"`
}

type UpdateGoalRequest struct {
	Name        *string  `json:"name,omitempty" example:"Updated Goal Name"`
	TotalAmount *float64 `json:"total_amount,omitempty" example:"12000.00"`
	SavedAmount *float64 `json:"saved_amount,omitempty" example:"3500.00"`
}

type GoalResponse struct {
	ID              string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name            string  `json:"name" example:"Emergency Fund"`
	TotalAmount     float64 `json:"total_amount" example:"10000.00"`
	SavedAmount     float64 `json:"saved_amount" example:"2500.00"`
	ProgressPercent float64 `json:"progress_percent" example:"25.0"`
	Status          string  `json:"status" example:"active"`
	StatusChangedAt *string `json:"status_changed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt       string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       string  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type GoalsListResponse struct {
	Goals []GoalResponse `json:"goals"`
	Count int            `json:"count" example:"3"`
}

// Helper function to convert model to response
func convertGoalToResponse(goal *models.Goal) GoalResponse {
	progressPercent := 0.0
	if goal.TotalAmount > 0 {
		progressPercent = (goal.SavedAmount / goal.TotalAmount) * 100
	}

	response := GoalResponse{
		ID:              goal.ID.String(),
		Name:            goal.Name,
		TotalAmount:     goal.TotalAmount,
		SavedAmount:     goal.SavedAmount,
		ProgressPercent: progressPercent,
		Status:          string(goal.Status),
		CreatedAt:       goal.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       goal.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if goal.StatusChangedAt != nil {
		statusChangedAtStr := goal.StatusChangedAt.Format("2006-01-02T15:04:05Z07:00")
		response.StatusChangedAt = &statusChangedAtStr
	}

	return response
}

// CreateGoalHandler creates a new goal
// @Summary Create a new goal
// @Description Creates a new savings goal for the authenticated user
// @Tags goals
// @Accept json
// @Produce json
// @Param goal body CreateGoalRequest true "Goal data"
// @Success 201 {object} GoalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals [post]
func CreateGoalHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "Goal name is required", http.StatusBadRequest)
		return
	}
	if req.TotalAmount <= 0 {
		http.Error(w, "Total amount must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.SavedAmount < 0 {
		http.Error(w, "Saved amount cannot be negative", http.StatusBadRequest)
		return
	}
	if req.SavedAmount > req.TotalAmount {
		http.Error(w, "Saved amount cannot exceed total amount", http.StatusBadRequest)
		return
	}

	// Create goal model
	goal := models.Goal{
		Name:        req.Name,
		TotalAmount: req.TotalAmount,
		SavedAmount: req.SavedAmount,
	}

	// Create goal
	createdGoal, err := services.CreateGoal(userID, goal)
	if err != nil {
		logger.Error("Error creating goal: %v", err)
		http.Error(w, "Error creating goal", http.StatusInternalServerError)
		return
	}

	response := convertGoalToResponse(createdGoal)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetAllGoalsHandler retrieves all goals for the authenticated user
// @Summary Get all goals
// @Description Retrieves all goals for the authenticated user (active and deleted)
// @Tags goals
// @Produce json
// @Success 200 {object} GoalsListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals [get]
func GetAllGoalsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	goals, err := services.GetGoals(userID, true) // Include deleted
	if err != nil {
		logger.Error("Error getting goals: %v", err)
		http.Error(w, "Error retrieving goals", http.StatusInternalServerError)
		return
	}

	var goalResponses []GoalResponse
	for _, goal := range goals {
		goalResponses = append(goalResponses, convertGoalToResponse(&goal))
	}

	response := GoalsListResponse{
		Goals: goalResponses,
		Count: len(goalResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveGoalsHandler retrieves only active goals
// @Summary Get active goals
// @Description Retrieves only active goals for the authenticated user
// @Tags goals
// @Produce json
// @Success 200 {object} GoalsListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals/active [get]
func GetActiveGoalsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	goals, err := services.GetGoals(userID, false) // Don't include deleted
	if err != nil {
		logger.Error("Error getting active goals: %v", err)
		http.Error(w, "Error retrieving active goals", http.StatusInternalServerError)
		return
	}

	var goalResponses []GoalResponse
	for _, goal := range goals {
		goalResponses = append(goalResponses, convertGoalToResponse(&goal))
	}

	response := GoalsListResponse{
		Goals: goalResponses,
		Count: len(goalResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDeletedGoalsHandler retrieves only deleted goals
// @Summary Get deleted goals
// @Description Retrieves only deleted goals for the authenticated user
// @Tags goals
// @Produce json
// @Success 200 {object} GoalsListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals/deleted [get]
func GetDeletedGoalsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Get all goals and filter deleted ones
	allGoals, err := services.GetGoals(userID, true)
	if err != nil {
		logger.Error("Error getting goals: %v", err)
		http.Error(w, "Error retrieving deleted goals", http.StatusInternalServerError)
		return
	}

	var deletedGoals []models.Goal
	for _, goal := range allGoals {
		if goal.Status == models.StatusDeleted {
			deletedGoals = append(deletedGoals, goal)
		}
	}

	var goalResponses []GoalResponse
	for _, goal := range deletedGoals {
		goalResponses = append(goalResponses, convertGoalToResponse(&goal))
	}

	response := GoalsListResponse{
		Goals: goalResponses,
		Count: len(goalResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetGoalByIDHandler retrieves a specific goal by ID
// @Summary Get goal by ID
// @Description Retrieves a specific goal by its ID for the authenticated user
// @Tags goals
// @Produce json
// @Param id path string true "Goal ID"
// @Success 200 {object} GoalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals/{id} [get]
func GetGoalByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Extract goal ID from URL
	goalID := strings.TrimPrefix(r.URL.Path, "/api/v1/goals/")
	if goalID == "" {
		http.Error(w, "Goal ID is required", http.StatusBadRequest)
		return
	}

	goal, err := services.GetGoalByID(userID, goalID)
	if err != nil {
		logger.Error("Error getting goal by ID: %v", err)
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	response := convertGoalToResponse(goal)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateGoalHandler updates an existing goal
// @Summary Update goal
// @Description Updates an existing goal for the authenticated user
// @Tags goals
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Param goal body UpdateGoalRequest true "Updated goal data"
// @Success 200 {object} GoalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals/{id} [patch]
func UpdateGoalHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Extract goal ID from URL
	goalID := strings.TrimPrefix(r.URL.Path, "/api/v1/goals/")
	if goalID == "" {
		http.Error(w, "Goal ID is required", http.StatusBadRequest)
		return
	}

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Build update model
	updates := models.Goal{}
	if req.Name != nil {
		if *req.Name == "" {
			http.Error(w, "Goal name cannot be empty", http.StatusBadRequest)
			return
		}
		updates.Name = *req.Name
	}
	if req.TotalAmount != nil {
		if *req.TotalAmount <= 0 {
			http.Error(w, "Total amount must be greater than 0", http.StatusBadRequest)
			return
		}
		updates.TotalAmount = *req.TotalAmount
	}
	if req.SavedAmount != nil {
		if *req.SavedAmount < 0 {
			http.Error(w, "Saved amount cannot be negative", http.StatusBadRequest)
			return
		}
		updates.SavedAmount = *req.SavedAmount
	}

	// Additional validation: if both amounts are provided, check relationship
	if req.TotalAmount != nil && req.SavedAmount != nil {
		if *req.SavedAmount > *req.TotalAmount {
			http.Error(w, "Saved amount cannot exceed total amount", http.StatusBadRequest)
			return
		}
	}

	updatedGoal, err := services.UpdateGoal(userID, goalID, updates)
	if err != nil {
		logger.Error("Error updating goal: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Goal not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error updating goal", http.StatusInternalServerError)
		}
		return
	}

	response := convertGoalToResponse(updatedGoal)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteGoalHandler soft deletes a goal
// @Summary Delete goal
// @Description Soft deletes a goal (changes status to deleted)
// @Tags goals
// @Param id path string true "Goal ID"
// @Success 204 "Goal deleted successfully"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals/{id} [delete]
func DeleteGoalHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Extract goal ID from URL
	goalID := strings.TrimPrefix(r.URL.Path, "/api/v1/goals/")
	if goalID == "" {
		http.Error(w, "Goal ID is required", http.StatusBadRequest)
		return
	}

	err := services.DeleteGoal(userID, goalID)
	if err != nil {
		logger.Error("Error deleting goal: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Goal not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting goal", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreGoalHandler restores a deleted goal
// @Summary Restore goal
// @Description Restores a deleted goal (changes status back to active)
// @Tags goals
// @Param id path string true "Goal ID"
// @Success 200 {object} GoalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals/{id}/restore [post]
func RestoreGoalHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Extract goal ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/goals/")
	goalID := strings.TrimSuffix(path, "/restore")
	if goalID == "" || goalID == path {
		http.Error(w, "Goal ID is required", http.StatusBadRequest)
		return
	}

	restoredGoal, err := services.RestoreGoal(userID, goalID)
	if err != nil {
		logger.Error("Error restoring goal: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Goal not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error restoring goal", http.StatusInternalServerError)
		}
		return
	}

	response := convertGoalToResponse(restoredGoal)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangeGoalStatusRequest represents the request to change goal status
type ChangeGoalStatusRequest struct {
	Status string `json:"status" example:"active"`
}

// ChangeGoalStatusHandler changes the status of a goal
// @Summary Change goal status
// @Description Changes the status of a goal (active, deleted, etc.)
// @Tags goals
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Param status body ChangeGoalStatusRequest true "New status"
// @Success 200 {object} GoalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security bearerAuth
// @Router /api/v1/goals/{id}/status [patch]
func ChangeGoalStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Extract goal ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/goals/")
	goalID := strings.TrimSuffix(path, "/status")
	if goalID == "" || goalID == path {
		http.Error(w, "Goal ID is required", http.StatusBadRequest)
		return
	}

	var req ChangeGoalStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate status
	var newStatus models.Status
	switch req.Status {
	case "active":
		newStatus = models.StatusActive
	case "deleted":
		newStatus = models.StatusDeleted
	default:
		http.Error(w, "Invalid status. Must be 'active' or 'deleted'", http.StatusBadRequest)
		return
	}

	updatedGoal, err := services.ChangeGoalStatus(userID, goalID, newStatus)
	if err != nil {
		logger.Error("Error changing goal status: %v", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Goal not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error changing goal status", http.StatusInternalServerError)
		}
		return
	}

	response := convertGoalToResponse(updatedGoal)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// Request and response structures
type CreateBudgetRequest struct {
	MonthYear      string  `json:"month_year" example:"2024-01-01"`
	NeedsBudget    float64 `json:"needs_budget" example:"2500.00"`
	WantsBudget    float64 `json:"wants_budget" example:"1500.00"`
	SavingsBudget  float64 `json:"savings_budget" example:"1000.00"`
}

type UpdateBudgetRequest struct {
	NeedsBudget    *float64 `json:"needs_budget,omitempty" example:"2700.00"`
	WantsBudget    *float64 `json:"wants_budget,omitempty" example:"1300.00"`
	SavingsBudget  *float64 `json:"savings_budget,omitempty" example:"1200.00"`
	ChangeReason   *string  `json:"change_reason,omitempty" example:"Salary increase"`
}

type BudgetResponse struct {
	ID              string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	MonthYear       string  `json:"month_year" example:"2024-01-01"`
	NeedsBudget     float64 `json:"needs_budget" example:"2500.00"`
	WantsBudget     float64 `json:"wants_budget" example:"1500.00"`
	SavingsBudget   float64 `json:"savings_budget" example:"1000.00"`
	TotalBudget     float64 `json:"total_budget" example:"5000.00"`
	Status          string  `json:"status" example:"active"`
	StatusChangedAt *string `json:"status_changed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt       string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       string  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type BudgetsListResponse struct {
	Budgets []BudgetResponse `json:"budgets"`
	Count   int              `json:"count" example:"5"`
}

// Helper function to convert model to response
func convertBudgetToResponse(budget *models.Budget) BudgetResponse {
	response := BudgetResponse{
		ID:            budget.ID.String(),
		MonthYear:     budget.MonthYear.Format("2006-01-02"),
		NeedsBudget:   budget.NeedsBudget,
		WantsBudget:   budget.WantsBudget,
		SavingsBudget: budget.SavingsBudget,
		TotalBudget:   budget.NeedsBudget + budget.WantsBudget + budget.SavingsBudget,
		Status:        string(budget.Status),
		CreatedAt:     budget.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     budget.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	if budget.StatusChangedAt != nil {
		statusChangedAt := budget.StatusChangedAt.Format("2006-01-02T15:04:05Z07:00")
		response.StatusChangedAt = &statusChangedAt
	}
	
	return response
}

// CreateBudgetHandler godoc
// @Summary Create a new budget
// @Description Creates a new budget for the authenticated user for a specific month/year
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateBudgetRequest true "Budget data"
// @Success 201 {object} BudgetResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 409 {string} string "Budget already exists for this month/year"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets [post]
func CreateBudgetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get userID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validations
	if req.NeedsBudget < 0 || req.WantsBudget < 0 || req.SavingsBudget < 0 {
		http.Error(w, "Budget amounts must be positive", http.StatusBadRequest)
		return
	}

	if req.MonthYear == "" {
		http.Error(w, "Month year is required", http.StatusBadRequest)
		return
	}

	// Parse the month/year (first day of month)
	monthYear, err := parseMonthYear(req.MonthYear)
	if err != nil {
		http.Error(w, "Invalid month_year format, use YYYY-MM-DD (first day of month)", http.StatusBadRequest)
		return
	}

	// Create the model
	budget := &models.Budget{
		MonthYear:     monthYear,
		NeedsBudget:   req.NeedsBudget,
		WantsBudget:   req.WantsBudget,
		SavingsBudget: req.SavingsBudget,
	}

	// Create in the database
	if err := services.CreateBudget(userID, budget); err != nil {
		logger.Error("Error creating budget: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Error creating budget", http.StatusInternalServerError)
		}
		return
	}

	// Convert to response
	response := convertBudgetToResponse(budget)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetBudgetByIDHandler godoc
// @Summary Get a budget by ID
// @Description Gets a specific budget for the authenticated user by their ID
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} BudgetResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Budget not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/{id} [get]
func GetBudgetByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	id := extractIDFromPath(r.URL.Path, "/api/v1/budgets/")
	if id == "" {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	// Get the budget
	budget, err := services.GetBudgetByID(userID, id)
	if err != nil {
		logger.Error("Error getting budget: %v", err)
		http.Error(w, "Budget not found", http.StatusNotFound)
		return
	}

	response := convertBudgetToResponse(budget)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBudgetByMonthYearHandler godoc
// @Summary Get a budget by month/year
// @Description Gets the budget for the authenticated user for a specific month/year
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param month_year query string true "Month year (YYYY-MM-DD, first day of month)"
// @Success 200 {object} BudgetResponse
// @Failure 400 {string} string "Invalid month_year parameter"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Budget not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/by-month [get]
func GetBudgetByMonthYearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameter
	monthYearStr := r.URL.Query().Get("month_year")
	if monthYearStr == "" {
		http.Error(w, "month_year parameter is required", http.StatusBadRequest)
		return
	}

	monthYear, err := parseMonthYear(monthYearStr)
	if err != nil {
		http.Error(w, "Invalid month_year format, use YYYY-MM-DD (first day of month)", http.StatusBadRequest)
		return
	}

	// Get the budget
	budget, err := services.GetBudgetByMonthYear(userID, monthYear)
	if err != nil {
		logger.Error("Error getting budget by month/year: %v", err)
		http.Error(w, "Budget not found", http.StatusNotFound)
		return
	}

	response := convertBudgetToResponse(budget)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllBudgetsHandler godoc
// @Summary Get all budgets
// @Description Gets all budgets for the authenticated user with option to include deleted
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param include_deleted query boolean false "Include deleted budgets"
// @Success 200 {object} BudgetsListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets [get]
func GetAllBudgetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check parameter to include deleted
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	// Get budgets
	budgets, err := services.GetAllBudgets(userID, includeDeleted)
	if err != nil {
		logger.Error("Error getting budgets: %v", err)
		http.Error(w, "Error retrieving budgets", http.StatusInternalServerError)
		return
	}

	// Convert to response
	budgetResponses := make([]BudgetResponse, len(budgets))
	for i, budget := range budgets {
		budgetResponses[i] = convertBudgetToResponse(&budget)
	}

	response := BudgetsListResponse{
		Budgets: budgetResponses,
		Count:   len(budgetResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveBudgetsHandler godoc
// @Summary Get active budgets
// @Description Gets all active budgets for the authenticated user
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} BudgetsListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/active [get]
func GetActiveBudgetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	budgets, err := services.GetActiveBudgets(userID)
	if err != nil {
		logger.Error("Error getting active budgets: %v", err)
		http.Error(w, "Error retrieving active budgets", http.StatusInternalServerError)
		return
	}

	budgetResponses := make([]BudgetResponse, len(budgets))
	for i, budget := range budgets {
		budgetResponses[i] = convertBudgetToResponse(&budget)
	}

	response := BudgetsListResponse{
		Budgets: budgetResponses,
		Count:   len(budgetResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDeletedBudgetsHandler godoc
// @Summary Get deleted budgets
// @Description Gets all deleted budgets for the authenticated user
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} BudgetsListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/deleted [get]
func GetDeletedBudgetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	budgets, err := services.GetDeletedBudgets(userID)
	if err != nil {
		logger.Error("Error getting deleted budgets: %v", err)
		http.Error(w, "Error retrieving deleted budgets", http.StatusInternalServerError)
		return
	}

	budgetResponses := make([]BudgetResponse, len(budgets))
	for i, budget := range budgets {
		budgetResponses[i] = convertBudgetToResponse(&budget)
	}

	response := BudgetsListResponse{
		Budgets: budgetResponses,
		Count:   len(budgetResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateBudgetHandler godoc
// @Summary Update a budget
// @Description Updates partially a budget for the authenticated user (creates history entry)
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Budget ID"
// @Param request body UpdateBudgetRequest true "Data to update"
// @Success 200 {object} BudgetResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Budget not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/{id} [patch]
func UpdateBudgetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/budgets/")
	if id == "" {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	var req UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current budget to use as base for updates
	currentBudget, err := services.GetBudgetByID(userID, id)
	if err != nil {
		logger.Error("Error getting current budget: %v", err)
		http.Error(w, "Budget not found", http.StatusNotFound)
		return
	}

	// Create model with the fields to update (start with current values)
	budget := &models.Budget{
		NeedsBudget:   currentBudget.NeedsBudget,
		WantsBudget:   currentBudget.WantsBudget,
		SavingsBudget: currentBudget.SavingsBudget,
	}

	// Apply updates if provided
	if req.NeedsBudget != nil {
		if *req.NeedsBudget < 0 {
			http.Error(w, "Needs budget must be positive", http.StatusBadRequest)
			return
		}
		budget.NeedsBudget = *req.NeedsBudget
	}

	if req.WantsBudget != nil {
		if *req.WantsBudget < 0 {
			http.Error(w, "Wants budget must be positive", http.StatusBadRequest)
			return
		}
		budget.WantsBudget = *req.WantsBudget
	}

	if req.SavingsBudget != nil {
		if *req.SavingsBudget < 0 {
			http.Error(w, "Savings budget must be positive", http.StatusBadRequest)
			return
		}
		budget.SavingsBudget = *req.SavingsBudget
	}

	// Update in the database (creates history entry)
	updatedBudget, err := services.PatchBudget(userID, id, budget, req.ChangeReason)
	if err != nil {
		logger.Error("Error updating budget: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Budget not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error updating budget", http.StatusInternalServerError)
		}
		return
	}

	response := convertBudgetToResponse(updatedBudget)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteBudgetHandler godoc
// @Summary Delete a budget (soft delete)
// @Description Marks a budget as deleted without permanently deleting it
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} BudgetResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Budget not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/{id} [delete]
func DeleteBudgetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/budgets/")
	if id == "" {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	if err := services.SoftDeleteBudget(userID, id); err != nil {
		logger.Error("Error deleting budget: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already deleted") {
			http.Error(w, "Budget not found or already deleted", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting budget", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreBudgetHandler godoc
// @Summary Restore a deleted budget
// @Description Restores a previously deleted budget (soft delete)
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} BudgetResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Budget not found or not deleted"
// @Failure 409 {string} string "Another active budget exists for this month/year"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/{id}/restore [post]
func RestoreBudgetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/budgets/")
	if id == "" {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	restoredBudget, err := services.RestoreBudget(userID, id)
	if err != nil {
		logger.Error("Error restoring budget: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not deleted") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "another active budget exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Error restoring budget", http.StatusInternalServerError)
		}
		return
	}

	response := convertBudgetToResponse(restoredBudget)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ChangeBudgetStatusHandler godoc
// @Summary Change the status of a budget
// @Description Changes the status of a budget (active, inactive, deleted, etc.)
// @Tags budget
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Budget ID"
// @Param request body ChangeStatusRequest true "New status"
// @Success 200 {object} BudgetResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Budget not found"
// @Failure 409 {string} string "Another active budget exists for this month/year"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/budgets/{id}/status [patch]
func ChangeBudgetStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/budgets/")
	if id == "" {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	var req ChangeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	// Convert string to Status
	status := models.Status(req.Status)

	updatedBudget, err := services.ChangeBudgetStatus(userID, id, status, req.Reason)
	if err != nil {
		logger.Error("Error changing budget status: %v", err)
		if strings.Contains(err.Error(), "invalid status") {
			http.Error(w, "Invalid status", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Budget not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "another active budget exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Error changing status", http.StatusInternalServerError)
		}
		return
	}

	response := convertBudgetToResponse(updatedBudget)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions

// parseMonthYear parses a month/year in format YYYY-MM-DD (should be first day of month)
func parseMonthYear(monthYearStr string) (time.Time, error) {
	const layout = "2006-01-02"
	parsedTime, err := time.Parse(layout, monthYearStr)
	if err != nil {
		return time.Time{}, err
	}
	
	// Ensure it's the first day of the month
	return time.Date(parsedTime.Year(), parsedTime.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}



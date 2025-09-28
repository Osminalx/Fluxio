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
type CreateIncomeRequest struct {
	Amount float64 `json:"amount" example:"2500.50"`
	Date   string  `json:"date" example:"2024-01-15"`
}

type UpdateIncomeRequest struct {
	Amount *float64 `json:"amount,omitempty" example:"2800.75"`
	Date   *string  `json:"date,omitempty" example:"2024-01-16"`
}

type IncomeResponse struct {
	ID              string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Amount          float64 `json:"amount" example:"2500.50"`
	Date            string  `json:"date" example:"2024-01-15"`
	Status          string  `json:"status" example:"active"`
	StatusChangedAt *string `json:"status_changed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt       string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       string  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type IncomesListResponse struct {
	Incomes []IncomeResponse `json:"incomes"`
	Count   int              `json:"count" example:"5"`
}

// Helper function to convert model to response
func convertIncomeToResponse(income *models.Income) IncomeResponse {
	response := IncomeResponse{
		ID:        income.ID.String(),
		Amount:    income.Amount,
		Date:      income.Date.Format("2006-01-02"),
		Status:    string(income.Status),
		CreatedAt: income.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: income.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	if income.StatusChangedAt != nil {
		statusChangedAt := income.StatusChangedAt.Format("2006-01-02T15:04:05Z07:00")
		response.StatusChangedAt = &statusChangedAt
	}
	
	return response
}

// CreateIncomeHandler godoc
// @Summary Create a new income
// @Description Creates a new income for the authenticated user
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateIncomeRequest true "Income data"
// @Success 201 {object} IncomeResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes [post]
func CreateIncomeHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CreateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validations
	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.Date == "" {
		http.Error(w, "Date is required", http.StatusBadRequest)
		return
	}

	// Create the model
	income := &models.Income{
		Amount: req.Amount,
	}

	// Parse the date
	if date, err := parseDate(req.Date); err != nil {
		http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	} else {
		income.Date = date
	}

	// Create in the database
	if err := services.CreateIncome(userID, income); err != nil {
		logger.Error("Error creating income: %v", err)
		http.Error(w, "Error creating income", http.StatusInternalServerError)
		return
	}

	// Convert to response
	response := convertIncomeToResponse(income)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetIncomeByIDHandler godoc
// @Summary Get an income by ID
// @Description Gets a specific income for the authenticated user by their ID
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Income ID"
// @Success 200 {object} IncomeResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Income not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes/{id} [get]
func GetIncomeByIDHandler(w http.ResponseWriter, r *http.Request) {
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
	id := extractIDFromPath(r.URL.Path, "/api/v1/incomes/")
	if id == "" {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
		return
	}

	// Get the income
	income, err := services.GetIncomeByID(userID, id)
	if err != nil {
		logger.Error("Error getting income: %v", err)
		http.Error(w, "Income not found", http.StatusNotFound)
		return
	}

	response := convertIncomeToResponse(income)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllIncomesHandler godoc
// @Summary Get all incomes
// @Description Gets all incomes for the authenticated user with option to include deleted
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param include_deleted query boolean false "Include deleted incomes"
// @Success 200 {object} IncomesListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes [get]
func GetAllIncomesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get incomes
	incomes, err := services.GetAllIncomes(userID, includeDeleted)
	if err != nil {
		logger.Error("Error getting incomes: %v", err)
		http.Error(w, "Error retrieving incomes", http.StatusInternalServerError)
		return
	}

	// Convert to response
	incomeResponses := make([]IncomeResponse, len(incomes))
	for i, income := range incomes {
		incomeResponses[i] = convertIncomeToResponse(&income)
	}

	response := IncomesListResponse{
		Incomes: incomeResponses,
		Count:   len(incomeResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveIncomesHandler godoc
// @Summary Get active incomes
// @Description Gets all active incomes for the authenticated user
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} IncomesListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes/active [get]
func GetActiveIncomesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	incomes, err := services.GetActiveIncomes(userID)
	if err != nil {
		logger.Error("Error getting active incomes: %v", err)
		http.Error(w, "Error retrieving active incomes", http.StatusInternalServerError)
		return
	}

	incomeResponses := make([]IncomeResponse, len(incomes))
	for i, income := range incomes {
		incomeResponses[i] = convertIncomeToResponse(&income)
	}

	response := IncomesListResponse{
		Incomes: incomeResponses,
		Count:   len(incomeResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDeletedIncomesHandler godoc
// @Summary Get deleted incomes
// @Description Gets all deleted incomes for the authenticated user
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} IncomesListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes/deleted [get]
func GetDeletedIncomesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	incomes, err := services.GetDeletedIncomes(userID)
	if err != nil {
		logger.Error("Error getting deleted incomes: %v", err)
		http.Error(w, "Error retrieving deleted incomes", http.StatusInternalServerError)
		return
	}

	incomeResponses := make([]IncomeResponse, len(incomes))
	for i, income := range incomes {
		incomeResponses[i] = convertIncomeToResponse(&income)
	}

	response := IncomesListResponse{
		Incomes: incomeResponses,
		Count:   len(incomeResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateIncomeHandler godoc
// @Summary Update an income
// @Description Updates partially an income for the authenticated user
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Income ID"
// @Param request body UpdateIncomeRequest true "Data to update"
// @Success 200 {object} IncomeResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Income not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes/{id} [patch]
func UpdateIncomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/incomes/")
	if id == "" {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
		return
	}

	var req UpdateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create model with the fields to update
	income := &models.Income{}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
			return
		}
		income.Amount = *req.Amount
	}

	if req.Date != nil {
		if date, err := parseDate(*req.Date); err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		} else {
			income.Date = date
		}
	}

	// Update in the database
	updatedIncome, err := services.PatchIncome(userID, id, income)
	if err != nil {
		logger.Error("Error updating income: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Income not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error updating income", http.StatusInternalServerError)
		}
		return
	}

	response := convertIncomeToResponse(updatedIncome)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteIncomeHandler godoc
// @Summary Delete an income (soft delete)
// @Description Marks an income as deleted without permanently deleting it
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Income ID"
// @Success 200 {object} IncomeResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Income not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes/{id} [delete]
func DeleteIncomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/incomes/")
	if id == "" {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
		return
	}

	if err := services.SoftDeleteIncome(userID, id); err != nil {
		logger.Error("Error deleting income: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already deleted") {
			http.Error(w, "Income not found or already deleted", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting income", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreIncomeHandler godoc
// @Summary Restore a deleted income
// @Description Restores a previously deleted income (soft delete)
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Income ID"
// @Success 200 {object} IncomeResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Income not found or not deleted"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes/{id}/restore [post]
func RestoreIncomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/incomes/")
	if id == "" {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
		return
	}

	restoredIncome, err := services.RestoreIncome(userID, id)
	if err != nil {
		logger.Error("Error restoring income: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not deleted") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Income not found, not deleted, or access denied", http.StatusNotFound)
		} else {
			http.Error(w, "Error restoring income", http.StatusInternalServerError)
		}
		return
	}

	response := convertIncomeToResponse(restoredIncome)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ChangeIncomeStatusHandler godoc
// @Summary Change the status of an income
// @Description Changes the status of an income (active, inactive, deleted, etc.)
// @Tags income
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Income ID"
// @Param request body ChangeStatusRequest true "New status"
// @Success 200 {object} IncomeResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Income not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/incomes/{id}/status [patch]
func ChangeIncomeStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/incomes/")
	if id == "" {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
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

	updatedIncome, err := services.ChangeIncomeStatus(userID, id, status, req.Reason)
	if err != nil {
		logger.Error("Error changing income status: %v", err)
		if strings.Contains(err.Error(), "invalid status") {
			http.Error(w, "Invalid status", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Income not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error changing status", http.StatusInternalServerError)
		}
		return
	}

	response := convertIncomeToResponse(updatedIncome)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions



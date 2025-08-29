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
type CreateFixedExpenseRequest struct {
	Name    string  `json:"name" example:"Monthly Rent"`
	Amount  float64 `json:"amount" example:"1200.00"`
	DueDate string  `json:"due_date" example:"2024-01-15"`
}

type UpdateFixedExpenseRequest struct {
	Name    *string  `json:"name,omitempty" example:"Updated Rent"`
	Amount  *float64 `json:"amount,omitempty" example:"1300.00"`
	DueDate *string  `json:"due_date,omitempty" example:"2024-01-20"`
}

type FixedExpenseResponse struct {
	ID        string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name      string  `json:"name" example:"Monthly Rent"`
	Amount    float64 `json:"amount" example:"1200.00"`
	DueDate   string  `json:"due_date" example:"2024-01-15"`
	Status    string  `json:"status" example:"active"`
	CreatedAt string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type FixedExpensesListResponse struct {
	FixedExpenses []FixedExpenseResponse `json:"fixed_expenses"`
	Count         int                    `json:"count" example:"5"`
}

// Helper function to convert model to response
func convertFixedExpenseToResponse(fixedExpense *models.FixedExpense) FixedExpenseResponse {
	return FixedExpenseResponse{
		ID:        fixedExpense.ID.String(),
		Name:      fixedExpense.Name,
		Amount:    fixedExpense.Amount,
		DueDate:   fixedExpense.DueDate.Format("2006-01-02"),
		Status:    string(fixedExpense.Status),
		CreatedAt: fixedExpense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: fixedExpense.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateFixedExpenseHandler godoc
// @Summary Create a new fixed expense
// @Description Creates a new fixed expense for the authenticated user
// @Tags fixed_expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateFixedExpenseRequest true "Fixed expense data"
// @Success 201 {object} FixedExpenseResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/fixed-expenses [post]
func CreateFixedExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateFixedExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validations
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.DueDate == "" {
		http.Error(w, "Due date is required", http.StatusBadRequest)
		return
	}

	// Parse the due date
	dueDate, err := parseDate(req.DueDate)
	if err != nil {
		http.Error(w, "Invalid due date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Create the model
	fixedExpense := models.FixedExpense{
		Name:    req.Name,
		Amount:  req.Amount,
		DueDate: dueDate,
	}

	// Create in the database
	createdFixedExpense, err := services.CreateFixedExpense(userID, fixedExpense)
	if err != nil {
		logger.Error("Error creating fixed expense: %v", err)
		http.Error(w, "Error creating fixed expense", http.StatusInternalServerError)
		return
	}

	response := convertFixedExpenseToResponse(createdFixedExpense)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetFixedExpenseByIDHandler godoc
// @Summary Get a fixed expense by ID
// @Description Gets a specific fixed expense for the authenticated user by their ID
// @Tags fixed_expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Fixed Expense ID"
// @Success 200 {object} FixedExpenseResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Fixed expense not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/fixed-expenses/{id} [get]
func GetFixedExpenseByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/fixed-expenses/")
	if id == "" {
		http.Error(w, "Invalid fixed expense ID", http.StatusBadRequest)
		return
	}

	fixedExpense, err := services.GetFixedExpenseByID(userID, id)
	if err != nil {
		logger.Error("Error getting fixed expense: %v", err)
		http.Error(w, "Fixed expense not found", http.StatusNotFound)
		return
	}

	response := convertFixedExpenseToResponse(fixedExpense)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllFixedExpensesHandler godoc
// @Summary Get all fixed expenses
// @Description Gets all fixed expenses for the authenticated user with option to include deleted
// @Tags fixed_expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param include_deleted query boolean false "Include deleted fixed expenses"
// @Success 200 {object} FixedExpensesListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/fixed-expenses [get]
func GetAllFixedExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	fixedExpenses, err := services.GetFixedExpenses(userID, includeDeleted)
	if err != nil {
		logger.Error("Error getting fixed expenses: %v", err)
		http.Error(w, "Error retrieving fixed expenses", http.StatusInternalServerError)
		return
	}

	fixedExpenseResponses := make([]FixedExpenseResponse, len(fixedExpenses))
	for i, fixedExpense := range fixedExpenses {
		fixedExpenseResponses[i] = convertFixedExpenseToResponse(&fixedExpense)
	}

	response := FixedExpensesListResponse{
		FixedExpenses: fixedExpenseResponses,
		Count:         len(fixedExpenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateFixedExpenseHandler godoc
// @Summary Update a fixed expense
// @Description Updates partially a fixed expense for the authenticated user
// @Tags fixed_expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Fixed Expense ID"
// @Param request body UpdateFixedExpenseRequest true "Data to update"
// @Success 200 {object} FixedExpenseResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Fixed expense not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/fixed-expenses/{id} [patch]
func UpdateFixedExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/fixed-expenses/")
	if id == "" {
		http.Error(w, "Invalid fixed expense ID", http.StatusBadRequest)
		return
	}

	var req UpdateFixedExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current fixed expense for base values
	currentFixedExpense, err := services.GetFixedExpenseByID(userID, id)
	if err != nil {
		logger.Error("Error getting current fixed expense: %v", err)
		http.Error(w, "Fixed expense not found", http.StatusNotFound)
		return
	}

	// Create model with updates
	fixedExpense := models.FixedExpense{
		Name:    currentFixedExpense.Name,
		Amount:  currentFixedExpense.Amount,
		DueDate: currentFixedExpense.DueDate,
	}

	if req.Name != nil {
		if *req.Name == "" {
			http.Error(w, "Name cannot be empty", http.StatusBadRequest)
			return
		}
		fixedExpense.Name = *req.Name
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
			return
		}
		fixedExpense.Amount = *req.Amount
	}

	if req.DueDate != nil {
		if *req.DueDate == "" {
			http.Error(w, "Due date cannot be empty", http.StatusBadRequest)
			return
		}
		dueDate, err := parseDate(*req.DueDate)
		if err != nil {
			http.Error(w, "Invalid due date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		fixedExpense.DueDate = dueDate
	}

	// Update in the database
	updatedFixedExpense, err := services.UpdateFixedExpense(userID, id, fixedExpense)
	if err != nil {
		logger.Error("Error updating fixed expense: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "deleted") {
			http.Error(w, "Fixed expense not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error updating fixed expense", http.StatusInternalServerError)
		}
		return
	}

	response := convertFixedExpenseToResponse(updatedFixedExpense)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteFixedExpenseHandler godoc
// @Summary Delete a fixed expense (soft delete)
// @Description Marks a fixed expense as deleted without permanently deleting it
// @Tags fixed_expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Fixed Expense ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Fixed expense not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/fixed-expenses/{id} [delete]
func DeleteFixedExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/fixed-expenses/")
	if id == "" {
		http.Error(w, "Invalid fixed expense ID", http.StatusBadRequest)
		return
	}

	_, err := services.DeleteFixedExpense(userID, id)
	if err != nil {
		logger.Error("Error deleting fixed expense: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "deleted") {
			http.Error(w, "Fixed expense not found or already deleted", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting fixed expense", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



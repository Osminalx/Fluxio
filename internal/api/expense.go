package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

// Request and response structures
type CreateExpenseRequest struct {
	CategoryID      string  `json:"category_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Amount          float64 `json:"amount" example:"150.75"`
	Date            string  `json:"date" example:"2024-01-15"`
	BankAccountID   string  `json:"bank_account_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Description     *string `json:"description,omitempty" example:"Grocery shopping"`
}

type UpdateExpenseRequest struct {
	CategoryID      *string  `json:"category_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Amount          *float64 `json:"amount,omitempty" example:"175.50"`
	Date            *string  `json:"date,omitempty" example:"2024-01-16"`
	BankAccountID   *string  `json:"bank_account_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Description     *string  `json:"description,omitempty" example:"Updated description"`
}



type ExpenseResponse struct {
	ID              string             `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CategoryID      string             `json:"category_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Amount          float64            `json:"amount" example:"150.75"`
	Date            string             `json:"date" example:"2024-01-15"`
	BankAccountID   string             `json:"bank_account_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Description     *string            `json:"description,omitempty" example:"Grocery shopping"`
	Status          string             `json:"status" example:"active"`
	StatusChangedAt *string            `json:"status_changed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt       string             `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       string             `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	Category        *CategoryResponse  `json:"category,omitempty"`
	BankAccount     *BankAccountResponse `json:"bank_account,omitempty"`
}

type CategoryResponse struct {
	ID           string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name         string `json:"name" example:"Food"`
	ExpenseType  *ExpenseTypeResponse `json:"expense_type,omitempty"`
}

type ExpenseTypeResponse struct {
	ID   string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name string `json:"name" example:"Needs"`
}



type ExpensesListResponse struct {
	Expenses []ExpenseResponse `json:"expenses"`
	Count    int               `json:"count" example:"5"`
}

type ExpenseSummaryResponse struct {
	TotalAmount     float64                    `json:"total_amount" example:"1250.75"`
	TotalCount      int64                      `json:"total_count" example:"25"`
	AverageAmount   float64                    `json:"average_amount" example:"50.03"`
	ByExpenseType   []ExpensesByTypeResponse   `json:"by_expense_type"`
	TopCategories   []ExpensesByCategoryResponse `json:"top_categories"`
}

type ExpensesByTypeResponse struct {
	ExpenseTypeName string  `json:"expense_type_name" example:"Needs"`
	TotalAmount     float64 `json:"total_amount" example:"625.00"`
	Count           int64   `json:"count" example:"15"`
}

type ExpensesByCategoryResponse struct {
	CategoryName    string  `json:"category_name" example:"Food"`
	ExpenseTypeName string  `json:"expense_type_name" example:"Needs"`
	TotalAmount     float64 `json:"total_amount" example:"325.50"`
	Count           int64   `json:"count" example:"8"`
}

type DateRangeRequest struct {
	StartDate string `json:"start_date" example:"2024-01-01"`
	EndDate   string `json:"end_date" example:"2024-01-31"`
}

// Helper function to convert model to response
func convertExpenseToResponse(expense *models.Expense) ExpenseResponse {
	response := ExpenseResponse{
		ID:            expense.ID.String(),
		CategoryID:    expense.CategoryID.String(),
		Amount:        expense.Amount,
		Date:          expense.Date.Format("2006-01-02"),
		BankAccountID: expense.BankAccountID.String(),
		Description:   expense.Description,
		Status:        string(expense.Status),
		CreatedAt:     expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     expense.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	if expense.StatusChangedAt != nil {
		statusChangedAt := expense.StatusChangedAt.Format("2006-01-02T15:04:05Z07:00")
		response.StatusChangedAt = &statusChangedAt
	}
	
	// Include category information if loaded
	if expense.Category.ID != (uuid.UUID{}) {
		categoryResp := &CategoryResponse{
			ID:   expense.Category.ID.String(),
			Name: expense.Category.Name,
		}
		
		// Include expense type if loaded
		if expense.Category.ExpenseType.ID != (uuid.UUID{}) {
			categoryResp.ExpenseType = &ExpenseTypeResponse{
				ID:   expense.Category.ExpenseType.ID.String(),
				Name: expense.Category.ExpenseType.Name,
			}
		}
		
		response.Category = categoryResp
	}
	
	// Include bank account information if loaded
	if expense.BankAccount.ID != (uuid.UUID{}) {
		response.BankAccount = &BankAccountResponse{
			ID:          expense.BankAccount.ID.String(),
			AccountName: expense.BankAccount.AccountName,
			Balance:     expense.BankAccount.Balance,
		}
	}
	
	return response
}

// CreateExpenseHandler godoc
// @Summary Create a new expense
// @Description Creates a new expense for the authenticated user
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateExpenseRequest true "Expense data"
// @Success 201 {object} ExpenseResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses [post]
func CreateExpenseHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CreateExpenseRequest
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

	if req.CategoryID == "" || req.BankAccountID == "" || req.Date == "" {
		http.Error(w, "Category ID, Bank Account ID, and Date are required", http.StatusBadRequest)
		return
	}

	// Create the model
	expense := &models.Expense{
		Amount:      req.Amount,
		Description: req.Description,
	}

	// Parse UUIDs
	if categoryUUID, err := uuid.Parse(req.CategoryID); err != nil {
		http.Error(w, "Invalid category ID format", http.StatusBadRequest)
		return
	} else {
		expense.CategoryID = categoryUUID
	}

	if bankAccountUUID, err := uuid.Parse(req.BankAccountID); err != nil {
		http.Error(w, "Invalid bank account ID format", http.StatusBadRequest)
		return
	} else {
		expense.BankAccountID = bankAccountUUID
	}

	// Parse the date
	if date, err := parseDate(req.Date); err != nil {
		http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	} else {
		expense.Date = date
	}

	// Create in the database
	if err := services.CreateExpense(userID, expense); err != nil {
		logger.Error("Error creating expense: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not active") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Error creating expense", http.StatusInternalServerError)
		}
		return
	}

	// Get the created expense with relations
	createdExpense, err := services.GetExpenseByID(userID, expense.ID.String())
	if err != nil {
		// If we can't get the full expense, return the basic one
		createdExpense = expense
	}

	// Convert to response
	response := convertExpenseToResponse(createdExpense)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetExpenseByIDHandler godoc
// @Summary Get an expense by ID
// @Description Gets a specific expense for the authenticated user by their ID
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Expense ID"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Expense not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/{id} [get]
func GetExpenseByIDHandler(w http.ResponseWriter, r *http.Request) {
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
	id := extractIDFromPath(r.URL.Path, "/api/v1/expenses/")
	if id == "" {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	// Get the expense
	expense, err := services.GetExpenseByID(userID, id)
	if err != nil {
		logger.Error("Error getting expense: %v", err)
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	response := convertExpenseToResponse(expense)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllExpensesHandler godoc
// @Summary Get all expenses
// @Description Gets all expenses for the authenticated user with option to include deleted
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param include_deleted query boolean false "Include deleted expenses"
// @Success 200 {object} ExpensesListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses [get]
func GetAllExpensesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get expenses
	expenses, err := services.GetAllExpenses(userID, includeDeleted)
	if err != nil {
		logger.Error("Error getting expenses: %v", err)
		http.Error(w, "Error retrieving expenses", http.StatusInternalServerError)
		return
	}

	// Convert to response
	expenseResponses := make([]ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = convertExpenseToResponse(&expense)
	}

	response := ExpensesListResponse{
		Expenses: expenseResponses,
		Count:    len(expenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveExpensesHandler godoc
// @Summary Get active expenses
// @Description Gets all active expenses for the authenticated user
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} ExpensesListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/active [get]
func GetActiveExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expenses, err := services.GetActiveExpenses(userID)
	if err != nil {
		logger.Error("Error getting active expenses: %v", err)
		http.Error(w, "Error retrieving active expenses", http.StatusInternalServerError)
		return
	}

	expenseResponses := make([]ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = convertExpenseToResponse(&expense)
	}

	response := ExpensesListResponse{
		Expenses: expenseResponses,
		Count:    len(expenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDeletedExpensesHandler godoc
// @Summary Get deleted expenses
// @Description Gets all deleted expenses for the authenticated user
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} ExpensesListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/deleted [get]
func GetDeletedExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expenses, err := services.GetDeletedExpenses(userID)
	if err != nil {
		logger.Error("Error getting deleted expenses: %v", err)
		http.Error(w, "Error retrieving deleted expenses", http.StatusInternalServerError)
		return
	}

	expenseResponses := make([]ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = convertExpenseToResponse(&expense)
	}

	response := ExpensesListResponse{
		Expenses: expenseResponses,
		Count:    len(expenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateExpenseHandler godoc
// @Summary Update an expense
// @Description Updates partially an expense for the authenticated user
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Expense ID"
// @Param request body UpdateExpenseRequest true "Data to update"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Expense not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/{id} [patch]
func UpdateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/expenses/")
	if id == "" {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	var req UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create model with the fields to update
	expense := &models.Expense{}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
			return
		}
		expense.Amount = *req.Amount
	}

	if req.CategoryID != nil {
		if categoryUUID, err := uuid.Parse(*req.CategoryID); err != nil {
			http.Error(w, "Invalid category ID format", http.StatusBadRequest)
			return
		} else {
			expense.CategoryID = categoryUUID
		}
	}

	if req.BankAccountID != nil {
		if bankAccountUUID, err := uuid.Parse(*req.BankAccountID); err != nil {
			http.Error(w, "Invalid bank account ID format", http.StatusBadRequest)
			return
		} else {
			expense.BankAccountID = bankAccountUUID
		}
	}

	if req.Date != nil {
		if date, err := parseDate(*req.Date); err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		} else {
			expense.Date = date
		}
	}

	if req.Description != nil {
		expense.Description = req.Description
	}

	// Update in the database
	updatedExpense, err := services.PatchExpense(userID, id, expense)
	if err != nil {
		logger.Error("Error updating expense: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Expense not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "not active") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Error updating expense", http.StatusInternalServerError)
		}
		return
	}

	response := convertExpenseToResponse(updatedExpense)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteExpenseHandler godoc
// @Summary Delete an expense (soft delete)
// @Description Marks an expense as deleted without permanently deleting it
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Expense ID"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Expense not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/{id} [delete]
func DeleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/expenses/")
	if id == "" {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	if err := services.SoftDeleteExpense(userID, id); err != nil {
		logger.Error("Error deleting expense: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already deleted") {
			http.Error(w, "Expense not found or already deleted", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting expense", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreExpenseHandler godoc
// @Summary Restore a deleted expense
// @Description Restores a previously deleted expense (soft delete)
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Expense ID"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Expense not found or not deleted"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/{id}/restore [post]
func RestoreExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/expenses/")
	if id == "" {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	restoredExpense, err := services.RestoreExpense(userID, id)
	if err != nil {
		logger.Error("Error restoring expense: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not deleted") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "not active") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Error restoring expense", http.StatusInternalServerError)
		}
		return
	}

	response := convertExpenseToResponse(restoredExpense)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ChangeExpenseStatusHandler godoc
// @Summary Change the status of an expense
// @Description Changes the status of an expense (active, inactive, deleted, etc.)
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Expense ID"
// @Param request body ChangeStatusRequest true "New status"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Expense not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/{id}/status [patch]
func ChangeExpenseStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/expenses/")
	if id == "" {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
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

	updatedExpense, err := services.ChangeExpenseStatus(userID, id, status, req.Reason)
	if err != nil {
		logger.Error("Error changing expense status: %v", err)
		if strings.Contains(err.Error(), "invalid status") {
			http.Error(w, "Invalid status", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Expense not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error changing status", http.StatusInternalServerError)
		}
		return
	}

	response := convertExpenseToResponse(updatedExpense)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// === ADDITIONAL ENDPOINTS FOR SPECIALIZED QUERIES ===

// GetExpensesByDateRangeHandler godoc
// @Summary Get expenses by date range
// @Description Gets expenses for the authenticated user within a date range
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param include_deleted query boolean false "Include deleted expenses"
// @Success 200 {object} ExpensesListResponse
// @Failure 400 {string} string "Invalid date parameters"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/date-range [get]
func GetExpensesByDateRangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date and end_date parameters are required", http.StatusBadRequest)
		return
	}

	startDate, err := parseDate(startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := parseDate(endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	expenses, err := services.GetExpensesByDateRange(userID, startDate, endDate, includeDeleted)
	if err != nil {
		logger.Error("Error getting expenses by date range: %v", err)
		http.Error(w, "Error retrieving expenses", http.StatusInternalServerError)
		return
	}

	expenseResponses := make([]ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = convertExpenseToResponse(&expense)
	}

	response := ExpensesListResponse{
		Expenses: expenseResponses,
		Count:    len(expenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExpensesByCategoryHandler godoc
// @Summary Get expenses by category
// @Description Gets expenses for the authenticated user filtered by category
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param category_id path string true "Category ID"
// @Param include_deleted query boolean false "Include deleted expenses"
// @Success 200 {object} ExpensesListResponse
// @Failure 400 {string} string "Invalid category ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/category/{category_id} [get]
func GetExpensesByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	categoryID := extractIDFromPath(r.URL.Path, "/api/v1/expenses/category/")
	if categoryID == "" {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	expenses, err := services.GetExpensesByCategory(userID, categoryID, includeDeleted)
	if err != nil {
		logger.Error("Error getting expenses by category: %v", err)
		http.Error(w, "Error retrieving expenses", http.StatusInternalServerError)
		return
	}

	expenseResponses := make([]ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = convertExpenseToResponse(&expense)
	}

	response := ExpensesListResponse{
		Expenses: expenseResponses,
		Count:    len(expenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExpensesByBankAccountHandler godoc
// @Summary Get expenses by bank account
// @Description Gets expenses for the authenticated user filtered by bank account
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param bank_account_id path string true "Bank Account ID"
// @Param include_deleted query boolean false "Include deleted expenses"
// @Success 200 {object} ExpensesListResponse
// @Failure 400 {string} string "Invalid bank account ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/bank-account/{bank_account_id} [get]
func GetExpensesByBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bankAccountID := extractIDFromPath(r.URL.Path, "/api/v1/expenses/bank-account/")
	if bankAccountID == "" {
		http.Error(w, "Invalid bank account ID", http.StatusBadRequest)
		return
	}

	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	expenses, err := services.GetExpensesByBankAccount(userID, bankAccountID, includeDeleted)
	if err != nil {
		logger.Error("Error getting expenses by bank account: %v", err)
		http.Error(w, "Error retrieving expenses", http.StatusInternalServerError)
		return
	}

	expenseResponses := make([]ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = convertExpenseToResponse(&expense)
	}

	response := ExpensesListResponse{
		Expenses: expenseResponses,
		Count:    len(expenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMonthlyExpensesHandler godoc
// @Summary Get monthly expenses
// @Description Gets expenses for the authenticated user for a specific month
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param year query int true "Year (e.g., 2024)"
// @Param month query int true "Month (1-12)"
// @Param include_deleted query boolean false "Include deleted expenses"
// @Success 200 {object} ExpensesListResponse
// @Failure 400 {string} string "Invalid year or month parameters"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/monthly [get]
func GetMonthlyExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	if yearStr == "" || monthStr == "" {
		http.Error(w, "year and month parameters are required", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		http.Error(w, "Invalid year, must be between 2000 and 2100", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		http.Error(w, "Invalid month, must be between 1 and 12", http.StatusBadRequest)
		return
	}

	expenses, err := services.GetMonthlyExpenses(userID, year, month, includeDeleted)
	if err != nil {
		logger.Error("Error getting monthly expenses: %v", err)
		http.Error(w, "Error retrieving expenses", http.StatusInternalServerError)
		return
	}

	expenseResponses := make([]ExpenseResponse, len(expenses))
	for i, expense := range expenses {
		expenseResponses[i] = convertExpenseToResponse(&expense)
	}

	response := ExpensesListResponse{
		Expenses: expenseResponses,
		Count:    len(expenseResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExpensesSummaryHandler godoc
// @Summary Get expenses summary
// @Description Gets expenses summary for the authenticated user within a date range
// @Tags expense
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} ExpenseSummaryResponse
// @Failure 400 {string} string "Invalid date parameters"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/expenses/summary [get]
func GetExpensesSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date and end_date parameters are required", http.StatusBadRequest)
		return
	}

	startDate, err := parseDate(startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := parseDate(endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	summary, err := services.GetExpensesSummaryByPeriod(userID, startDate, endDate)
	if err != nil {
		logger.Error("Error getting expenses summary: %v", err)
		http.Error(w, "Error retrieving summary", http.StatusInternalServerError)
		return
	}

	// Convert the map response to structured response
	response := ExpenseSummaryResponse{
		TotalAmount:   summary["total_amount"].(float64),
		TotalCount:    summary["total_count"].(int64),
		AverageAmount: summary["average_amount"].(float64),
	}

	// Convert by expense type
	if byExpenseType, ok := summary["by_expense_type"].([]struct {
		ExpenseTypeName string  `json:"expense_type_name"`
		TotalAmount     float64 `json:"total_amount"`
		Count           int64   `json:"count"`
	}); ok {
		response.ByExpenseType = make([]ExpensesByTypeResponse, len(byExpenseType))
		for i, item := range byExpenseType {
			response.ByExpenseType[i] = ExpensesByTypeResponse{
				ExpenseTypeName: item.ExpenseTypeName,
				TotalAmount:     item.TotalAmount,
				Count:           item.Count,
			}
		}
	}

	// Convert top categories
	if topCategories, ok := summary["top_categories"].([]struct {
		CategoryName    string  `json:"category_name"`
		ExpenseTypeName string  `json:"expense_type_name"`
		TotalAmount     float64 `json:"total_amount"`
		Count           int64   `json:"count"`
	}); ok {
		response.TopCategories = make([]ExpensesByCategoryResponse, len(topCategories))
		for i, item := range topCategories {
			response.TopCategories[i] = ExpensesByCategoryResponse{
				CategoryName:    item.CategoryName,
				ExpenseTypeName: item.ExpenseTypeName,
				TotalAmount:     item.TotalAmount,
				Count:           item.Count,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}



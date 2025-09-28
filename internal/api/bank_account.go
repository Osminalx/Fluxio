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
type CreateBankAccountRequest struct {
	AccountName string  `json:"account_name" example:"Main Checking Account"`
	Balance     float64 `json:"balance" example:"2500.00"`
}

type UpdateBankAccountRequest struct {
	AccountName *string  `json:"account_name,omitempty" example:"Updated Account Name"`
	Balance     *float64 `json:"balance,omitempty" example:"3000.00"`
}

type BankAccountFullResponse struct {
	ID              string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AccountName     string  `json:"account_name" example:"Main Checking Account"`
	Balance         float64 `json:"balance" example:"2500.00"`
	Status          string  `json:"status" example:"active"`
	StatusChangedAt *string `json:"status_changed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt       string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       string  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type BankAccountsListResponse struct {
	BankAccounts []BankAccountFullResponse `json:"bank_accounts"`
	Count        int                       `json:"count" example:"3"`
}

// Helper function to convert model to response
func convertBankAccountToResponse(bankAccount *models.BankAccount) BankAccountFullResponse {
	response := BankAccountFullResponse{
		ID:          bankAccount.ID.String(),
		AccountName: bankAccount.AccountName,
		Balance:     bankAccount.Balance,
		Status:      string(bankAccount.Status),
		CreatedAt:   bankAccount.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   bankAccount.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	if bankAccount.StatusChangedAt != nil {
		statusChangedAt := bankAccount.StatusChangedAt.Format("2006-01-02T15:04:05Z07:00")
		response.StatusChangedAt = &statusChangedAt
	}
	
	return response
}

// CreateBankAccountHandler godoc
// @Summary Create a new bank account
// @Description Creates a new bank account for the authenticated user
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateBankAccountRequest true "Bank account data"
// @Success 201 {object} BankAccountFullResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts [post]
func CreateBankAccountHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CreateBankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validations
	if req.AccountName == "" {
		http.Error(w, "Account name is required", http.StatusBadRequest)
		return
	}

	if req.Balance < 0 {
		http.Error(w, "Balance cannot be negative", http.StatusBadRequest)
		return
	}

	// Create the model
	bankAccount := &models.BankAccount{
		AccountName: req.AccountName,
		Balance:     req.Balance,
	}

	// Create in the database
	if err := services.CreateBankAccount(userID, bankAccount); err != nil {
		logger.Error("Error creating bank account: %v", err)
		http.Error(w, "Error creating bank account", http.StatusInternalServerError)
		return
	}

	// Convert to response
	response := convertBankAccountToResponse(bankAccount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetBankAccountByIDHandler godoc
// @Summary Get a bank account by ID
// @Description Gets a specific bank account for the authenticated user by their ID
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Bank Account ID"
// @Success 200 {object} BankAccountResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Bank account not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts/{id} [get]
func GetBankAccountByIDHandler(w http.ResponseWriter, r *http.Request) {
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
	id := extractIDFromPath(r.URL.Path, "/api/v1/bank-accounts/")
	if id == "" {
		http.Error(w, "Invalid bank account ID", http.StatusBadRequest)
		return
	}

	// Get the bank account
	bankAccount, err := services.GetBankAccountByID(userID, id)
	if err != nil {
		logger.Error("Error getting bank account: %v", err)
		http.Error(w, "Bank account not found", http.StatusNotFound)
		return
	}

	response := convertBankAccountToResponse(bankAccount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllBankAccountsHandler godoc
// @Summary Get all bank accounts
// @Description Gets all bank accounts for the authenticated user with option to include deleted
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param include_deleted query boolean false "Include deleted bank accounts"
// @Success 200 {object} BankAccountsListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts [get]
func GetAllBankAccountsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get bank accounts
	bankAccounts, err := services.GetAllBankAccounts(userID, includeDeleted)
	if err != nil {
		logger.Error("Error getting bank accounts: %v", err)
		http.Error(w, "Error retrieving bank accounts", http.StatusInternalServerError)
		return
	}

	// Convert to response
	bankAccountResponses := make([]BankAccountFullResponse, len(bankAccounts))
	for i, bankAccount := range bankAccounts {
		bankAccountResponses[i] = convertBankAccountToResponse(&bankAccount)
	}

	response := BankAccountsListResponse{
		BankAccounts: bankAccountResponses,
		Count:        len(bankAccountResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveBankAccountsHandler godoc
// @Summary Get active bank accounts
// @Description Gets all active bank accounts for the authenticated user
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} BankAccountsListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts/active [get]
func GetActiveBankAccountsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bankAccounts, err := services.GetActiveBankAccounts(userID)
	if err != nil {
		logger.Error("Error getting active bank accounts: %v", err)
		http.Error(w, "Error retrieving active bank accounts", http.StatusInternalServerError)
		return
	}

	bankAccountResponses := make([]BankAccountFullResponse, len(bankAccounts))
	for i := range bankAccounts {
		bankAccountResponses[i] = convertBankAccountToResponse(&bankAccounts[i])
	}

	response := BankAccountsListResponse{
		BankAccounts: bankAccountResponses,
		Count:        len(bankAccountResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDeletedBankAccountsHandler godoc
// @Summary Get deleted bank accounts
// @Description Gets all deleted bank accounts for the authenticated user
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} BankAccountsListResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts/deleted [get]
func GetDeletedBankAccountsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bankAccounts, err := services.GetDeletedBankAccounts(userID)
	if err != nil {
		logger.Error("Error getting deleted bank accounts: %v", err)
		http.Error(w, "Error retrieving deleted bank accounts", http.StatusInternalServerError)
		return
	}

	bankAccountResponses := make([]BankAccountFullResponse, len(bankAccounts))
	for i := range bankAccounts {
		bankAccountResponses[i] = convertBankAccountToResponse(&bankAccounts[i])
	}

	response := BankAccountsListResponse{
		BankAccounts: bankAccountResponses,
		Count:        len(bankAccountResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateBankAccountHandler godoc
// @Summary Update a bank account
// @Description Updates partially a bank account for the authenticated user
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Bank Account ID"
// @Param request body UpdateBankAccountRequest true "Data to update"
// @Success 200 {object} BankAccountResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Bank account not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts/{id} [patch]
func UpdateBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/bank-accounts/")
	if id == "" {
		http.Error(w, "Invalid bank account ID", http.StatusBadRequest)
		return
	}

	var req UpdateBankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current bank account to use as base for updates
	currentBankAccount, err := services.GetBankAccountByID(userID, id)
	if err != nil {
		logger.Error("Error getting current bank account: %v", err)
		http.Error(w, "Bank account not found", http.StatusNotFound)
		return
	}

	// Create model with the fields to update (start with current values)
	bankAccount := &models.BankAccount{
		AccountName: currentBankAccount.AccountName,
		Balance:     currentBankAccount.Balance,
	}

	// Apply updates if provided
	if req.AccountName != nil {
		if *req.AccountName == "" {
			http.Error(w, "Account name cannot be empty", http.StatusBadRequest)
			return
		}
		bankAccount.AccountName = *req.AccountName
	}

	if req.Balance != nil {
		if *req.Balance < 0 {
			http.Error(w, "Balance cannot be negative", http.StatusBadRequest)
			return
		}
		bankAccount.Balance = *req.Balance
	}

	// Update in the database
	updatedBankAccount, err := services.PatchBankAccount(userID, id, bankAccount)
	if err != nil {
		logger.Error("Error updating bank account: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Bank account not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error updating bank account", http.StatusInternalServerError)
		}
		return
	}

	response := convertBankAccountToResponse(updatedBankAccount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteBankAccountHandler godoc
// @Summary Delete a bank account (soft delete)
// @Description Marks a bank account as deleted without permanently deleting it
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Bank Account ID"
// @Success 200 {object} BankAccountResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Bank account not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts/{id} [delete]
func DeleteBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/bank-accounts/")
	if id == "" {
		http.Error(w, "Invalid bank account ID", http.StatusBadRequest)
		return
	}

	if err := services.SoftDeleteBankAccount(userID, id); err != nil {
		logger.Error("Error deleting bank account: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already deleted") {
			http.Error(w, "Bank account not found or already deleted", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting bank account", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreBankAccountHandler godoc
// @Summary Restore a bank account to active status
// @Description Restores a previously deleted, archived, or locked bank account to active status
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Bank Account ID"
// @Success 200 {object} BankAccountResponse
// @Failure 400 {string} string "Invalid ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Bank account not found or not restorable"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts/{id}/restore [post]
func RestoreBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/bank-accounts/")
	if id == "" {
		http.Error(w, "Invalid bank account ID", http.StatusBadRequest)
		return
	}

	restoredAccount, err := services.RestoreBankAccount(userID, id)
	if err != nil {
		logger.Error("Error restoring bank account: %v", err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not restorable") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error restoring bank account", http.StatusInternalServerError)
		}
		return
	}

	response := convertBankAccountToResponse(restoredAccount)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ChangeBankAccountStatusHandler godoc
// @Summary Change the status of a bank account
// @Description Changes the status of a bank account (active, inactive, deleted, etc.) and returns the updated account
// @Tags bank_account
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Bank Account ID"
// @Param request body ChangeStatusRequest true "New status"
// @Success 200 {object} BankAccountFullResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Bank account not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/bank-accounts/{id}/status [patch]
func ChangeBankAccountStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := extractIDFromPath(r.URL.Path, "/api/v1/bank-accounts/")
	if id == "" {
		http.Error(w, "Invalid bank account ID", http.StatusBadRequest)
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

	if err := services.ChangeAccountStatus(userID, id, status, req.Reason); err != nil {
		logger.Error("Error changing bank account status: %v", err)
		if strings.Contains(err.Error(), "invalid status") {
			http.Error(w, "Invalid status", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, "Bank account not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error changing status", http.StatusInternalServerError)
		}
		return
	}

	// Get the updated bank account to return to the frontend
	updatedBankAccount, err := services.GetBankAccountByID(userID, id)
	if err != nil {
		logger.Error("Error retrieving updated bank account: %v", err)
		http.Error(w, "Error retrieving updated bank account", http.StatusInternalServerError)
		return
	}

	response := convertBankAccountToResponse(updatedBankAccount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}




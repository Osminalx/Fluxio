package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

// CreateTransferRequest represents the request body for creating a transfer
type CreateTransferRequest struct {
	FromAccountID uuid.UUID `json:"from_account_id" validate:"required"`
	ToAccountID   uuid.UUID `json:"to_account_id" validate:"required"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	Description   *string   `json:"description,omitempty"`
	Date          time.Time `json:"date" validate:"required"`
}

// UpdateTransferRequest represents the request body for updating a transfer
type UpdateTransferRequest struct {
	Amount      *float64   `json:"amount,omitempty"`
	Description *string    `json:"description,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
}

// CreateTransferHandler godoc
// @Summary Create a new transfer
// @Description Creates a new transfer between bank accounts for the authenticated user
// @Tags transfers
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateTransferRequest true "Transfer data"
// @Success 201 {object} models.Transfer
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/transfers [post]
func CreateTransferHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding transfer request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FromAccountID == uuid.Nil {
		http.Error(w, "Source account ID is required", http.StatusBadRequest)
		return
	}

	if req.ToAccountID == uuid.Nil {
		http.Error(w, "Destination account ID is required", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
		return
	}

	if req.Date.IsZero() {
		http.Error(w, "Date is required", http.StatusBadRequest)
		return
	}

	transferService := services.NewTransferService()
	transfer, err := transferService.CreateTransfer(userID, req.FromAccountID, req.ToAccountID, req.Amount, req.Description, req.Date)
	if err != nil {
		logger.Error("Error creating transfer: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("Transfer created successfully: %s", transfer.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transfer)
}

// GetAllTransfersHandler godoc
// @Summary Get all transfers for user
// @Description Retrieves all transfers for the authenticated user with optional filtering
// @Tags transfers
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Param account_id query string false "Filter by account ID (source or destination)"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {array} models.Transfer
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/transfers [get]
func GetAllTransfersHandler(w http.ResponseWriter, r *http.Request) {
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
	accountIDStr := r.URL.Query().Get("account_id")
	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	var accountID *uuid.UUID
	if accountIDStr != "" {
		if id, err := uuid.Parse(accountIDStr); err == nil {
			accountID = &id
		}
	}

	var fromDate, toDate *time.Time
	if fromDateStr != "" {
		if date, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			fromDate = &date
		}
	}
	if toDateStr != "" {
		if date, err := time.Parse("2006-01-02", toDateStr); err == nil {
			toDate = &date
		}
	}

	transferService := services.NewTransferService()
	transfers, err := transferService.GetUserTransfers(userID, accountID, nil, fromDate, toDate, limit, offset)
	if err != nil {
		logger.Error("Error retrieving transfers: %v", err)
		http.Error(w, "Error retrieving transfers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfers)
}

// GetTransferByIDHandler godoc
// @Summary Get transfer by ID
// @Description Retrieves a specific transfer by ID for the authenticated user
// @Tags transfers
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Transfer ID"
// @Success 200 {object} models.Transfer
// @Failure 400 {string} string "Invalid transfer ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Transfer not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/transfers/{id} [get]
func GetTransferByIDHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract transfer ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	transferIDStr := pathParts[len(pathParts)-1]
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		logger.Error("Invalid transfer ID format: %v", err)
		http.Error(w, "Invalid transfer ID", http.StatusBadRequest)
		return
	}

	transferService := services.NewTransferService()
	transfer, err := transferService.GetTransferByID(transferID, userID)
	if err != nil {
		if err.Error() == "transfer not found" {
			http.Error(w, "Transfer not found", http.StatusNotFound)
			return
		}
		logger.Error("Error retrieving transfer: %v", err)
		http.Error(w, "Error retrieving transfer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfer)
}

// UpdateTransferHandler godoc
// @Summary Update transfer
// @Description Updates a transfer for the authenticated user
// @Tags transfers
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Transfer ID"
// @Param request body UpdateTransferRequest true "Update data"
// @Success 200 {object} models.Transfer
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Transfer not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/transfers/{id} [patch]
func UpdateTransferHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract transfer ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	transferIDStr := pathParts[len(pathParts)-1]
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		logger.Error("Invalid transfer ID format: %v", err)
		http.Error(w, "Invalid transfer ID", http.StatusBadRequest)
		return
	}

	var req UpdateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding update transfer request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transferService := services.NewTransferService()
	
	// Build updates map
	updates := make(map[string]interface{})
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Date != nil {
		updates["date"] = *req.Date
	}
	
	transfer, err := transferService.UpdateTransfer(userID, transferID, updates)
	if err != nil {
		if err.Error() == "transfer not found" {
			http.Error(w, "Transfer not found", http.StatusNotFound)
			return
		}
		logger.Error("Error updating transfer: %v", err)
		http.Error(w, "Error updating transfer", http.StatusInternalServerError)
		return
	}

	logger.Info("Transfer updated successfully: %s", transfer.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfer)
}

// DeleteTransferHandler godoc
// @Summary Delete transfer
// @Description Soft deletes a transfer for the authenticated user
// @Tags transfers
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Transfer ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid transfer ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Transfer not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/transfers/{id} [delete]
func DeleteTransferHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract transfer ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	transferIDStr := pathParts[len(pathParts)-1]
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		logger.Error("Invalid transfer ID format: %v", err)
		http.Error(w, "Invalid transfer ID", http.StatusBadRequest)
		return
	}

	transferService := services.NewTransferService()
	err = transferService.DeleteTransfer(userID, transferID)
	if err != nil {
		if err.Error() == "transfer not found" {
			http.Error(w, "Transfer not found", http.StatusNotFound)
			return
		}
		logger.Error("Error deleting transfer: %v", err)
		http.Error(w, "Error deleting transfer", http.StatusInternalServerError)
		return
	}

	logger.Info("Transfer deleted successfully: %s", transferID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transfer deleted successfully",
	})
}

// GetTransfersByAccountHandler godoc
// @Summary Get transfers by account
// @Description Retrieves all transfers for a specific bank account
// @Tags transfers
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param account_id path string true "Bank Account ID"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} models.Transfer
// @Failure 400 {string} string "Invalid account ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/transfers/account/{account_id} [get]
func GetTransfersByAccountHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract account ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	accountIDStr := pathParts[len(pathParts)-1]
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		logger.Error("Invalid account ID format: %v", err)
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	transferService := services.NewTransferService()
	// Use GetUserTransfers with account filter
	transfers, err := transferService.GetUserTransfers(userID, &accountID, &accountID, nil, nil, limit, offset)
	if err != nil {
		logger.Error("Error retrieving transfers by account: %v", err)
		http.Error(w, "Error retrieving transfers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfers)
}

// GetTransferStatsHandler godoc
// @Summary Get transfer statistics
// @Description Retrieves transfer statistics for the authenticated user
// @Tags transfers
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param account_id query string false "Filter by account ID"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/transfers/stats [get]
func GetTransferStatsHandler(w http.ResponseWriter, r *http.Request) {
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
	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	var fromDate, toDate *time.Time
	if fromDateStr != "" {
		if date, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			fromDate = &date
		}
	}
	if toDateStr != "" {
		if date, err := time.Parse("2006-01-02", toDateStr); err == nil {
			toDate = &date
		}
	}

	transferService := services.NewTransferService()
	stats, err := transferService.GetTransferStats(userID, fromDate, toDate)
	if err != nil {
		logger.Error("Error retrieving transfer stats: %v", err)
		http.Error(w, "Error retrieving transfer stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

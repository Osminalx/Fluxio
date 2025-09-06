package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransferService struct {
	db *gorm.DB
}

func NewTransferService() *TransferService {
	return &TransferService{
		db: db.DB,
	}
}

// CreateTransfer creates a new transfer between bank accounts
func (s *TransferService) CreateTransfer(userID, fromAccountID, toAccountID uuid.UUID, amount float64, description *string, date time.Time) (*models.Transfer, error) {
	// Validate amount
	if amount <= 0 {
		return nil, errors.New("transfer amount must be greater than zero")
	}

	// Validate that accounts are different
	if fromAccountID == toAccountID {
		return nil, errors.New("source and destination accounts must be different")
	}

	// Verify both accounts belong to the user and are active
	var fromAccount, toAccount models.BankAccount
	if err := s.db.Where("id = ? AND user_id = ? AND status = ?", fromAccountID, userID, models.StatusActive).First(&fromAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("source account not found or not accessible")
		}
		return nil, err
	}

	if err := s.db.Where("id = ? AND user_id = ? AND status = ?", toAccountID, userID, models.StatusActive).First(&toAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("destination account not found or not accessible")
		}
		return nil, err
	}

	// Check if source account has sufficient balance
	if fromAccount.Balance < amount {
		return nil, errors.New("insufficient balance in source account")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create transfer record
	transfer := &models.Transfer{
		ID:            uuid.New(),
		UserID:        userID,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Description:   description,
		Date:          date,
		Status:        models.StatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update account balances
	if err := tx.Model(&fromAccount).Update("balance", fromAccount.Balance-amount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Model(&toAccount).Update("balance", toAccount.Balance+amount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Load relationships for response
	if err := s.db.Preload("FromAccount").Preload("ToAccount").First(transfer, transfer.ID).Error; err != nil {
		return nil, err
	}

	return transfer, nil
}

// GetTransferByID retrieves a transfer by ID for a specific user
func (s *TransferService) GetTransferByID(userID, transferID uuid.UUID) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := s.db.Preload("FromAccount").Preload("ToAccount").
		Where("id = ? AND user_id = ? AND status IN ?", transferID, userID, models.GetActiveStatuses()).
		First(&transfer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transfer not found")
		}
		return nil, err
	}

	return &transfer, nil
}

// GetUserTransfers retrieves all transfers for a user with filters
func (s *TransferService) GetUserTransfers(userID uuid.UUID, fromAccountID, toAccountID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*models.Transfer, error) {
	query := s.db.Preload("FromAccount").Preload("ToAccount").
		Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses())

	// Filter by from account
	if fromAccountID != nil {
		query = query.Where("from_account_id = ?", *fromAccountID)
	}

	// Filter by to account
	if toAccountID != nil {
		query = query.Where("to_account_id = ?", *toAccountID)
	}

	// Filter by date range
	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	// Order by date (most recent first)
	query = query.Order("date DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var transfers []*models.Transfer
	if err := query.Find(&transfers).Error; err != nil {
		return nil, err
	}

	return transfers, nil
}

// GetAccountTransfers retrieves all transfers involving a specific account
func (s *TransferService) GetAccountTransfers(userID, accountID uuid.UUID, limit, offset int) ([]*models.Transfer, error) {
	query := s.db.Preload("FromAccount").Preload("ToAccount").
		Where("user_id = ? AND (from_account_id = ? OR to_account_id = ?) AND status IN ?", 
			userID, accountID, accountID, models.GetActiveStatuses()).
		Order("date DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var transfers []*models.Transfer
	if err := query.Find(&transfers).Error; err != nil {
		return nil, err
	}

	return transfers, nil
}

// UpdateTransfer updates a transfer's information (limited fields)
func (s *TransferService) UpdateTransfer(userID, transferID uuid.UUID, updates map[string]interface{}) (*models.Transfer, error) {
	// Get existing transfer
	transfer, err := s.GetTransferByID(userID, transferID)
	if err != nil {
		return nil, err
	}

	// Only allow updating description and date for now
	allowedFields := map[string]bool{
		"description": true,
		"date":        true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return transfer, nil
	}

	// Add updated_at timestamp
	filteredUpdates["updated_at"] = time.Now()

	// Update transfer
	if err := s.db.Model(transfer).Updates(filteredUpdates).Error; err != nil {
		return nil, err
	}

	// Return updated transfer
	return s.GetTransferByID(userID, transferID)
}

// DeleteTransfer soft deletes a transfer and reverses the balance changes
func (s *TransferService) DeleteTransfer(userID, transferID uuid.UUID) error {
	// Get the transfer to be deleted
	transfer, err := s.GetTransferByID(userID, transferID)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current account balances
	var fromAccount, toAccount models.BankAccount
	if err := tx.Where("id = ? AND user_id = ?", transfer.FromAccountID, userID).First(&fromAccount).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ? AND user_id = ?", transfer.ToAccountID, userID).First(&toAccount).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Check if destination account has sufficient balance to reverse the transfer
	if toAccount.Balance < transfer.Amount {
		tx.Rollback()
		return errors.New("cannot delete transfer: insufficient balance in destination account to reverse the transaction")
	}

	// Reverse the balance changes
	if err := tx.Model(&fromAccount).Update("balance", fromAccount.Balance+transfer.Amount).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&toAccount).Update("balance", toAccount.Balance-transfer.Amount).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Soft delete the transfer
	updates := map[string]interface{}{
		"status":            models.StatusDeleted,
		"status_changed_at": time.Now(),
		"updated_at":        time.Now(),
	}

	if err := tx.Model(transfer).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// GetTransferStats returns statistics about user's transfers
func (s *TransferService) GetTransferStats(userID uuid.UUID, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	query := s.db.Model(&models.Transfer{}).Where("user_id = ? AND status = ?", userID, models.StatusActive)

	// Apply date filters if provided
	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	// Total number of transfers
	var totalCount int64
	query.Count(&totalCount)
	stats["total_transfers"] = totalCount

	// Total amount transferred
	var totalAmount float64
	query.Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)
	stats["total_amount"] = totalAmount

	// Average transfer amount
	if totalCount > 0 {
		stats["average_amount"] = totalAmount / float64(totalCount)
	} else {
		stats["average_amount"] = 0.0
	}

	// Transfers by month (if no date range specified, get last 12 months)
	if startDate == nil && endDate == nil {
		now := time.Now()
		lastYear := now.AddDate(-1, 0, 0)
		
		var monthlyStats []map[string]interface{}
		rows, err := s.db.Raw(`
			SELECT 
				DATE_TRUNC('month', date) as month,
				COUNT(*) as count,
				COALESCE(SUM(amount), 0) as total_amount
			FROM transfers 
			WHERE user_id = ? AND status = ? AND date >= ?
			GROUP BY DATE_TRUNC('month', date)
			ORDER BY month DESC
		`, userID, models.StatusActive, lastYear).Rows()
		
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var month time.Time
				var count int64
				var amount float64
				if err := rows.Scan(&month, &count, &amount); err == nil {
					monthlyStats = append(monthlyStats, map[string]interface{}{
						"month":        month.Format("2006-01"),
						"count":        count,
						"total_amount": amount,
					})
				}
			}
		}
		stats["monthly_breakdown"] = monthlyStats
	}

	return stats, nil
}

// GetRecentTransfers retrieves the most recent transfers for a user
func (s *TransferService) GetRecentTransfers(userID uuid.UUID, limit int) ([]*models.Transfer, error) {
	if limit <= 0 {
		limit = 10
	}

	var transfers []*models.Transfer
	if err := s.db.Preload("FromAccount").Preload("ToAccount").
		Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses()).
		Order("created_at DESC").
		Limit(limit).
		Find(&transfers).Error; err != nil {
		return nil, err
	}

	return transfers, nil
}

// ValidateTransferPossibility checks if a transfer between two accounts is possible
func (s *TransferService) ValidateTransferPossibility(userID, fromAccountID, toAccountID uuid.UUID, amount float64) error {
	// Validate amount
	if amount <= 0 {
		return errors.New("transfer amount must be greater than zero")
	}

	// Validate that accounts are different
	if fromAccountID == toAccountID {
		return errors.New("source and destination accounts must be different")
	}

	// Verify both accounts belong to the user and are active
	var fromAccount, toAccount models.BankAccount
	if err := s.db.Where("id = ? AND user_id = ? AND status = ?", fromAccountID, userID, models.StatusActive).First(&fromAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("source account not found or not accessible")
		}
		return err
	}

	if err := s.db.Where("id = ? AND user_id = ? AND status = ?", toAccountID, userID, models.StatusActive).First(&toAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("destination account not found or not accessible")
		}
		return err
	}

	// Check if source account has sufficient balance
	if fromAccount.Balance < amount {
		return errors.New("insufficient balance in source account")
	}

	return nil
}

package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

func CreateBankAccount(userID string, bankAccount *models.BankAccount) error {
	// Force the UserID and Status to prevent manipulation
	bankAccount.UserID = uuid.MustParse(userID)
	bankAccount.Status = models.StatusActive

	result := db.DB.Create(bankAccount)
	if result.Error != nil{
		logger.Error("Error creating bank account: %v", result.Error)
		return result.Error
	}
	logger.Info("Bank account created successfully: %+v", bankAccount)
	return nil
}

func GetBankAccountByID(userID string, id string) (*models.BankAccount, error) {
	var bankAccount models.BankAccount
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&bankAccount)
	if result.Error != nil{
		logger.Error("Error getting bank account by id: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("Bank account retrieved successfully: %+v", bankAccount)
	return &bankAccount, nil
}

func GetAllBankAccounts(userID string, includeDeleted bool) ([]models.BankAccount, error){
	var bankAccounts []models.BankAccount
	query := db.DB.Where("user_id = ?", userID)
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("created_at DESC").Find(&bankAccounts)
	if result.Error != nil{
		logger.Error("Error getting all bank accounts: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("All bank accounts retrieved successfully: %+v", bankAccounts)
	return bankAccounts, nil
}

func GetActiveBankAccounts(userID string) ([]models.BankAccount, error){
	var bankAccounts []models.BankAccount
	result := db.DB.Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses()).
		Order("created_at DESC").Find(&bankAccounts)
	if result.Error != nil{
		logger.Error("Error getting active bank accounts: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("Active bank accounts retrieved successfully: %+v", bankAccounts)
	return bankAccounts, nil
}

func GetDeletedBankAccounts(userID string) ([]models.BankAccount, error){
	var bankAccounts []models.BankAccount
	result := db.DB.Where("user_id = ? AND status = ?", userID, models.StatusDeleted).
		Order("status_changed_at DESC").Find(&bankAccounts)
	if result.Error != nil{
		logger.Error("Error getting deleted bank accounts: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("Deleted bank accounts retrieved successfully: %+v", bankAccounts)
	return bankAccounts, nil
}

func PatchBankAccount(userID string, id string, bankAccount *models.BankAccount) (*models.BankAccount, error) {
	var existingAccount models.BankAccount
	
	// Check if the account exists, belongs to the user and is not deleted
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&existingAccount)
	if result.Error != nil{
		logger.Error("Bank account not found or doesn't belong to the user: %v", result.Error)
		return nil, errors.New("bank account not found or access denied")
	}
	
	// Prevent modification of protected fields
	bankAccount.UserID = existingAccount.UserID
	bankAccount.ID = existingAccount.ID
	bankAccount.CreatedAt = existingAccount.CreatedAt
	
	// Do not allow status change through normal patch (use specific functions)
	bankAccount.Status = existingAccount.Status
	bankAccount.StatusChangedAt = existingAccount.StatusChangedAt
	
	// Update only if the account belongs to the user
	result = db.DB.Model(&existingAccount).Where("user_id = ? AND id = ?", userID, id).Updates(bankAccount)
	if result.Error != nil{
		logger.Error("Error patching bank account: %v", result.Error)
		return nil, result.Error
	}
	
	// Get the updated account
	result = db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingAccount)
	if result.Error != nil{
		logger.Error("Error retrieving updated bank account: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Bank account patched successfully: %+v", existingAccount)
	return &existingAccount, nil
}

func SoftDeleteBankAccount(userID string, id string) error {
	// Check if the account exists and belongs to the user
	var existingAccount models.BankAccount
	result := db.DB.Where("user_id = ? AND id = ? AND status != ?", userID, id, models.StatusDeleted).First(&existingAccount)
	if result.Error != nil {
		logger.Error("Bank account not found or already deleted: %v", result.Error)
		return errors.New("bank account not found or already deleted")
	}
	
	// Mark as deleted
	now := time.Now()
	result = db.DB.Model(&existingAccount).Updates(map[string]interface{}{
		"status": models.StatusDeleted,
		"status_changed_at": &now,
	})
	
	if result.Error != nil{
		logger.Error("Error soft deleting bank account: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Bank account soft deleted successfully: %s", id)
	return nil
}

func RestoreBankAccount(userID string, id string) error {
	// Check if the account exists, belongs to the user and is deleted
	var existingAccount models.BankAccount
	result := db.DB.Where("user_id = ? AND id = ? AND status = ?", userID, id, models.StatusDeleted).First(&existingAccount)
	if result.Error != nil {
		logger.Error("Bank account not found, not deleted, or access denied: %v", result.Error)
		return errors.New("bank account not found, not deleted, or access denied")
	}
	
	// Restore as active
	now := time.Now()
	result = db.DB.Model(&existingAccount).Updates(map[string]interface{}{
		"status": models.StatusActive,
		"status_changed_at": &now,
	})
	
	if result.Error != nil{
		logger.Error("Error restoring bank account: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Bank account restored successfully: %s", id)
	return nil
}

func ChangeAccountStatus(userID string, id string, newStatus models.Status, reason *string) error {
	// Validate that the status is valid
	if !models.ValidateStatus(newStatus) {
		return errors.New("invalid status")
	}
	
	// Check if the account exists and belongs to the user
	var existingAccount models.BankAccount
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingAccount)
	if result.Error != nil {
		logger.Error("Bank account not found: %v", result.Error)
		return errors.New("bank account not found or access denied")
	}
	
	// Do nothing if it already has that status
	if existingAccount.Status == newStatus {
		return nil
	}
	
	// Update status
	now := time.Now()
	updates := map[string]interface{}{
		"status": newStatus,
		"status_changed_at": &now,
	}
	
	result = db.DB.Model(&existingAccount).Updates(updates)
	if result.Error != nil{
		logger.Error("Error changing bank account status: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Bank account status changed to %s successfully: %s", newStatus, id)
	return nil
}

func HardDeleteBankAccount(userID string, id string) error {
	// Only for special cases - permanently delete
	// Check if the account exists and belongs to the user
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&models.BankAccount{})
	if result.Error != nil{
		logger.Error("Error hard deleting bank account: %v", result.Error)
		return result.Error
	}
	
	if result.RowsAffected == 0{
		logger.Error("Bank account not found or doesn't belong to user")
		return errors.New("bank account not found or access denied")
	}

	logger.Info("Bank account permanently deleted: %s", id)
	return nil
}
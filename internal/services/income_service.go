package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateIncome(userID string, income *models.Income) error {
	// Forzar el UserID y Status para que no puedan ser manipulados
	income.UserID = uuid.MustParse(userID)
	income.Status = models.StatusActive
	
	// Verify that the bank account exists, is active and belongs to the user
	var bankAccount models.BankAccount
	result := db.DB.Where("id = ? AND user_id = ? AND status IN ?", 
		income.BankAccountID, userID, models.GetActiveStatuses()).First(&bankAccount)
	if result.Error != nil {
		logger.Error("Bank account not found, not active, or doesn't belong to user")
		return errors.New("bank account not found, not active, or access denied")
	}
	
	// Verify that the amount is positive
	if income.Amount <= 0 {
		logger.Error("Income amount must be positive")
		return errors.New("income amount must be positive")
	}
	
	result = db.DB.Create(income)
	if result.Error != nil{
		logger.Error("Error creating income: %v", result.Error)
		return result.Error
	}
	
	// Add income to bank account balance
	if err := db.DB.Model(&bankAccount).
		Update("balance", gorm.Expr("balance + ?", income.Amount)).Error; err != nil {
		logger.Error("Error updating bank account balance: %v", err)
		return errors.New("error updating bank account balance")
	}
	
	logger.Info("Income created successfully: %+v", income)
	return nil
}

func GetIncomeByID(userID string, id string) (*models.Income, error) {
	var income models.Income
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&income)
	if result.Error != nil{
		logger.Error("Error getting income by id: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("Income retrieved successfully: %+v", income)
	return &income, nil
}

func GetAllIncomes(userID string, includeDeleted bool) ([]models.Income, error) {
	var incomes []models.Income
	query := db.DB.Where("user_id = ?", userID)
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("date DESC, created_at DESC").Find(&incomes)
	if result.Error != nil{
		logger.Error("Error getting all incomes: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("All incomes retrieved successfully: %+v", incomes)
	return incomes, nil
}

func GetActiveIncomes(userID string) ([]models.Income, error) {
	var incomes []models.Income
	result := db.DB.Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses()).
		Order("date DESC, created_at DESC").Find(&incomes)
	if result.Error != nil{
		logger.Error("Error getting active incomes: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("Active incomes retrieved successfully: %+v", incomes)
	return incomes, nil
}

func GetDeletedIncomes(userID string) ([]models.Income, error) {
	var incomes []models.Income
	result := db.DB.Where("user_id = ? AND status = ?", userID, models.StatusDeleted).
		Order("status_changed_at DESC").Find(&incomes)
	if result.Error != nil{
		logger.Error("Error getting deleted incomes: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("Deleted incomes retrieved successfully: %+v", incomes)
	return incomes, nil
}

func PatchIncome(userID string, id string, income *models.Income) (*models.Income, error) {
	var existingIncome models.Income
	
	// Verificar que el income existe, pertenece al usuario y no está eliminado
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&existingIncome)
	if result.Error != nil {
		logger.Error("Income not found or doesn't belong to user: %v", result.Error)
		return nil, errors.New("income not found or access denied")
	}
	
	// Prevenir modificación de campos protegidos
	income.UserID = existingIncome.UserID
	income.ID = existingIncome.ID
	income.CreatedAt = existingIncome.CreatedAt
	
	// No permitir cambio de status a través de patch normal (usar funciones específicas)
	income.Status = existingIncome.Status
	income.StatusChangedAt = existingIncome.StatusChangedAt
	
	// Actualizar solo si pertenece al usuario
	result = db.DB.Model(&existingIncome).Where("user_id = ? AND id = ?", userID, id).Updates(income)
	if result.Error != nil{
		logger.Error("Error patching income: %v", result.Error)
		return nil, result.Error
	}
	
	if result.RowsAffected == 0{
		logger.Error("Income not found or doesn't belong to user")
		return nil, errors.New("income not found or access denied")
	}
	
	// Obtener el income actualizado
	result = db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingIncome)
	if result.Error != nil {
		logger.Error("Error retrieving updated income: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Income patched successfully: %+v", existingIncome)
	return &existingIncome, nil
}

func SoftDeleteIncome(userID string, id string) error {
	// Verificar que el income existe y pertenece al usuario
	var existingIncome models.Income
	result := db.DB.Where("user_id = ? AND id = ? AND status != ?", userID, id, models.StatusDeleted).First(&existingIncome)
	if result.Error != nil {
		logger.Error("Income not found or already deleted: %v", result.Error)
		return errors.New("income not found or already deleted")
	}
	
	// Marcar como eliminado
	now := time.Now()
	result = db.DB.Model(&existingIncome).Updates(map[string]interface{}{
		"status": models.StatusDeleted,
		"status_changed_at": &now,
	})
	
	if result.Error != nil{
		logger.Error("Error soft deleting income: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Income soft deleted successfully: %s", id)
	return nil
}

func RestoreIncome(userID string, id string) (*models.Income, error) {
	// Verificar que el income existe, pertenece al usuario y está eliminado
	var existingIncome models.Income
	result := db.DB.Where("user_id = ? AND id = ? AND status = ?", userID, id, models.StatusDeleted).First(&existingIncome)
	if result.Error != nil {
		logger.Error("Income not found, not deleted, or access denied: %v", result.Error)
		return nil, errors.New("income not found, not deleted, or access denied")
	}
	
	// Restaurar como activo
	now := time.Now()
	result = db.DB.Model(&existingIncome).Updates(map[string]interface{}{
		"status": models.StatusActive,
		"status_changed_at": &now,
	})
	
	if result.Error != nil{
		logger.Error("Error restoring income: %v", result.Error)
		return nil, result.Error
	}
	
	// Get the updated income
	updatedIncome, err := GetIncomeByID(userID, id)
	if err != nil {
		logger.Error("Error retrieving updated income: %v", err)
		return nil, errors.New("error retrieving updated income")
	}
	
	logger.Info("Income restored successfully: %s", id)
	return updatedIncome, nil
}

func ChangeIncomeStatus(userID string, id string, newStatus models.Status, reason *string) (*models.Income, error) {
	// Validar que el status es válido
	if !models.ValidateStatus(newStatus) {
		return nil, errors.New("invalid status")
	}
	
	// Verificar que el income existe y pertenece al usuario
	var existingIncome models.Income
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingIncome)
	if result.Error != nil {
		logger.Error("Income not found: %v", result.Error)
		return nil, errors.New("income not found or access denied")
	}
	
	// No hacer nada si ya tiene ese status - return current income
	if existingIncome.Status == newStatus {
		updatedIncome, err := GetIncomeByID(userID, id)
		if err != nil {
			logger.Error("Error retrieving income: %v", err)
			return nil, errors.New("error retrieving income")
		}
		return updatedIncome, nil
	}
	
	// Actualizar status
	now := time.Now()
	updates := map[string]interface{}{
		"status": newStatus,
		"status_changed_at": &now,
	}
	
	result = db.DB.Model(&existingIncome).Updates(updates)
	if result.Error != nil{
		logger.Error("Error changing income status: %v", result.Error)
		return nil, result.Error
	}
	
	// Get the updated income
	updatedIncome, err := GetIncomeByID(userID, id)
	if err != nil {
		logger.Error("Error retrieving updated income: %v", err)
		return nil, errors.New("error retrieving updated income")
	}
	
	logger.Info("Income status changed to %s successfully: %s", newStatus, id)
	return updatedIncome, nil
}

func HardDeleteIncome(userID string, id string) error {
	// SOLO para casos especiales - elimina permanentemente
	// Verificar que el income existe y pertenece al usuario
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Income{})
	if result.Error != nil{
		logger.Error("Error hard deleting income: %v", result.Error)
		return result.Error
	}
	
	// Verificar que realmente se eliminó algo
	if result.RowsAffected == 0 {
		logger.Error("Income not found or doesn't belong to user")
		return errors.New("income not found or access denied")
	}
	
	logger.Info("Income permanently deleted: %s", id)
	return nil
}
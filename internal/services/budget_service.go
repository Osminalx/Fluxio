package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

// CreateBudget creates a new budget for the user
func CreateBudget(userID string, budget *models.Budget) error {
	// Forzar el UserID y Status para que no puedan ser manipulados
	budget.UserID = uuid.MustParse(userID)
	budget.Status = models.StatusActive
	
	// Verificar que no existe un presupuesto activo para ese mes/año
	var existingBudget models.Budget
	result := db.DB.Where("user_id = ? AND month_year = ? AND status IN ?", userID, budget.MonthYear, models.GetActiveStatuses()).First(&existingBudget)
	if result.Error == nil {
		logger.Error("Active budget already exists for this month/year")
		return errors.New("active budget already exists for this month and year")
	}
	
	result = db.DB.Create(budget)
	if result.Error != nil {
		logger.Error("Error creating budget: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Budget created successfully: %+v", budget)
	return nil
}

// GetBudgetByID gets a specific budget for the user
func GetBudgetByID(userID string, id string) (*models.Budget, error) {
	var budget models.Budget
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&budget)
	if result.Error != nil {
		logger.Error("Error getting budget by id: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Budget retrieved successfully: %+v", budget)
	return &budget, nil
}

// GetBudgetByMonthYear gets the budget for a specific month/year
func GetBudgetByMonthYear(userID string, monthYear time.Time) (*models.Budget, error) {
	var budget models.Budget
	result := db.DB.Where("user_id = ? AND month_year = ? AND status IN ?", userID, monthYear, models.GetVisibleStatuses()).First(&budget)
	if result.Error != nil {
		logger.Error("Error getting budget by month/year: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Budget retrieved successfully by month/year: %+v", budget)
	return &budget, nil
}

// GetActiveBudgetByMonthYear gets the active budget for a specific month/year
func GetActiveBudgetByMonthYear(userID string, monthYear time.Time) (*models.Budget, error) {
	var budget models.Budget
	result := db.DB.Where("user_id = ? AND month_year = ? AND status IN ?", userID, monthYear, models.GetActiveStatuses()).First(&budget)
	if result.Error != nil {
		logger.Error("Error getting active budget by month/year: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Active budget retrieved successfully by month/year: %+v", budget)
	return &budget, nil
}

// GetAllBudgets gets all the budgets for the user
func GetAllBudgets(userID string, includeDeleted bool) ([]models.Budget, error) {
	var budgets []models.Budget
	query := db.DB.Where("user_id = ?", userID)
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("month_year DESC").Find(&budgets)
	if result.Error != nil {
		logger.Error("Error getting all budgets: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("All budgets retrieved successfully: %+v", budgets)
	return budgets, nil
}

// GetActiveBudgets gets all active budgets for the user
func GetActiveBudgets(userID string) ([]models.Budget, error) {
	var budgets []models.Budget
	result := db.DB.Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses()).
		Order("month_year DESC").Find(&budgets)
	if result.Error != nil {
		logger.Error("Error getting active budgets: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Active budgets retrieved successfully: %+v", budgets)
	return budgets, nil
}

// GetDeletedBudgets gets all deleted budgets for the user
func GetDeletedBudgets(userID string) ([]models.Budget, error) {
	var budgets []models.Budget
	result := db.DB.Where("user_id = ? AND status = ?", userID, models.StatusDeleted).
		Order("status_changed_at DESC").Find(&budgets)
	if result.Error != nil {
		logger.Error("Error getting deleted budgets: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Deleted budgets retrieved successfully: %+v", budgets)
	return budgets, nil
}

// PatchBudget updates a budget and automatically creates the history
func PatchBudget(userID string, id string, updatedBudget *models.Budget, changeReason *string) (*models.Budget, error) {
	var existingBudget models.Budget
	
	// Verificar que el presupuesto existe, pertenece al usuario y no está eliminado
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&existingBudget)
	if result.Error != nil {
		logger.Error("Budget not found or doesn't belong to user: %v", result.Error)
		return nil, errors.New("budget not found or access denied")
	}
	
	// Create an entry in BudgetHistory before updating
	historyEntry := &models.BudgetHistory{
		BudgetID:         existingBudget.ID,
		UserID:           existingBudget.UserID,
		OldNeedsBudget:   &existingBudget.NeedsBudget,
		OldWantsBudget:   &existingBudget.WantsBudget,
		OldSavingsBudget: &existingBudget.SavingsBudget,
		NewNeedsBudget:   &updatedBudget.NeedsBudget,
		NewWantsBudget:   &updatedBudget.WantsBudget,
		NewSavingsBudget: &updatedBudget.SavingsBudget,
		ChangedAt:        time.Now(),
		ChangeReason:     changeReason,
	}
	
	// Start a transaction to ensure consistency
	tx := db.DB.Begin()
	if tx.Error != nil {
		logger.Error("Error starting transaction: %v", tx.Error)
		return nil, tx.Error
	}
	
	// Create an entry in the history
	if err := tx.Create(historyEntry).Error; err != nil {
		tx.Rollback()
		logger.Error("Error creating budget history: %v", err)
		return nil, err
	}
	
	// Prevenir modificación de campos protegidos
	updatedBudget.UserID = existingBudget.UserID
	updatedBudget.ID = existingBudget.ID
	updatedBudget.CreatedAt = existingBudget.CreatedAt
	
	// No permitir cambio de status a través de patch normal (usar funciones específicas)
	updatedBudget.Status = existingBudget.Status
	updatedBudget.StatusChangedAt = existingBudget.StatusChangedAt
	
	// Update the budget
	if err := tx.Model(&existingBudget).Where("user_id = ? AND id = ?", userID, id).Updates(updatedBudget).Error; err != nil {
		tx.Rollback()
		logger.Error("Error updating budget: %v", err)
		return nil, err
	}
	
	// Confirm the transaction
	if err := tx.Commit().Error; err != nil {
		logger.Error("Error committing transaction: %v", err)
		return nil, err
	}
	
	// Get the updated budget
	result = db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingBudget)
	if result.Error != nil {
		logger.Error("Error retrieving updated budget: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Budget updated successfully with history: %+v", existingBudget)
	return &existingBudget, nil
}

// SoftDeleteBudget marks a budget as deleted for the user
func SoftDeleteBudget(userID string, id string) error {
	// Verificar que el presupuesto existe y pertenece al usuario
	var existingBudget models.Budget
	result := db.DB.Where("user_id = ? AND id = ? AND status != ?", userID, id, models.StatusDeleted).First(&existingBudget)
	if result.Error != nil {
		logger.Error("Budget not found or already deleted: %v", result.Error)
		return errors.New("budget not found or already deleted")
	}
	
	// Marcar como eliminado
	now := time.Now()
	result = db.DB.Model(&existingBudget).Updates(map[string]interface{}{
		"status": models.StatusDeleted,
		"status_changed_at": &now,
	})
	
	if result.Error != nil{
		logger.Error("Error soft deleting budget: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Budget soft deleted successfully: %s", id)
	return nil
}

// RestoreBudget restores a deleted budget for the user
func RestoreBudget(userID string, id string) error {
	// Verificar que el presupuesto existe, pertenece al usuario y está eliminado
	var existingBudget models.Budget
	result := db.DB.Where("user_id = ? AND id = ? AND status = ?", userID, id, models.StatusDeleted).First(&existingBudget)
	if result.Error != nil {
		logger.Error("Budget not found, not deleted, or access denied: %v", result.Error)
		return errors.New("budget not found, not deleted, or access denied")
	}
	
	// Verificar que no existe otro presupuesto activo para ese mes/año
	var activeBudget models.Budget
	checkResult := db.DB.Where("user_id = ? AND month_year = ? AND status IN ? AND id != ?", 
		userID, existingBudget.MonthYear, models.GetActiveStatuses(), id).First(&activeBudget)
	if checkResult.Error == nil {
		logger.Error("Another active budget exists for this month/year")
		return errors.New("cannot restore: another active budget exists for this month/year")
	}
	
	// Restaurar como activo
	now := time.Now()
	result = db.DB.Model(&existingBudget).Updates(map[string]interface{}{
		"status": models.StatusActive,
		"status_changed_at": &now,
	})
	
	if result.Error != nil{
		logger.Error("Error restoring budget: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Budget restored successfully: %s", id)
	return nil
}

// ChangeBudgetStatus changes the status of a budget
func ChangeBudgetStatus(userID string, id string, newStatus models.Status, reason *string) error {
	// Validar que el status es válido
	if !models.ValidateStatus(newStatus) {
		return errors.New("invalid status")
	}
	
	// Verificar que el presupuesto existe y pertenece al usuario
	var existingBudget models.Budget
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingBudget)
	if result.Error != nil {
		logger.Error("Budget not found: %v", result.Error)
		return errors.New("budget not found or access denied")
	}
	
	// No hacer nada si ya tiene ese status
	if existingBudget.Status == newStatus {
		return nil
	}
	
	// Si se está activando, verificar que no haya otro activo para ese mes/año
	if newStatus == models.StatusActive {
		var activeBudget models.Budget
		checkResult := db.DB.Where("user_id = ? AND month_year = ? AND status IN ? AND id != ?", 
			userID, existingBudget.MonthYear, models.GetActiveStatuses(), id).First(&activeBudget)
		if checkResult.Error == nil {
			logger.Error("Another active budget exists for this month/year")
			return errors.New("cannot activate: another active budget exists for this month/year")
		}
	}
	
	// Actualizar status
	now := time.Now()
	updates := map[string]interface{}{
		"status": newStatus,
		"status_changed_at": &now,
	}
	
	result = db.DB.Model(&existingBudget).Updates(updates)
	if result.Error != nil{
		logger.Error("Error changing budget status: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Budget status changed to %s successfully: %s", newStatus, id)
	return nil
}

// HardDeleteBudget permanently deletes a budget for the user
func HardDeleteBudget(userID string, id string) error {
	// SOLO para casos especiales - elimina permanentemente
	// Verificar que el presupuesto existe y pertenece al usuario
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Budget{})
	if result.Error != nil {
		logger.Error("Error hard deleting budget: %v", result.Error)
		return result.Error
	}
	
	// Verificar que realmente se eliminó algo
	if result.RowsAffected == 0 {
		logger.Error("Budget not found or doesn't belong to user")
		return errors.New("budget not found or access denied")
	}
	
	logger.Info("Budget permanently deleted: %s", id)
	return nil
}

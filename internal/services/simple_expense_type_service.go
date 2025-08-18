package services

import (
	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// GetAllExpenseTypes gets all expense types (fixed: Needs, Wants, Savings)
func GetAllExpenseTypes() ([]models.ExpenseType, error) {
	var expenseTypes []models.ExpenseType
	result := db.DB.Where("status IN ?", models.GetActiveStatuses()).
		Order("created_at ASC").Find(&expenseTypes)
	if result.Error != nil {
		logger.Error("Error getting all expense types: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("All expense types retrieved successfully")
	return expenseTypes, nil
}

// GetExpenseTypeByID gets an expense type by ID
func GetExpenseTypeByID(id string) (*models.ExpenseType, error) {
	var expenseType models.ExpenseType
	result := db.DB.Where("id = ? AND status IN ?", id, models.GetActiveStatuses()).First(&expenseType)
	if result.Error != nil {
		logger.Error("Error getting expense type by id: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("ExpenseType retrieved successfully: %+v", expenseType)
	return &expenseType, nil
}

// GetExpenseTypeByName gets an expense type by name
func GetExpenseTypeByName(name string) (*models.ExpenseType, error) {
	var expenseType models.ExpenseType
	result := db.DB.Where("LOWER(name) = LOWER(?) AND status IN ?", name, models.GetActiveStatuses()).First(&expenseType)
	if result.Error != nil {
		logger.Error("Error getting expense type by name: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("ExpenseType retrieved by name successfully: %+v", expenseType)
	return &expenseType, nil
}

// GetActiveExpenseTypes gets all active expense types (alias for backward compatibility)
func GetActiveExpenseTypes() ([]models.ExpenseType, error) {
	return GetAllExpenseTypes()
}

// InitializeDefaultExpenseTypes creates the basic expense types for the 50/30/20 rule
// This should only be called once during system setup
func InitializeDefaultExpenseTypes() error {
	defaultTypes := []models.ExpenseType{
		{Name: "Needs"},   // 50%
		{Name: "Wants"},   // 30%
		{Name: "Savings"}, // 20%
	}
	
	for _, expenseType := range defaultTypes {
		// Check if it already exists
		var existing models.ExpenseType
		result := db.DB.Where("LOWER(name) = LOWER(?)", expenseType.Name).First(&existing)
		if result.Error != nil {
			// If it doesn't exist, create it
			expenseType.Status = models.StatusActive
			if err := db.DB.Create(&expenseType).Error; err != nil {
				logger.Error("Error creating default expense type %s: %v", expenseType.Name, err)
				return err
			}
			logger.Info("Default expense type created: %s", expenseType.Name)
		} else {
			logger.Info("Default expense type already exists: %s", expenseType.Name)
		}
	}
	
	return nil
}

// GetExpenseTypesWithUserCategories gets expense types with user's categories loaded
func GetExpenseTypesWithUserCategories(userID string) ([]models.ExpenseType, error) {
	var expenseTypes []models.ExpenseType
	
	// Get all expense types
	result := db.DB.Where("status IN ?", models.GetActiveStatuses()).
		Order("created_at ASC").Find(&expenseTypes)
	if result.Error != nil {
		logger.Error("Error getting expense types: %v", result.Error)
		return nil, result.Error
	}
	
	// Load user categories for each expense type
	for i := range expenseTypes {
		var categories []models.Category
		db.DB.Where("user_id = ? AND expense_type_id = ? AND status IN ?", 
			userID, expenseTypes[i].ID, models.GetActiveStatuses()).
			Order("name ASC").Find(&categories)
		
		// Add categories to the expense type (note: this won't persist to DB)
		// This is just for the response
		expenseTypes[i].Categories = categories
	}
	
	logger.Info("Expense types with user categories retrieved successfully for user %s", userID)
	return expenseTypes, nil
}

package services

import (
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// InitializeExpenseSystem initializes the basic expense types (Needs/Wants/Savings)
// This should only be called once during system setup
func InitializeExpenseSystem() error {
	logger.Info("Initializing expense system with default data...")
	
	// Creates the default expense types (only 3: Needs, Wants, Savings)
	if err := InitializeDefaultExpenseTypes(); err != nil {
		logger.Error("Error initializing default expense types: %v", err)
		return err
	}
	
	logger.Info("Expense system initialized successfully!")
	return nil
}

// SetupNewUser creates default categories for a new user
func SetupNewUser(userID string) error {
	logger.Info("Setting up default categories for new user: %s", userID)
	
	// Creates the default categories for the user
	if err := CreateDefaultUserCategories(userID); err != nil {
		logger.Error("Error creating default categories for user %s: %v", userID, err)
		return err
	}
	
	logger.Info("New user setup completed successfully for user: %s", userID)
	return nil
}

// GetSystemOverview gets an overview of the expense system setup
func GetSystemOverview() (map[string]interface{}, error) {
	overview := make(map[string]interface{})
	
	// Count expense types (these are fixed: Needs, Wants, Savings)
	expenseTypes, err := GetActiveExpenseTypes()
	if err != nil {
		return nil, err
	}
	overview["expense_types_count"] = len(expenseTypes)
	overview["expense_types"] = expenseTypes
	
	// In the new architecture, categories belong to specific users
	// We only show information about the fixed expense types
	overview["system_info"] = map[string]interface{}{
		"architecture": "user_specific_categories",
		"expense_types": []string{"Needs (50%)", "Wants (30%)", "Savings (20%)"},
		"note": "Categories are created per user. Use GetUserCategoriesGroupedByType(userID) for user-specific data.",
	}
	
	logger.Info("System overview generated successfully")
	return overview, nil
}

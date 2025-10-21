package services

import (
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// InitializeExpenseSystem initializes the expense system
// Since expense types are now fixed enums (needs/wants/savings), no database initialization is needed
func InitializeExpenseSystem() error {
	logger.Info("Expense system initialization check...")
	logger.Info("Expense types are fixed: %v", models.ValidExpenseTypes())
	logger.Info("Expense system ready!")
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
	
	// Expense types are now fixed enums
	expenseTypes := models.ValidExpenseTypes()
	overview["expense_types_count"] = len(expenseTypes)
	
	// Build expense types info
	var expenseTypesInfo []map[string]string
	for _, et := range expenseTypes {
		expenseTypesInfo = append(expenseTypesInfo, map[string]string{
			"value": string(et),
			"name":  models.GetExpenseTypeName(et),
		})
	}
	overview["expense_types"] = expenseTypesInfo
	
	// System info
	overview["system_info"] = map[string]interface{}{
		"architecture":   "user_specific_categories",
		"expense_types":  []string{"Needs (50%)", "Wants (30%)", "Savings (20%)"},
		"note":           "ExpenseTypes are now fixed enums. Categories are created per user.",
		"valid_types":    []string{"needs", "wants", "savings"},
	}
	
	logger.Info("System overview generated successfully")
	return overview, nil
}

package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

// CreateUserCategory creates a new category for the user
func CreateUserCategory(userID string, category *models.Category) error {
	// Force the UserID and Status to prevent manipulation
	category.UserID = uuid.MustParse(userID)
	category.Status = models.StatusActive
	
	// Validate that the ExpenseType is valid
	if !models.IsValidExpenseType(string(category.ExpenseType)) {
		logger.Error("Invalid expense type: %s", category.ExpenseType)
		return errors.New("invalid expense type. Must be one of: needs, wants, savings")
	}
	
	// Check if there is another category with the same name for this user in this type
	var existingCategory models.Category
	result := db.DB.Where("LOWER(name) = LOWER(?) AND user_id = ? AND expense_type = ? AND status IN ?", 
		category.Name, userID, category.ExpenseType, models.GetActiveStatuses()).First(&existingCategory)
	if result.Error == nil {
		logger.Error("Category with this name already exists for this user in this expense type")
		return errors.New("you already have a category with this name in this expense type")
	}
	
	result = db.DB.Create(category)
	if result.Error != nil {
		logger.Error("Error creating user category: %v", result.Error)
		return result.Error
	}
	
	logger.Info("User category created successfully: %+v", category)
	return nil
}

// GetUserCategoryByID gets a specific category for the user
func GetUserCategoryByID(userID string, id string) (*models.Category, error) {
	var category models.Category
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).
		First(&category)
	if result.Error != nil {
		logger.Error("Error getting user category by id: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("User category retrieved successfully: %+v", category)
	return &category, nil
}

// GetUserCategories gets all categories for the user
func GetUserCategories(userID string, includeDeleted bool) ([]models.Category, error) {
	var categories []models.Category
	query := db.DB.Where("user_id = ?", userID)
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("expense_type, name ASC").Find(&categories)
	if result.Error != nil {
		logger.Error("Error getting user categories: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("User categories retrieved successfully for user %s", userID)
	return categories, nil
}

// GetUserCategoriesByExpenseType gets user categories for a specific expense type
func GetUserCategoriesByExpenseType(userID string, expenseType string, includeDeleted bool) ([]models.Category, error) {
	// Validate expense type
	if !models.IsValidExpenseType(expenseType) {
		return nil, errors.New("invalid expense type. Must be one of: needs, wants, savings")
	}
	
	var categories []models.Category
	query := db.DB.Where("user_id = ? AND expense_type = ?", userID, expenseType)
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("name ASC").Find(&categories)
	if result.Error != nil {
		logger.Error("Error getting user categories by expense type: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("User categories by expense type retrieved successfully for user %s", userID)
	return categories, nil
}

// GetUserCategoriesByExpenseTypeName gets user categories for a specific expense type by name (kept for backward compatibility)
func GetUserCategoriesByExpenseTypeName(userID string, expenseTypeName string) ([]models.Category, error) {
	// Convert name to lowercase enum value
	expenseType := expenseTypeName
	switch expenseTypeName {
	case "Needs":
		expenseType = string(models.ExpenseTypeNeeds)
	case "Wants":
		expenseType = string(models.ExpenseTypeWants)
	case "Savings":
		expenseType = string(models.ExpenseTypeSavings)
	default:
		// Try as-is if already lowercase
		expenseType = expenseTypeName
	}
	
	return GetUserCategoriesByExpenseType(userID, expenseType, false)
}

// GetUserCategoriesGroupedByType gets user categories grouped by expense type
func GetUserCategoriesGroupedByType(userID string) (map[string][]models.Category, error) {
	categories, err := GetUserCategories(userID, false)
	if err != nil {
		return nil, err
	}
	
	grouped := make(map[string][]models.Category)
	for _, category := range categories {
		typeName := models.GetExpenseTypeName(category.ExpenseType)
		grouped[typeName] = append(grouped[typeName], category)
	}
	
	logger.Info("User categories grouped by type retrieved successfully for user %s", userID)
	return grouped, nil
}

// UpdateUserCategory updates a user's category
func UpdateUserCategory(userID string, id string, updatedCategory *models.Category) (*models.Category, error) {
	var existingCategory models.Category
	
	// Verify that the category exists, belongs to the user and is not deleted
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&existingCategory)
	if result.Error != nil {
		logger.Error("User category not found: %v", result.Error)
		return nil, errors.New("category not found or access denied")
	}
	
	// Validate the ExpenseType if it's being changed
	if existingCategory.ExpenseType != updatedCategory.ExpenseType {
		if !models.IsValidExpenseType(string(updatedCategory.ExpenseType)) {
			logger.Error("Invalid expense type: %s", updatedCategory.ExpenseType)
			return nil, errors.New("invalid expense type. Must be one of: needs, wants, savings")
		}
	}
	
	// Check if the name is unique in the type for this user if it is being changed
	if existingCategory.Name != updatedCategory.Name || existingCategory.ExpenseType != updatedCategory.ExpenseType {
		var duplicateCategory models.Category
		checkResult := db.DB.Where("LOWER(name) = LOWER(?) AND user_id = ? AND expense_type = ? AND id != ? AND status IN ?", 
			updatedCategory.Name, userID, updatedCategory.ExpenseType, id, models.GetActiveStatuses()).First(&duplicateCategory)
		if checkResult.Error == nil {
			logger.Error("Category name already exists for this user in this expense type")
			return nil, errors.New("you already have a category with this name in this expense type")
		}
	}
	
	// Prevent modification of protected fields
	updatedCategory.UserID = existingCategory.UserID
	updatedCategory.ID = existingCategory.ID
	updatedCategory.CreatedAt = existingCategory.CreatedAt
	updatedCategory.Status = existingCategory.Status
	updatedCategory.StatusChangedAt = existingCategory.StatusChangedAt
	
	// Update
	result = db.DB.Model(&existingCategory).Updates(updatedCategory)
	if result.Error != nil {
		logger.Error("Error updating user category: %v", result.Error)
		return nil, result.Error
	}
	
	// Get the updated category
	result = db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingCategory)
	if result.Error != nil {
		logger.Error("Error retrieving updated user category: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("User category updated successfully: %+v", existingCategory)
	return &existingCategory, nil
}

// SoftDeleteUserCategory marks a user's category as deleted
func SoftDeleteUserCategory(userID string, id string) error {
	// Check if the category exists, belongs to the user and is not deleted
	var existingCategory models.Category
	result := db.DB.Where("user_id = ? AND id = ? AND status != ?", userID, id, models.StatusDeleted).First(&existingCategory)
	if result.Error != nil {
		logger.Error("User category not found or already deleted: %v", result.Error)
		return errors.New("category not found, already deleted, or access denied")
	}
	
	// Check if it has active expenses
	var activeExpenses int64
	db.DB.Model(&models.Expense{}).Where("user_id = ? AND category_id = ? AND status IN ?", 
		userID, id, models.GetActiveStatuses()).Count(&activeExpenses)
	if activeExpenses > 0 {
		logger.Error("Cannot delete category with active expenses")
		return errors.New("cannot delete category: you have active expenses in this category")
	}
	
	// Mark as deleted
	now := time.Now()
	result = db.DB.Model(&existingCategory).Updates(map[string]interface{}{
		"status": models.StatusDeleted,
		"status_changed_at": &now,
	})
	
	if result.Error != nil {
		logger.Error("Error soft deleting user category: %v", result.Error)
		return result.Error
	}
	
	logger.Info("User category soft deleted successfully: %s", id)
	return nil
}

// RestoreUserCategory restores a deleted user category
func RestoreUserCategory(userID string, id string) (*models.Category, error) {
	// Check if the category exists, belongs to the user and is deleted
	var existingCategory models.Category
	result := db.DB.Where("user_id = ? AND id = ? AND status = ?", userID, id, models.StatusDeleted).First(&existingCategory)
	if result.Error != nil {
		logger.Error("User category not found, not deleted, or access denied: %v", result.Error)
		return nil, errors.New("category not found, not deleted, or access denied")
	}
	
	// Validate that the ExpenseType is still valid (it should always be since it's an enum)
	if !models.IsValidExpenseType(string(existingCategory.ExpenseType)) {
		logger.Error("Cannot restore category: expense type is not valid")
		return nil, errors.New("cannot restore category: expense type is not valid")
	}
	
	// Check if there is a conflict of names
	var duplicateCategory models.Category
	checkResult := db.DB.Where("LOWER(name) = LOWER(?) AND user_id = ? AND expense_type = ? AND id != ? AND status IN ?", 
		existingCategory.Name, userID, existingCategory.ExpenseType, id, models.GetActiveStatuses()).First(&duplicateCategory)
	if checkResult.Error == nil {
		logger.Error("Cannot restore: category name already exists for this user in this expense type")
		return nil, errors.New("cannot restore: you already have a category with this name in this expense type")
	}
	
	// Restore as active
	now := time.Now()
	result = db.DB.Model(&existingCategory).Updates(map[string]interface{}{
		"status": models.StatusActive,
		"status_changed_at": &now,
	})
	
	if result.Error != nil {
		logger.Error("Error restoring user category: %v", result.Error)
		return nil, result.Error
	}
	
	// Get the updated category
	updatedCategory, err := GetUserCategoryByID(userID, id)
	if err != nil {
		logger.Error("Error retrieving updated category: %v", err)
		return nil, errors.New("error retrieving updated category")
	}
	
	logger.Info("User category restored successfully: %s", id)
	return updatedCategory, nil
}

// CreateDefaultUserCategories creates default categories for a new user
func CreateDefaultUserCategories(userID string) error {
	// Define default categories for each expense type
	defaultCategories := map[models.ExpenseType][]string{
		models.ExpenseTypeNeeds: {
			"Vivienda", "Alimentación", "Transporte", "Salud", "Servicios básicos",
		},
		models.ExpenseTypeWants: {
			"Entretenimiento", "Restaurantes", "Shopping", "Hobbies", "Viajes",
		},
		models.ExpenseTypeSavings: {
			"Fondo de emergencia", "Ahorro general", "Inversiones",
		},
	}
	
	for expenseType, categoryNames := range defaultCategories {
		for _, categoryName := range categoryNames {
			category := models.Category{
				UserID:      uuid.MustParse(userID),
				Name:        categoryName,
				ExpenseType: expenseType,
			}
			
			// Create category (CreateUserCategory already checks for duplicates)
			if err := CreateUserCategory(userID, &category); err != nil {
				logger.Error("Error creating default category %s for user %s: %v", categoryName, userID, err)
				// Continue with the other categories
			} else {
				logger.Info("Default category created: %s for user %s", categoryName, userID)
			}
		}
	}
	
	return nil
}

// GetUserCategoryStats gets statistics about user's categories
func GetUserCategoryStats(userID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total categories by user
	var totalCategories int64
	db.DB.Model(&models.Category{}).Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses()).Count(&totalCategories)
	stats["total_categories"] = totalCategories
	
	// Categories by type
	typeStats := make(map[string]int64)
	for _, expenseType := range models.ValidExpenseTypes() {
		var count int64
		db.DB.Model(&models.Category{}).Where("user_id = ? AND expense_type = ? AND status IN ?", 
			userID, expenseType, models.GetActiveStatuses()).Count(&count)
		typeStats[models.GetExpenseTypeName(expenseType)] = count
	}
	stats["categories_by_type"] = typeStats
	
	// Deleted categories
	var deletedCategories int64
	db.DB.Model(&models.Category{}).Where("user_id = ? AND status = ?", userID, models.StatusDeleted).Count(&deletedCategories)
	stats["deleted_categories"] = deletedCategories
	
	logger.Info("User category stats retrieved successfully for user %s", userID)
	return stats, nil
}

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
	
	// Check if the ExpenseType exists and is active
	var expenseType models.ExpenseType
	result := db.DB.Where("id = ? AND status IN ?", category.ExpenseTypeID, models.GetActiveStatuses()).First(&expenseType)
	if result.Error != nil {
		logger.Error("ExpenseType not found or not active")
		return errors.New("expense type not found or not active")
	}
	
	// Check if there is another category with the same name for this user in this type
	var existingCategory models.Category
	result = db.DB.Where("LOWER(name) = LOWER(?) AND user_id = ? AND expense_type_id = ? AND status IN ?", 
		category.Name, userID, category.ExpenseTypeID, models.GetActiveStatuses()).First(&existingCategory)
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
		Preload("ExpenseType").First(&category)
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
	query := db.DB.Where("user_id = ?", userID).Preload("ExpenseType")
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("expense_type_id, name ASC").Find(&categories)
	if result.Error != nil {
		logger.Error("Error getting user categories: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("User categories retrieved successfully for user %s", userID)
	return categories, nil
}

// GetUserCategoriesByExpenseType gets user categories for a specific expense type
func GetUserCategoriesByExpenseType(userID string, expenseTypeID string, includeDeleted bool) ([]models.Category, error) {
	var categories []models.Category
	query := db.DB.Where("user_id = ? AND expense_type_id = ?", userID, expenseTypeID).Preload("ExpenseType")
	
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

// GetUserCategoriesByExpenseTypeName gets user categories for a specific expense type by name
func GetUserCategoriesByExpenseTypeName(userID string, expenseTypeName string) ([]models.Category, error) {
	// Primero obtener el ExpenseType
	expenseType, err := GetExpenseTypeByName(expenseTypeName)
	if err != nil {
		return nil, err
	}
	
	return GetUserCategoriesByExpenseType(userID, expenseType.ID.String(), false)
}

// GetUserCategoriesGroupedByType gets user categories grouped by expense type
func GetUserCategoriesGroupedByType(userID string) (map[string][]models.Category, error) {
	categories, err := GetUserCategories(userID, false)
	if err != nil {
		return nil, err
	}
	
	grouped := make(map[string][]models.Category)
	for _, category := range categories {
		typeName := category.ExpenseType.Name
		grouped[typeName] = append(grouped[typeName], category)
	}
	
	logger.Info("User categories grouped by type retrieved successfully for user %s", userID)
	return grouped, nil
}

// UpdateUserCategory updates a user's category
func UpdateUserCategory(userID string, id string, updatedCategory *models.Category) (*models.Category, error) {
	var existingCategory models.Category
	
	// Verificar que la categoría existe, pertenece al usuario y no está eliminada
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&existingCategory)
	if result.Error != nil {
		logger.Error("User category not found: %v", result.Error)
		return nil, errors.New("category not found or access denied")
	}
	
	// Verificar que el ExpenseType existe y está activo si se está cambiando
	if existingCategory.ExpenseTypeID != updatedCategory.ExpenseTypeID {
		var expenseType models.ExpenseType
		result := db.DB.Where("id = ? AND status IN ?", updatedCategory.ExpenseTypeID, models.GetActiveStatuses()).First(&expenseType)
		if result.Error != nil {
			logger.Error("ExpenseType not found or not active")
			return nil, errors.New("expense type not found or not active")
		}
	}
	
	// Check if the name is unique in the type for this user if it is being changed
	if existingCategory.Name != updatedCategory.Name || existingCategory.ExpenseTypeID != updatedCategory.ExpenseTypeID {
		var duplicateCategory models.Category
		checkResult := db.DB.Where("LOWER(name) = LOWER(?) AND user_id = ? AND expense_type_id = ? AND id != ? AND status IN ?", 
			updatedCategory.Name, userID, updatedCategory.ExpenseTypeID, id, models.GetActiveStatuses()).First(&duplicateCategory)
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
	
	// Get the updated category with relations
	result = db.DB.Where("user_id = ? AND id = ?", userID, id).Preload("ExpenseType").First(&existingCategory)
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
func RestoreUserCategory(userID string, id string) error {
	// Check if the category exists, belongs to the user and is deleted
	var existingCategory models.Category
	result := db.DB.Where("user_id = ? AND id = ? AND status = ?", userID, id, models.StatusDeleted).First(&existingCategory)
	if result.Error != nil {
		logger.Error("User category not found, not deleted, or access denied: %v", result.Error)
		return errors.New("category not found, not deleted, or access denied")
	}
	
	// Check if the ExpenseType is active
	var expenseType models.ExpenseType
	result = db.DB.Where("id = ? AND status IN ?", existingCategory.ExpenseTypeID, models.GetActiveStatuses()).First(&expenseType)
	if result.Error != nil {
		logger.Error("Cannot restore category: expense type is not active")
		return errors.New("cannot restore category: expense type is not active")
	}
	
	// Check if there is a conflict of names
	var duplicateCategory models.Category
	checkResult := db.DB.Where("LOWER(name) = LOWER(?) AND user_id = ? AND expense_type_id = ? AND id != ? AND status IN ?", 
		existingCategory.Name, userID, existingCategory.ExpenseTypeID, id, models.GetActiveStatuses()).First(&duplicateCategory)
	if checkResult.Error == nil {
		logger.Error("Cannot restore: category name already exists for this user in this expense type")
		return errors.New("cannot restore: you already have a category with this name in this expense type")
	}
	
	// Restore as active
	now := time.Now()
	result = db.DB.Model(&existingCategory).Updates(map[string]interface{}{
		"status": models.StatusActive,
		"status_changed_at": &now,
	})
	
	if result.Error != nil {
		logger.Error("Error restoring user category: %v", result.Error)
		return result.Error
	}
	
	logger.Info("User category restored successfully: %s", id)
	return nil
}

// CreateDefaultUserCategories creates default categories for a new user
func CreateDefaultUserCategories(userID string) error {
	// Get the expense types
	expenseTypes, err := GetActiveExpenseTypes()
	if err != nil {
		return err
	}
	
	// Map types by name to facilitate creation
	typeMap := make(map[string]uuid.UUID)
	for _, et := range expenseTypes {
		typeMap[et.Name] = et.ID
	}
	
	// Define default categories more simple and common
	defaultCategories := map[string][]string{
		"Needs": {
			"Vivienda", "Alimentación", "Transporte", "Salud", "Servicios básicos",
		},
		"Wants": {
			"Entretenimiento", "Restaurantes", "Shopping", "Hobbies", "Viajes",
		},
		"Savings": {
			"Fondo de emergencia", "Ahorro general", "Inversiones",
		},
	}
	
	for typeName, categoryNames := range defaultCategories {
		typeID, exists := typeMap[typeName]
		if !exists {
			logger.Error("ExpenseType %s not found", typeName)
			continue
		}
		
		for _, categoryName := range categoryNames {
			category := models.Category{
				UserID:        uuid.MustParse(userID),
				Name:          categoryName,
				ExpenseTypeID: typeID,
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
	expenseTypes, err := GetActiveExpenseTypes()
	if err != nil {
		return nil, err
	}
	
	for _, expenseType := range expenseTypes {
		var count int64
		db.DB.Model(&models.Category{}).Where("user_id = ? AND expense_type_id = ? AND status IN ?", 
			userID, expenseType.ID, models.GetActiveStatuses()).Count(&count)
		typeStats[expenseType.Name] = count
	}
	stats["categories_by_type"] = typeStats
	
	// Deleted categories
	var deletedCategories int64
	db.DB.Model(&models.Category{}).Where("user_id = ? AND status = ?", userID, models.StatusDeleted).Count(&deletedCategories)
	stats["deleted_categories"] = deletedCategories
	
	logger.Info("User category stats retrieved successfully for user %s", userID)
	return stats, nil
}

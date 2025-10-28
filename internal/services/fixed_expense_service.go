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

// CreateFixedExpense creates a new fixed expense
func CreateFixedExpense(userID string,fixedExpense models.FixedExpense)(*models.FixedExpense,error){
	// Force basic Fields
	fixedExpense.UserID = uuid.MustParse(userID)
	fixedExpense.Status = models.StatusActive
	fixedExpense.CreatedAt = time.Now()
	fixedExpense.UpdatedAt = time.Now()

	// Verify bank account exists and belongs to user
	var bankAccount models.BankAccount
	result := db.DB.Where("id = ? AND user_id = ? AND status IN ?", 
		fixedExpense.BankAccountID, userID, models.GetActiveStatuses()).First(&bankAccount)
	if result.Error != nil {
		logger.Error("Bank account not found or not active")
		return nil, errors.New("bank account not found or not active")
	}

	// Set next due date
	fixedExpense.NextDueDate = fixedExpense.DueDate

	result = db.DB.Create(&fixedExpense)
	if result.Error != nil {
		logger.Error("Error creating fixed expense: %v", result.Error)
		return nil,errors.New("error creating fixed expense")
	}

	return &fixedExpense,nil
}

// GetFixedExpenseByID returns a fixed expense by its ID
func GetFixedExpenseByID(userID string,id string)(*models.FixedExpense,error){
	var fixedExpense models.FixedExpense
	result := db.DB.Where("user_id = ? AND id = ?",userID,id).First(&fixedExpense)
	if result.Error != nil {
		logger.Error("Error getting fixed expense: %v", result.Error)
		return nil,errors.New("error getting fixed expense")
	}

	return &fixedExpense,nil
}

func GetFixedExpenses(userID string,includeDeleted bool)([]models.FixedExpense,error){
	var fixedExpenses []models.FixedExpense
	query := db.DB.Where("user_id = ?",userID)

	if !includeDeleted{
		query = query.Where("status = ?",models.StatusActive)
	}

	result := query.Find(&fixedExpenses)
	if result.Error != nil {
		logger.Error("Error getting fixed expenses: %v", result.Error)
		return nil,errors.New("error getting fixed expenses")
	}

	return fixedExpenses,nil
}

func UpdateFixedExpense(userID string,id string,fixedExpense models.FixedExpense)(*models.FixedExpense,error){
	var existingFixedExpense models.FixedExpense
	result := db.DB.Where("user_id = ? AND id = ?",userID,id).First(&existingFixedExpense)
	if result.Error != nil {
		logger.Error("Error getting fixed expense: %v", result.Error)
		return nil,errors.New("error getting fixed expense")
	}

	if existingFixedExpense.Status.IsDeleted(){
		logger.Error("Fixed expense is deleted")
		return nil,errors.New("fixed expense is deleted")
	}

	existingFixedExpense.Name = fixedExpense.Name
	existingFixedExpense.Amount = fixedExpense.Amount
	existingFixedExpense.DueDate = fixedExpense.DueDate
	existingFixedExpense.UpdatedAt = time.Now()

	result = db.DB.Save(&existingFixedExpense)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound{
		logger.Error("Error updating fixed expense: %v", result.Error)
		return nil,errors.New("error updating fixed expense")
	}

	return &existingFixedExpense,nil
}

func DeleteFixedExpense(userID string,id string)(*models.FixedExpense,error){
	var existingFixedExpense models.FixedExpense
	result := db.DB.Where("user_id = ? AND id = ?",userID,id).First(&existingFixedExpense)
	if result.Error != nil {
		logger.Error("Error getting fixed expense: %v", result.Error)
		return nil,errors.New("error getting fixed expense")
	}

	if existingFixedExpense.Status.IsDeleted(){
		logger.Error("Fixed expense is deleted")
		return nil,errors.New("fixed expense is deleted")
	}

	result = db.DB.Model(&existingFixedExpense).Update("status",models.StatusDeleted)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound{
		logger.Error("Error deleting fixed expense: %v", result.Error)
		return nil,errors.New("error deleting fixed expense")
	}

	return &existingFixedExpense,nil
}

// GetFixedExpensesForMonth returns all fixed expenses that should apply for a specific year/month
// This includes recurring expenses and handles the logic of whether they should apply
func GetFixedExpensesForMonth(userID string, year int, month time.Month) ([]models.FixedExpense, error) {
	var allFixedExpenses []models.FixedExpense
	result := db.DB.Where("user_id = ? AND status = ? AND is_recurring = ?", 
		userID, models.StatusActive, true).Find(&allFixedExpenses)
	
	if result.Error != nil {
		logger.Error("Error getting fixed expenses: %v", result.Error)
		return nil, errors.New("error getting fixed expenses")
	}

	// Filter to only those that should apply this month
	var applicableExpenses []models.FixedExpense
	for _, expense := range allFixedExpenses {
		if expense.ShouldApplyForMonth(year, month) {
			applicableExpenses = append(applicableExpenses, expense)
		}
	}

	return applicableExpenses, nil
}

// GetCommittedBudgetForMonth calculates the total amount committed to fixed expenses for a month
func GetCommittedBudgetForMonth(userID string, year int, month time.Month) (float64, error) {
	fixedExpenses, err := GetFixedExpensesForMonth(userID, year, month)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, expense := range fixedExpenses {
		total += expense.Amount
	}

	logger.Info("Committed budget for %d-%02d: $%.2f", year, month, total)
	return total, nil
}

// GetFixedExpensesByCategoryType returns fixed expenses grouped by their category's expense type (needs/wants/savings)
func GetFixedExpensesByCategoryType(userID string, year int, month time.Month) (map[string]float64, error) {
	fixedExpenses, err := GetFixedExpensesForMonth(userID, year, month)
	if err != nil {
		return nil, err
	}

	// If no fixed expenses, return empty map
	if len(fixedExpenses) == 0 {
		return make(map[string]float64), nil
	}

	// Group by expense type
	result := make(map[string]float64)
	
	for _, expense := range fixedExpenses {
		// If no category, we need to assign to a default (let's use "Wants" as default)
		if expense.CategoryID == nil {
			result["Wants"] += expense.Amount
			continue
		}

		// Get the category
		var category models.Category
		if db.DB.Where("id = ?", expense.CategoryID).First(&category).Error != nil {
			// If category not found, default to "Wants"
			result["Wants"] += expense.Amount
			continue
		}

		// Map category expense_type to the type name
		var typeName string
		switch category.ExpenseType {
		case models.ExpenseTypeNeeds:
			typeName = "Needs"
		case models.ExpenseTypeWants:
			typeName = "Wants"
		case models.ExpenseTypeSavings:
			typeName = "Savings"
		default:
			typeName = "Wants"
		}
		
		result[typeName] += expense.Amount
	}

	return result, nil
}

// GetUpcomingFixedExpenses returns fixed expenses due in the next N days
func GetUpcomingFixedExpenses(userID string, days int) ([]models.FixedExpense, error) {
	var fixedExpenses []models.FixedExpense
	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	result := db.DB.Where("user_id = ? AND status = ? AND is_recurring = ? AND due_date >= ? AND due_date <= ?",
		userID, models.StatusActive, true, now, futureDate).
		Order("due_date ASC").
		Find(&fixedExpenses)

	if result.Error != nil {
		logger.Error("Error getting upcoming fixed expenses: %v", result.Error)
		return nil, result.Error
	}

	return fixedExpenses, nil
}

// ProcessDueFixedExpenses processes all fixed expenses that are due today
// This should be called by a scheduled job (cron/task scheduler)
func ProcessDueFixedExpenses() error {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	
	var dueFixedExpenses []models.FixedExpense
	result := db.DB.Where("next_due_date <= ? AND status = ? AND is_recurring = ?",
		today, models.StatusActive, true).
		Preload("BankAccount").
		Find(&dueFixedExpenses)
	
	if result.Error != nil {
		logger.Error("Error fetching due fixed expenses: %v", result.Error)
		return result.Error
	}
	
	for _, fixedExpense := range dueFixedExpenses {
		if err := processFixedExpense(&fixedExpense); err != nil {
			logger.Error("Error processing fixed expense %s: %v", fixedExpense.ID, err)
			continue // Continue processing others even if one fails
		}
	}
	
	logger.Info("Processed %d fixed expenses", len(dueFixedExpenses))
	return nil
}

// processFixedExpense creates an expense record and updates bank account
func processFixedExpense(fixedExpense *models.FixedExpense) error {
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Check if bank account has sufficient balance (warning only)
	var bankAccount models.BankAccount
	if err := tx.Where("id = ?", fixedExpense.BankAccountID).First(&bankAccount).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	if bankAccount.Balance < fixedExpense.Amount {
		logger.Warn("Fixed expense %s will cause negative balance in account %s",
			fixedExpense.Name, bankAccount.AccountName)
	}
	
	// Create an expense record
	expense := &models.Expense{
		UserID:        fixedExpense.UserID,
		Amount:        fixedExpense.Amount,
		Date:          time.Now().UTC(),
		BankAccountID: fixedExpense.BankAccountID,
		Description:   &fixedExpense.Name,
		Status:        models.StatusActive,
	}
	
	// Handle category if provided
	if fixedExpense.CategoryID != nil {
		expense.CategoryID = *fixedExpense.CategoryID
	} else {
		// If no category, we need to create a default one or skip
		// For now, we'll skip creating the expense if no category
		tx.Rollback()
		logger.Warn("Fixed expense %s has no category, skipping", fixedExpense.Name)
		return nil
	}
	
	if err := tx.Create(expense).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Update bank account balance
	if err := tx.Model(&bankAccount).
		Update("balance", gorm.Expr("balance - ?", fixedExpense.Amount)).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Update fixed expense next due date
	nextDueDate := calculateNextDueDate(fixedExpense)
	now := time.Now()
	if err := tx.Model(fixedExpense).Updates(map[string]interface{}{
		"last_processed_at": &now,
		"next_due_date":     nextDueDate,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	tx.Commit()
	logger.Info("Processed fixed expense: %s, created expense: %s", fixedExpense.Name, expense.ID)
	return nil
}

// calculateNextDueDate calculates the next due date based on recurrence type
func calculateNextDueDate(fixedExpense *models.FixedExpense) time.Time {
	currentDue := fixedExpense.NextDueDate
	
	if fixedExpense.RecurrenceType == "monthly" {
		// Add one month
		return currentDue.AddDate(0, 1, 0)
	} else if fixedExpense.RecurrenceType == "yearly" {
		// Add one year
		return currentDue.AddDate(1, 0, 0)
	}
	
	// Default: monthly
	return currentDue.AddDate(0, 1, 0)
}
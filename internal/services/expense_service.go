package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

// CreateExpense creates a new expense for the user
func CreateExpense(userID string, expense *models.Expense) error {
	// Force the UserID and Status to prevent manipulation
	expense.UserID = uuid.MustParse(userID)
	expense.Status = models.StatusActive
	
	// Verify that the category exists and is active
	var category models.Category
	result := db.DB.Where("id = ? AND status IN ?", expense.CategoryID, models.GetActiveStatuses()).First(&category)
	if result.Error != nil {
		logger.Error("Category not found or not active")
		return errors.New("category not found or not active")
	}
	
	// Verify that the bank account exists, is active and belongs to the user
	var bankAccount models.BankAccount
	result = db.DB.Where("id = ? AND user_id = ? AND status IN ?", 
		expense.BankAccountID, userID, models.GetActiveStatuses()).First(&bankAccount)
	if result.Error != nil {
		logger.Error("Bank account not found, not active, or doesn't belong to user")
		return errors.New("bank account not found, not active, or access denied")
	}
	
	// Verify that the amount is positive
	if expense.Amount <= 0 {
		logger.Error("Expense amount must be positive")
		return errors.New("expense amount must be positive")
	}
	
	result = db.DB.Create(expense)
	if result.Error != nil {
		logger.Error("Error creating expense: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Expense created successfully: %+v", expense)
	return nil
}

// GetExpenseByID gets a specific expense for the user
func GetExpenseByID(userID string, id string) (*models.Expense, error) {
	var expense models.Expense
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).
		Preload("Category").Preload("BankAccount").First(&expense)
	if result.Error != nil {
		logger.Error("Error getting expense by id: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Expense retrieved successfully: %+v", expense)
	return &expense, nil
}

// GetAllExpenses gets all expenses for the user
func GetAllExpenses(userID string, includeDeleted bool) ([]models.Expense, error) {
	var expenses []models.Expense
	query := db.DB.Where("user_id = ?", userID).
		Preload("Category").Preload("BankAccount")
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("date DESC, created_at DESC").Find(&expenses)
	if result.Error != nil {
		logger.Error("Error getting all expenses: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("All expenses retrieved successfully: %+v", expenses)
	return expenses, nil
}

// GetActiveExpenses gets all active expenses for the user
func GetActiveExpenses(userID string) ([]models.Expense, error) {
	var expenses []models.Expense
	result := db.DB.Where("user_id = ? AND status IN ?", userID, models.GetActiveStatuses()).
		Preload("Category").Preload("BankAccount").
		Order("date DESC, created_at DESC").Find(&expenses)
	if result.Error != nil {
		logger.Error("Error getting active expenses: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Active expenses retrieved successfully: %+v", expenses)
	return expenses, nil
}

// GetDeletedExpenses gets all deleted expenses for the user
func GetDeletedExpenses(userID string) ([]models.Expense, error) {
	var expenses []models.Expense
	result := db.DB.Where("user_id = ? AND status = ?", userID, models.StatusDeleted).
		Preload("Category").Preload("BankAccount").
		Order("status_changed_at DESC").Find(&expenses)
	if result.Error != nil {
		logger.Error("Error getting deleted expenses: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Deleted expenses retrieved successfully: %+v", expenses)
	return expenses, nil
}

// GetExpensesByDateRange gets expenses in a date range for the user
func GetExpensesByDateRange(userID string, startDate, endDate time.Time, includeDeleted bool) ([]models.Expense, error) {
	var expenses []models.Expense
	query := db.DB.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Preload("Category").Preload("BankAccount")
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("date DESC, created_at DESC").Find(&expenses)
	if result.Error != nil {
		logger.Error("Error getting expenses by date range: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Expenses by date range retrieved successfully: %+v", expenses)
	return expenses, nil
}

// GetExpensesByCategory gets expenses for a specific category for the user
func GetExpensesByCategory(userID string, categoryID string, includeDeleted bool) ([]models.Expense, error) {
	var expenses []models.Expense
	query := db.DB.Where("user_id = ? AND category_id = ?", userID, categoryID).
		Preload("Category").Preload("BankAccount")
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("date DESC, created_at DESC").Find(&expenses)
	if result.Error != nil {
		logger.Error("Error getting expenses by category: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Expenses by category retrieved successfully: %+v", expenses)
	return expenses, nil
}

// GetExpensesByBankAccount gets expenses for a specific bank account for the user
func GetExpensesByBankAccount(userID string, bankAccountID string, includeDeleted bool) ([]models.Expense, error) {
	var expenses []models.Expense
	query := db.DB.Where("user_id = ? AND bank_account_id = ?", userID, bankAccountID).
		Preload("Category").Preload("BankAccount")
	
	if !includeDeleted {
		query = query.Where("status IN ?", models.GetVisibleStatuses())
	}
	
	result := query.Order("date DESC, created_at DESC").Find(&expenses)
	if result.Error != nil {
		logger.Error("Error getting expenses by bank account: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Expenses by bank account retrieved successfully: %+v", expenses)
	return expenses, nil
}

// GetMonthlyExpenses gets expenses for a specific month for the user
func GetMonthlyExpenses(userID string, year int, month int, includeDeleted bool) ([]models.Expense, error) {
	// Calcular el rango de fechas del mes
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Último día del mes
	
	return GetExpensesByDateRange(userID, startDate, endDate, includeDeleted)
}

// PatchExpense updates an expense for the user
func PatchExpense(userID string, id string, expense *models.Expense) (*models.Expense, error) {
	var existingExpense models.Expense
	
	// Verificar que el gasto existe, pertenece al usuario y no está eliminado
	result := db.DB.Where("user_id = ? AND id = ? AND status IN ?", userID, id, models.GetVisibleStatuses()).First(&existingExpense)
	if result.Error != nil {
		logger.Error("Expense not found or doesn't belong to user: %v", result.Error)
		return nil, errors.New("expense not found or access denied")
	}
	
	// Verificar que la categoría existe y está activa si se está cambiando
	if existingExpense.CategoryID != expense.CategoryID {
		var category models.Category
		result := db.DB.Where("id = ? AND status IN ?", expense.CategoryID, models.GetActiveStatuses()).First(&category)
		if result.Error != nil {
			logger.Error("Category not found or not active")
			return nil, errors.New("category not found or not active")
		}
	}
	
	// Verificar que la cuenta bancaria existe, está activa y pertenece al usuario si se está cambiando
	if existingExpense.BankAccountID != expense.BankAccountID {
		var bankAccount models.BankAccount
		result := db.DB.Where("id = ? AND user_id = ? AND status IN ?", 
			expense.BankAccountID, userID, models.GetActiveStatuses()).First(&bankAccount)
		if result.Error != nil {
			logger.Error("Bank account not found, not active, or doesn't belong to user")
			return nil, errors.New("bank account not found, not active, or access denied")
		}
	}
	
	// Validar que el monto es positivo
	if expense.Amount <= 0 {
		logger.Error("Expense amount must be positive")
		return nil, errors.New("expense amount must be positive")
	}
	
	// Prevenir modificación de campos protegidos
	expense.UserID = existingExpense.UserID
	expense.ID = existingExpense.ID
	expense.CreatedAt = existingExpense.CreatedAt
	
	// No permitir cambio de status a través de patch normal (usar funciones específicas)
	expense.Status = existingExpense.Status
	expense.StatusChangedAt = existingExpense.StatusChangedAt
	
	// Actualizar
	result = db.DB.Model(&existingExpense).Where("user_id = ? AND id = ?", userID, id).Updates(expense)
	if result.Error != nil {
		logger.Error("Error patching expense: %v", result.Error)
		return nil, result.Error
	}
	
	// Obtener el gasto actualizado con relaciones
	result = db.DB.Where("user_id = ? AND id = ?", userID, id).
		Preload("Category").Preload("BankAccount").First(&existingExpense)
	if result.Error != nil {
		logger.Error("Error retrieving updated expense: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Expense patched successfully: %+v", existingExpense)
	return &existingExpense, nil
}

// SoftDeleteExpense marks an expense as deleted for the user
func SoftDeleteExpense(userID string, id string) error {
	// Verificar que el gasto existe y pertenece al usuario
	var existingExpense models.Expense
	result := db.DB.Where("user_id = ? AND id = ? AND status != ?", userID, id, models.StatusDeleted).First(&existingExpense)
	if result.Error != nil {
		logger.Error("Expense not found or already deleted: %v", result.Error)
		return errors.New("expense not found or already deleted")
	}
	
	// Marcar como eliminado
	now := time.Now()
	result = db.DB.Model(&existingExpense).Updates(map[string]interface{}{
		"status": models.StatusDeleted,
		"status_changed_at": &now,
	})
	
	if result.Error != nil {
		logger.Error("Error soft deleting expense: %v", result.Error)
		return result.Error
	}
	
	logger.Info("Expense soft deleted successfully: %s", id)
	return nil
}

// RestoreExpense restores a deleted expense for the user
func RestoreExpense(userID string, id string) (*models.Expense, error) {
	// Verificar que el gasto existe, pertenece al usuario y está eliminado
	var existingExpense models.Expense
	result := db.DB.Where("user_id = ? AND id = ? AND status = ?", userID, id, models.StatusDeleted).First(&existingExpense)
	if result.Error != nil {
		logger.Error("Expense not found, not deleted, or access denied: %v", result.Error)
		return nil, errors.New("expense not found, not deleted, or access denied")
	}
	
	// Verificar que la categoría y cuenta bancaria siguen activas
	var category models.Category
	result = db.DB.Where("id = ? AND status IN ?", existingExpense.CategoryID, models.GetActiveStatuses()).First(&category)
	if result.Error != nil {
		logger.Error("Cannot restore expense: category is not active")
		return nil, errors.New("cannot restore expense: category is not active")
	}
	
	var bankAccount models.BankAccount
	result = db.DB.Where("id = ? AND user_id = ? AND status IN ?", 
		existingExpense.BankAccountID, userID, models.GetActiveStatuses()).First(&bankAccount)
	if result.Error != nil {
		logger.Error("Cannot restore expense: bank account is not active")
		return nil, errors.New("cannot restore expense: bank account is not active")
	}
	
	// Restaurar como activo
	now := time.Now()
	result = db.DB.Model(&existingExpense).Updates(map[string]interface{}{
		"status": models.StatusActive,
		"status_changed_at": &now,
	})
	
	if result.Error != nil {
		logger.Error("Error restoring expense: %v", result.Error)
		return nil, result.Error
	}
	
	// Get the updated expense with all relationships
	updatedExpense, err := GetExpenseByID(userID, id)
	if err != nil {
		logger.Error("Error retrieving updated expense: %v", err)
		return nil, errors.New("error retrieving updated expense")
	}
	
	logger.Info("Expense restored successfully: %s", id)
	return updatedExpense, nil
}

// ChangeExpenseStatus changes the status of an expense for the user
func ChangeExpenseStatus(userID string, id string, newStatus models.Status, reason *string) (*models.Expense, error) {
	// Validar que el status es válido
	if !models.ValidateStatus(newStatus) {
		return nil, errors.New("invalid status")
	}
	
	// Verificar que el gasto existe y pertenece al usuario
	var existingExpense models.Expense
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).First(&existingExpense)
	if result.Error != nil {
		logger.Error("Expense not found: %v", result.Error)
		return nil, errors.New("expense not found or access denied")
	}
	
	// No hacer nada si ya tiene ese status - return current expense
	if existingExpense.Status == newStatus {
		updatedExpense, err := GetExpenseByID(userID, id)
		if err != nil {
			logger.Error("Error retrieving expense: %v", err)
			return nil, errors.New("error retrieving expense")
		}
		return updatedExpense, nil
	}
	
	// Actualizar status
	now := time.Now()
	updates := map[string]interface{}{
		"status": newStatus,
		"status_changed_at": &now,
	}
	
	result = db.DB.Model(&existingExpense).Updates(updates)
	if result.Error != nil {
		logger.Error("Error changing expense status: %v", result.Error)
		return nil, result.Error
	}
	
	// Get the updated expense with all relationships
	updatedExpense, err := GetExpenseByID(userID, id)
	if err != nil {
		logger.Error("Error retrieving updated expense: %v", err)
		return nil, errors.New("error retrieving updated expense")
	}
	
	logger.Info("Expense status changed to %s successfully: %s", newStatus, id)
	return updatedExpense, nil
}

// HardDeleteExpense permanently deletes an expense for the user
func HardDeleteExpense(userID string, id string) error {
	// SOLO para casos especiales - elimina permanentemente
	// Verificar que el gasto existe y pertenece al usuario
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&models.Expense{})
	if result.Error != nil {
		logger.Error("Error hard deleting expense: %v", result.Error)
		return result.Error
	}
	
	// Verificar que realmente se eliminó algo
	if result.RowsAffected == 0 {
		logger.Error("Expense not found or doesn't belong to user")
		return errors.New("expense not found or access denied")
	}
	
	logger.Info("Expense permanently deleted: %s", id)
	return nil
}

// === ANÁLISIS Y ESTADÍSTICAS ===

// GetExpensesSummaryByPeriod gets expense summary for a period
func GetExpensesSummaryByPeriod(userID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var summary map[string]interface{}
	summary = make(map[string]interface{})
	
	// Total gastado en el período
	var totalAmount float64
	result := db.DB.Model(&models.Expense{}).
		Where("user_id = ? AND date BETWEEN ? AND ? AND status IN ?", 
			userID, startDate, endDate, models.GetActiveStatuses()).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)
	if result.Error != nil {
		logger.Error("Error calculating total expenses: %v", result.Error)
		return nil, result.Error
	}
	summary["total_amount"] = totalAmount
	
	// Contar total de gastos
	var totalCount int64
	db.DB.Model(&models.Expense{}).
		Where("user_id = ? AND date BETWEEN ? AND ? AND status IN ?", 
			userID, startDate, endDate, models.GetActiveStatuses()).Count(&totalCount)
	summary["total_count"] = totalCount
	
	// Promedio por gasto
	if totalCount > 0 {
		summary["average_amount"] = totalAmount / float64(totalCount)
	} else {
		summary["average_amount"] = 0.0
	}
	
	// Gastos por ExpenseType (50/30/20)
	var expensesByType []struct {
		ExpenseTypeName string  `json:"expense_type_name"`
		TotalAmount     float64 `json:"total_amount"`
		Count           int64   `json:"count"`
	}
	
	result = db.DB.Table("expenses e").
		Select(`(CASE 
			WHEN c.expense_type = 'needs' THEN 'Needs'
			WHEN c.expense_type = 'wants' THEN 'Wants'
			WHEN c.expense_type = 'savings' THEN 'Savings'
			ELSE c.expense_type::text
		END)::text as expense_type_name, 
		COALESCE(SUM(e.amount), 0) as total_amount, 
		COUNT(e.id) as count`).
		Joins("JOIN categories c ON e.category_id = c.id").
		Where("e.user_id = ? AND e.date BETWEEN ? AND ? AND e.status IN ?", 
			userID, startDate, endDate, models.GetActiveStatuses()).
		Group("c.expense_type").
		Order("total_amount DESC").
		Scan(&expensesByType)
	
	if result.Error != nil {
		logger.Error("Error getting expenses by type: %v", result.Error)
		return nil, result.Error
	}
	summary["by_expense_type"] = expensesByType
	
	// Top 10 categorías
	var expensesByCategory []struct {
		CategoryName    string  `json:"category_name"`
		ExpenseTypeName string  `json:"expense_type_name"`
		TotalAmount     float64 `json:"total_amount"`
		Count           int64   `json:"count"`
	}
	
	result = db.DB.Table("expenses e").
		Select(`c.name as category_name, 
		(CASE 
			WHEN c.expense_type = 'needs' THEN 'Needs'
			WHEN c.expense_type = 'wants' THEN 'Wants'
			WHEN c.expense_type = 'savings' THEN 'Savings'
			ELSE c.expense_type::text
		END)::text as expense_type_name, 
		COALESCE(SUM(e.amount), 0) as total_amount, 
		COUNT(e.id) as count`).
		Joins("JOIN categories c ON e.category_id = c.id").
		Where("e.user_id = ? AND e.date BETWEEN ? AND ? AND e.status IN ?", 
			userID, startDate, endDate, models.GetActiveStatuses()).
		Group("c.id, c.name, c.expense_type").
		Order("total_amount DESC").
		Limit(10).
		Scan(&expensesByCategory)
	
	if result.Error != nil {
		logger.Error("Error getting top categories: %v", result.Error)
		return nil, result.Error
	}
	summary["top_categories"] = expensesByCategory
	
	logger.Info("Expense summary calculated successfully for user %s", userID)
	return summary, nil
}

// GetMonthlyExpensesSummary gets monthly expenses summary for the user
func GetMonthlyExpensesSummary(userID string, year int, month int) (map[string]interface{}, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Último día del mes
	
	return GetExpensesSummaryByPeriod(userID, startDate, endDate)
}

// GetExpensesByExpenseType gets expenses grouped by expense type for budget validation
func GetExpensesByExpenseType(userID string, startDate, endDate time.Time) (map[string]float64, error) {
	var results []struct {
		ExpenseTypeName string  `json:"expense_type_name"`
		TotalAmount     float64 `json:"total_amount"`
	}
	
	result := db.DB.Table("expenses e").
		Select(`(CASE 
			WHEN c.expense_type = 'needs' THEN 'Needs'
			WHEN c.expense_type = 'wants' THEN 'Wants'
			WHEN c.expense_type = 'savings' THEN 'Savings'
			ELSE c.expense_type::text
		END)::text as expense_type_name, 
		COALESCE(SUM(e.amount), 0) as total_amount`).
		Joins("JOIN categories c ON e.category_id = c.id").
		Where("e.user_id = ? AND e.date BETWEEN ? AND ? AND e.status IN ?", 
			userID, startDate, endDate, models.GetActiveStatuses()).
		Group("c.expense_type").
		Scan(&results)
	
	if result.Error != nil {
		logger.Error("Error getting expenses by expense type: %v", result.Error)
		return nil, result.Error
	}
	
	// Convertir a mapa para fácil acceso
	expensesByType := make(map[string]float64)
	for _, item := range results {
		expensesByType[item.ExpenseTypeName] = item.TotalAmount
	}
	
	logger.Info("Expenses by expense type retrieved successfully for user %s", userID)
	return expensesByType, nil
}

// ValidateMonthlyBudgetCompliance validates if user is within budget for a month
func ValidateMonthlyBudgetCompliance(userID string, year int, month int) (map[string]interface{}, error) {
	var compliance map[string]interface{}
	compliance = make(map[string]interface{})
	
	// Obtener el presupuesto del mes
	monthYear := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	budget, err := GetActiveBudgetByMonthYear(userID, monthYear)
	if err != nil {
		logger.Error("No active budget found for %d-%02d", year, month)
		return nil, errors.New("no active budget found for this month")
	}
	
	// Obtener gastos reales del mes
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	
	expensesByType, err := GetExpensesByExpenseType(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	// Calcular compliance para cada tipo
	needsSpent := expensesByType["Needs"]
	wantsSpent := expensesByType["Wants"]
	savingsSpent := expensesByType["Savings"]
	
	compliance["budget"] = map[string]interface{}{
		"needs_budget":   budget.NeedsBudget,
		"wants_budget":   budget.WantsBudget,
		"savings_budget": budget.SavingsBudget,
	}
	
	compliance["actual"] = map[string]interface{}{
		"needs_spent":   needsSpent,
		"wants_spent":   wantsSpent,
		"savings_spent": savingsSpent,
	}
	
	compliance["compliance"] = map[string]interface{}{
		"needs_compliance":   ((budget.NeedsBudget - needsSpent) / budget.NeedsBudget) * 100,
		"wants_compliance":   ((budget.WantsBudget - wantsSpent) / budget.WantsBudget) * 100,
		"savings_compliance": ((budget.SavingsBudget - savingsSpent) / budget.SavingsBudget) * 100,
	}
	
	compliance["status"] = map[string]interface{}{
		"needs_over_budget":   needsSpent > budget.NeedsBudget,
		"wants_over_budget":   wantsSpent > budget.WantsBudget,
		"savings_under_goal":  savingsSpent < budget.SavingsBudget,
	}
	
	totalBudget := budget.NeedsBudget + budget.WantsBudget + budget.SavingsBudget
	totalSpent := needsSpent + wantsSpent + savingsSpent
	
	compliance["overall"] = map[string]interface{}{
		"total_budget":     totalBudget,
		"total_spent":      totalSpent,
		"remaining_budget": totalBudget - totalSpent,
		"budget_used_pct":  (totalSpent / totalBudget) * 100,
	}
	
	logger.Info("Budget compliance calculated successfully for user %s", userID)
	return compliance, nil
}

// GetSpendingTrends gets spending trends over time for the user
func GetSpendingTrends(userID string, months int) (map[string]interface{}, error) {
	var trends map[string]interface{}
	trends = make(map[string]interface{})
	
	// Calcular fechas
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)
	
	// Gastos por mes
	var monthlyTrends []struct {
		Month       string  `json:"month"`
		TotalAmount float64 `json:"total_amount"`
		Count       int64   `json:"count"`
	}
	
	result := db.DB.Table("expenses").
		Select("TO_CHAR(date, 'YYYY-MM') as month, COALESCE(SUM(amount), 0) as total_amount, COUNT(id) as count").
		Where("user_id = ? AND date >= ? AND status IN ?", 
			userID, startDate, models.GetActiveStatuses()).
		Group("TO_CHAR(date, 'YYYY-MM')").
		Order("month ASC").
		Scan(&monthlyTrends)
	
	if result.Error != nil {
		logger.Error("Error getting monthly trends: %v", result.Error)
		return nil, result.Error
	}
	trends["monthly_trends"] = monthlyTrends
	
	// Tendencias por tipo de gasto
	var typesTrends []struct {
		Month           string  `json:"month"`
		ExpenseTypeName string  `json:"expense_type_name"`
		TotalAmount     float64 `json:"total_amount"`
	}
	
	result = db.DB.Table("expenses e").
		Select(`TO_CHAR(e.date, 'YYYY-MM') as month, 
		(CASE 
			WHEN c.expense_type = 'needs' THEN 'Needs'
			WHEN c.expense_type = 'wants' THEN 'Wants'
			WHEN c.expense_type = 'savings' THEN 'Savings'
			ELSE c.expense_type::text
		END)::text as expense_type_name, 
		COALESCE(SUM(e.amount), 0) as total_amount`).
		Joins("JOIN categories c ON e.category_id = c.id").
		Where("e.user_id = ? AND e.date >= ? AND e.status IN ?", 
			userID, startDate, models.GetActiveStatuses()).
		Group("TO_CHAR(e.date, 'YYYY-MM'), c.expense_type").
		Order("month ASC, expense_type_name").
		Scan(&typesTrends)
	
	if result.Error != nil {
		logger.Error("Error getting trends by type: %v", result.Error)
		return nil, result.Error
	}
	trends["trends_by_type"] = typesTrends
	
	logger.Info("Spending trends calculated successfully for user %s", userID)
	return trends, nil
}

// GetExpenseAnalyticsForML gets data formatted for ML analysis
func GetExpenseAnalyticsForML(userID string, months int) (map[string]interface{}, error) {
	var analytics map[string]interface{}
	analytics = make(map[string]interface{})
	
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)
	
	// Obtener todos los gastos del período para análisis detallado
	expenses, err := GetExpensesByDateRange(userID, startDate, endDate, false)
	if err != nil {
		return nil, err
	}
	
	// Preparar datos para ML
	var mlData []map[string]interface{}
	for _, expense := range expenses {
		mlData = append(mlData, map[string]interface{}{
			"amount":            expense.Amount,
			"date":              expense.Date,
			"day_of_week":       int(expense.Date.Weekday()),
			"month":             int(expense.Date.Month()),
			"category_name":     expense.Category.Name,
			"expense_type_name": models.GetExpenseTypeName(expense.Category.ExpenseType),
			"description":       expense.Description,
		})
	}
	
	analytics["raw_data"] = mlData
	analytics["total_records"] = len(mlData)
	analytics["period_start"] = startDate
	analytics["period_end"] = endDate
	
	// Estadísticas agregadas para features
	analytics["features"] = map[string]interface{}{
		"avg_daily_spending":   calculateAverageDaily(expenses),
		"spending_volatility":  calculateSpendingVolatility(expenses),
		"most_active_day":      getMostActiveDay(expenses),
		"category_diversity":   getCategoryDiversity(expenses),
		"largest_expense":      getLargestExpense(expenses),
		"typical_expense_size": getTypicalExpenseSize(expenses),
	}
	
	logger.Info("ML analytics prepared successfully for user %s", userID)
	return analytics, nil
}

// Helper functions for ML analytics
func calculateAverageDaily(expenses []models.Expense) float64 {
	if len(expenses) == 0 {
		return 0
	}
	
	total := 0.0
	for _, expense := range expenses {
		total += expense.Amount
	}
	
	// Calcular días únicos
	days := make(map[string]bool)
	for _, expense := range expenses {
		days[expense.Date.Format("2006-01-02")] = true
	}
	
	if len(days) == 0 {
		return 0
	}
	
	return total / float64(len(days))
}

func calculateSpendingVolatility(expenses []models.Expense) float64 {
	if len(expenses) < 2 {
		return 0
	}
	
	// Calculate the mean
	total := 0.0
	for _, expense := range expenses {
		total += expense.Amount
	}
	mean := total / float64(len(expenses))
	
	variance := 0.0
	for _, expense := range expenses {
		variance += (expense.Amount - mean) * (expense.Amount - mean)
	}
	variance /= float64(len(expenses))
	
	return variance // Variance as a measure of volatility
}

func getMostActiveDay(expenses []models.Expense) int {
	dayCount := make(map[int]int)
	for _, expense := range expenses {
		dayCount[int(expense.Date.Weekday())]++
	}
	
	maxCount := 0
	mostActiveDay := 0
	for day, count := range dayCount {
		if count > maxCount {
			maxCount = count
			mostActiveDay = day
		}
	}
	
	return mostActiveDay
}

func getCategoryDiversity(expenses []models.Expense) int {
	categories := make(map[string]bool)
	for _, expense := range expenses {
		categories[expense.Category.Name] = true
	}
	return len(categories)
}

func getLargestExpense(expenses []models.Expense) float64 {
	largest := 0.0
	for _, expense := range expenses {
		if expense.Amount > largest {
			largest = expense.Amount
		}
	}
	return largest
}

func getTypicalExpenseSize(expenses []models.Expense) float64 {
	if len(expenses) == 0 {
		return 0
	}
	
	// Calculate median as a measure of "typical"
	amounts := make([]float64, len(expenses))
	for i, expense := range expenses {
		amounts[i] = expense.Amount
	}
	
	// Sort to find median (simple implementation)
	for i := 0; i < len(amounts); i++ {
		for j := i + 1; j < len(amounts); j++ {
			if amounts[i] > amounts[j] {
				amounts[i], amounts[j] = amounts[j], amounts[i]
			}
		}
	}
	
	mid := len(amounts) / 2
	if len(amounts)%2 == 0 {
		return (amounts[mid-1] + amounts[mid]) / 2
	}
	return amounts[mid]
}

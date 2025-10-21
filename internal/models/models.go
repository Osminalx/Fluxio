package models

// GetAllModels returns all model structs for database migration
func GetAllModels() []interface{} {
	return []interface{}{
		&User{},
		&BankAccount{},
		// ExpenseType is now an enum (needs/wants/savings) - no longer a DB table
		&Category{},
		&FixedExpense{},
		&Goal{},
		&Expense{},
		&Income{},
		&Budget{},
		&Transfer{},
		&Reminder{},
		&BudgetHistory{},
		&RefreshToken{},
	}
}

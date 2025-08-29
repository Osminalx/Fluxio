package models

// GetAllModels returns all model structs for database migration
func GetAllModels() []interface{} {
	return []interface{}{
		&User{},
		&BankAccount{},
		&ExpenseType{},
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

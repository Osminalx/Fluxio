package models

// ExpenseType represents the fixed expense types for the 50/30/20 budgeting philosophy
type ExpenseType string

const (
	ExpenseTypeNeeds   ExpenseType = "needs"   // 50% - Essential expenses
	ExpenseTypeWants   ExpenseType = "wants"   // 30% - Non-essential expenses
	ExpenseTypeSavings ExpenseType = "savings" // 20% - Savings and investments
)

// ValidExpenseTypes returns all valid expense types
func ValidExpenseTypes() []ExpenseType {
	return []ExpenseType{
		ExpenseTypeNeeds,
		ExpenseTypeWants,
		ExpenseTypeSavings,
	}
}

// IsValidExpenseType checks if a given string is a valid expense type
func IsValidExpenseType(expenseType string) bool {
	switch ExpenseType(expenseType) {
	case ExpenseTypeNeeds, ExpenseTypeWants, ExpenseTypeSavings:
		return true
	default:
		return false
	}
}

// GetExpenseTypeName returns the display name for an expense type
func GetExpenseTypeName(expenseType ExpenseType) string {
	switch expenseType {
	case ExpenseTypeNeeds:
		return "Needs"
	case ExpenseTypeWants:
		return "Wants"
	case ExpenseTypeSavings:
		return "Savings"
	default:
		return string(expenseType)
	}
}

// String returns the string representation of the expense type
func (e ExpenseType) String() string {
	return string(e)
}


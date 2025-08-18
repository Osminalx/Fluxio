package services

import (
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

// GetBudgetHistoryByID gets a specific entry from the budget history
func GetBudgetHistoryByID(userID string, id string) (*models.BudgetHistory, error) {
	var history models.BudgetHistory
	result := db.DB.Where("user_id = ? AND id = ?", userID, id).First(&history)
	if result.Error != nil {
		logger.Error("Error getting budget history by id: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Budget history retrieved successfully: %+v", history)
	return &history, nil
}

// GetBudgetHistoryByBudgetID gets all the history of a specific budget
func GetBudgetHistoryByBudgetID(userID string, budgetID string) ([]models.BudgetHistory, error) {
	var histories []models.BudgetHistory
	result := db.DB.Where("user_id = ? AND budget_id = ?", userID, budgetID).
		Order("changed_at DESC").Find(&histories)
	if result.Error != nil {
		logger.Error("Error getting budget history by budget ID: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Budget history for budget retrieved successfully: %+v", histories)
	return histories, nil
}

// GetAllBudgetHistory gets all the budget history for the user
func GetAllBudgetHistory(userID string) ([]models.BudgetHistory, error) {
	var histories []models.BudgetHistory
	result := db.DB.Where("user_id = ?", userID).
		Order("changed_at DESC").Find(&histories)
	if result.Error != nil {
		logger.Error("Error getting all budget history: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("All budget history retrieved successfully: %+v", histories)
	return histories, nil
}

// GetBudgetHistoryByDateRange gets budget history in a date range
func GetBudgetHistoryByDateRange(userID string, startDate, endDate time.Time) ([]models.BudgetHistory, error) {
	var histories []models.BudgetHistory
	result := db.DB.Where("user_id = ? AND changed_at BETWEEN ? AND ?", userID, startDate, endDate).
		Order("changed_at DESC").Find(&histories)
	if result.Error != nil {
		logger.Error("Error getting budget history by date range: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Budget history by date range retrieved successfully: %+v", histories)
	return histories, nil
}

// GetBudgetHistoryWithReasons gets budget history filtered by change reason
func GetBudgetHistoryWithReasons(userID string, reasonFilter string) ([]models.BudgetHistory, error) {
	var histories []models.BudgetHistory
	result := db.DB.Where("user_id = ? AND change_reason ILIKE ?", userID, "%"+reasonFilter+"%").
		Order("changed_at DESC").Find(&histories)
	if result.Error != nil {
		logger.Error("Error getting budget history with reasons: %v", result.Error)
		return nil, result.Error
	}
	
	logger.Info("Budget history with reasons retrieved successfully: %+v", histories)
	return histories, nil
}

// GetBudgetHistoryStats gets statistics from the history for analysis
func GetBudgetHistoryStats(userID string) (map[string]interface{}, error) {
	var stats map[string]interface{}
	stats = make(map[string]interface{})
	
	// Count total changes
	var totalChanges int64
	result := db.DB.Model(&models.BudgetHistory{}).Where("user_id = ?", userID).Count(&totalChanges)
	if result.Error != nil {
		logger.Error("Error counting budget history: %v", result.Error)
		return nil, result.Error
	}
	stats["total_changes"] = totalChanges
	
	// Get the date of the first and last change
	var firstChange, lastChange models.BudgetHistory
	
	// First change
	db.DB.Where("user_id = ?", userID).Order("changed_at ASC").First(&firstChange)
	if firstChange.ID != uuid.Nil {
		stats["first_change_date"] = firstChange.ChangedAt
	}
	
	// Last change
	db.DB.Where("user_id = ?", userID).Order("changed_at DESC").First(&lastChange)
	if lastChange.ID != uuid.Nil {
		stats["last_change_date"] = lastChange.ChangedAt
	}
	
	// Count changes by month (last 12 months)
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	var monthlyChanges []struct {
		Month  string `json:"month"`
		Count  int64  `json:"count"`
	}
	
	result = db.DB.Table("budget_histories").
		Select("TO_CHAR(changed_at, 'YYYY-MM') as month, COUNT(*) as count").
		Where("user_id = ? AND changed_at >= ?", userID, oneYearAgo).
		Group("TO_CHAR(changed_at, 'YYYY-MM')").
		Order("month DESC").
		Scan(&monthlyChanges)
	
	if result.Error != nil {
		logger.Error("Error getting monthly budget history stats: %v", result.Error)
		return nil, result.Error
	}
	
	stats["monthly_changes"] = monthlyChanges
	
	logger.Info("Budget history stats retrieved successfully: %+v", stats)
	return stats, nil
}

// AnalyzeBudgetPatterns analyzes patterns in budget changes for ML
func AnalyzeBudgetPatterns(userID string, months int) (map[string]interface{}, error) {
	var patterns map[string]interface{}
	patterns = make(map[string]interface{})
	
	startDate := time.Now().AddDate(0, -months, 0)
	
	// Get all changes in the period
	var histories []models.BudgetHistory
	result := db.DB.Where("user_id = ? AND changed_at >= ?", userID, startDate).
		Order("changed_at ASC").Find(&histories)
	if result.Error != nil {
		logger.Error("Error getting budget history for pattern analysis: %v", result.Error)
		return nil, result.Error
	}
	
	// Trend analysis
	var needsChanges, wantsChanges, savingsChanges []float64
	var changeFrequency []time.Duration
	var lastChangeTime *time.Time
	
	for _, history := range histories {
		// Calculate changes in each category
		if history.OldNeedsBudget != nil && history.NewNeedsBudget != nil {
			needsChanges = append(needsChanges, *history.NewNeedsBudget - *history.OldNeedsBudget)
		}
		if history.OldWantsBudget != nil && history.NewWantsBudget != nil {
			wantsChanges = append(wantsChanges, *history.NewWantsBudget - *history.OldWantsBudget)
		}
		if history.OldSavingsBudget != nil && history.NewSavingsBudget != nil {
			savingsChanges = append(savingsChanges, *history.NewSavingsBudget - *history.OldSavingsBudget)
		}
		
		// Calculate change frequency
		if lastChangeTime != nil {
			changeFrequency = append(changeFrequency, history.ChangedAt.Sub(*lastChangeTime))
		}
		lastChangeTime = &history.ChangedAt
	}
	
	patterns["needs_changes"] = needsChanges
	patterns["wants_changes"] = wantsChanges
	patterns["savings_changes"] = savingsChanges
	patterns["change_frequency"] = changeFrequency
	patterns["total_analyzed_period_months"] = months
	patterns["analyzed_changes_count"] = len(histories)
	
	logger.Info("Budget patterns analyzed successfully for user %s: %+v", userID, patterns)
	return patterns, nil
}

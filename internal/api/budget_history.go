package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// Request and response structures
type BudgetHistoryResponse struct {
	ID                string   `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	BudgetID          string   `json:"budget_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	OldNeedsBudget    *float64 `json:"old_needs_budget,omitempty" example:"2500.00"`
	NewNeedsBudget    *float64 `json:"new_needs_budget,omitempty" example:"2700.00"`
	OldWantsBudget    *float64 `json:"old_wants_budget,omitempty" example:"1500.00"`
	NewWantsBudget    *float64 `json:"new_wants_budget,omitempty" example:"1300.00"`
	OldSavingsBudget  *float64 `json:"old_savings_budget,omitempty" example:"1000.00"`
	NewSavingsBudget  *float64 `json:"new_savings_budget,omitempty" example:"1200.00"`
	ChangeReason      *string  `json:"change_reason,omitempty" example:"Salary increase"`
	ChangedAt         string   `json:"changed_at" example:"2024-01-15T10:30:00Z"`
}

type BudgetHistoryListResponse struct {
	History []BudgetHistoryResponse `json:"history"`
	Count   int                     `json:"count" example:"10"`
}

type BudgetHistoryStatsResponse struct {
	TotalChanges     int64                    `json:"total_changes" example:"25"`
	FirstChangeDate  *string                  `json:"first_change_date,omitempty" example:"2024-01-01T00:00:00Z"`
	LastChangeDate   *string                  `json:"last_change_date,omitempty" example:"2024-12-15T10:30:00Z"`
	MonthlyChanges   []map[string]interface{} `json:"monthly_changes"`
}

type BudgetHistoryPatternsResponse struct {
	NeedsChanges             []float64 `json:"needs_changes"`
	WantsChanges             []float64 `json:"wants_changes"`
	SavingsChanges           []float64 `json:"savings_changes"`
	ChangeFrequencySeconds   []int64   `json:"change_frequency_seconds" example:"86400,172800,604800"`
	AnalyzedPeriodMonths     int       `json:"analyzed_period_months" example:"12"`
	AnalyzedChangesCount     int       `json:"analyzed_changes_count" example:"15"`
}

// Helper function to convert model to response
func convertBudgetHistoryToResponse(history *models.BudgetHistory) BudgetHistoryResponse {
	response := BudgetHistoryResponse{
		ID:               history.ID.String(),
		BudgetID:         history.BudgetID.String(),
		OldNeedsBudget:   history.OldNeedsBudget,
		NewNeedsBudget:   history.NewNeedsBudget,
		OldWantsBudget:   history.OldWantsBudget,
		NewWantsBudget:   history.NewWantsBudget,
		OldSavingsBudget: history.OldSavingsBudget,
		NewSavingsBudget: history.NewSavingsBudget,
		ChangeReason:     history.ChangeReason,
		ChangedAt:        history.ChangedAt.Format(time.RFC3339),
	}
	return response
}

// @Summary Get budget history by ID
// @Description Get a specific budget history entry by ID
// @Tags Budget History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget History ID"
// @Success 200 {object} BudgetHistoryResponse
// @Failure 400 {string} string "Budget history ID is required"
// @Failure 404 {string} string "Budget history not found"
// @Failure 500 {string} string "Internal server error"
// @Router /budget-history/{id} [get]
func GetBudgetHistoryByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Budget history ID is required", http.StatusBadRequest)
		return
	}

	history, err := services.GetBudgetHistoryByID(userID, id)
	if err != nil {
		logger.Error("Error getting budget history by ID: %v", err)
		http.Error(w, "Budget history not found", http.StatusNotFound)
		return
	}

	response := convertBudgetHistoryToResponse(history)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get budget history by budget ID
// @Description Get all history entries for a specific budget
// @Tags Budget History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param budget_id path string true "Budget ID"
// @Success 200 {object} BudgetHistoryListResponse
// @Failure 400 {string} string "Budget ID is required"
// @Failure 500 {string} string "Internal server error"
// @Router /budgets/{budget_id}/history [get]
func GetBudgetHistoryByBudgetID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	budgetID := r.PathValue("budget_id")

	if budgetID == "" {
		http.Error(w, "Budget ID is required", http.StatusBadRequest)
		return
	}

	histories, err := services.GetBudgetHistoryByBudgetID(userID, budgetID)
	if err != nil {
		logger.Error("Error getting budget history by budget ID: %v", err)
		http.Error(w, "Error retrieving budget history", http.StatusInternalServerError)
		return
	}

	var responseHistories []BudgetHistoryResponse
	for _, history := range histories {
		responseHistories = append(responseHistories, convertBudgetHistoryToResponse(&history))
	}

	response := BudgetHistoryListResponse{
		History: responseHistories,
		Count:   len(responseHistories),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get all budget history
// @Description Get all budget history entries for the user
// @Tags Budget History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BudgetHistoryListResponse
// @Failure 500 {string} string "Internal server error"
// @Router /budget-history [get]
func GetAllBudgetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	histories, err := services.GetAllBudgetHistory(userID)
	if err != nil {
		logger.Error("Error getting all budget history: %v", err)
		http.Error(w, "Error retrieving budget history", http.StatusInternalServerError)
		return
	}

	var responseHistories []BudgetHistoryResponse
	for _, history := range histories {
		responseHistories = append(responseHistories, convertBudgetHistoryToResponse(&history))
	}

	response := BudgetHistoryListResponse{
		History: responseHistories,
		Count:   len(responseHistories),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get budget history by date range
// @Description Get budget history entries within a specific date range
// @Tags Budget History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)" example:"2024-01-01"
// @Param end_date query string true "End date (YYYY-MM-DD)" example:"2024-12-31"
// @Success 200 {object} BudgetHistoryListResponse
// @Failure 400 {string} string "Date parameters are required or invalid"
// @Failure 500 {string} string "Internal server error"
// @Router /budget-history/date-range [get]
func GetBudgetHistoryByDateRange(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "Both start_date and end_date are required", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Set end date to end of day
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	histories, err := services.GetBudgetHistoryByDateRange(userID, startDate, endDate)
	if err != nil {
		logger.Error("Error getting budget history by date range: %v", err)
		http.Error(w, "Error retrieving budget history", http.StatusInternalServerError)
		return
	}

	var responseHistories []BudgetHistoryResponse
	for _, history := range histories {
		responseHistories = append(responseHistories, convertBudgetHistoryToResponse(&history))
	}

	response := BudgetHistoryListResponse{
		History: responseHistories,
		Count:   len(responseHistories),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get budget history with reasons filter
// @Description Get budget history entries filtered by change reason
// @Tags Budget History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reason query string true "Change reason filter" example:"salary"
// @Success 200 {object} BudgetHistoryListResponse
// @Failure 400 {string} string "Reason filter is required"
// @Failure 500 {string} string "Internal server error"
// @Router /budget-history/reasons [get]
func GetBudgetHistoryWithReasons(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	reasonFilter := r.URL.Query().Get("reason")

	if reasonFilter == "" {
		http.Error(w, "Reason filter is required", http.StatusBadRequest)
		return
	}

	histories, err := services.GetBudgetHistoryWithReasons(userID, reasonFilter)
	if err != nil {
		logger.Error("Error getting budget history with reasons: %v", err)
		http.Error(w, "Error retrieving budget history", http.StatusInternalServerError)
		return
	}

	var responseHistories []BudgetHistoryResponse
	for _, history := range histories {
		responseHistories = append(responseHistories, convertBudgetHistoryToResponse(&history))
	}

	response := BudgetHistoryListResponse{
		History: responseHistories,
		Count:   len(responseHistories),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get budget history statistics
// @Description Get statistical information about budget history
// @Tags Budget History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BudgetHistoryStatsResponse
// @Failure 500 {string} string "Internal server error"
// @Router /budget-history/stats [get]
func GetBudgetHistoryStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	stats, err := services.GetBudgetHistoryStats(userID)
	if err != nil {
		logger.Error("Error getting budget history stats: %v", err)
		http.Error(w, "Error retrieving budget history statistics", http.StatusInternalServerError)
		return
	}

	response := BudgetHistoryStatsResponse{
		TotalChanges:   stats["total_changes"].(int64),
		MonthlyChanges: stats["monthly_changes"].([]map[string]interface{}),
	}

	if firstDate, ok := stats["first_change_date"].(time.Time); ok {
		formattedDate := firstDate.Format(time.RFC3339)
		response.FirstChangeDate = &formattedDate
	}

	if lastDate, ok := stats["last_change_date"].(time.Time); ok {
		formattedDate := lastDate.Format(time.RFC3339)
		response.LastChangeDate = &formattedDate
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Analyze budget patterns
// @Description Analyze patterns in budget changes for machine learning insights
// @Tags Budget History
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param months query int false "Number of months to analyze" default:12 example:12
// @Success 200 {object} BudgetHistoryPatternsResponse
// @Failure 400 {string} string "Invalid months parameter"
// @Failure 500 {string} string "Internal server error"
// @Router /budget-history/patterns [get]
func AnalyzeBudgetPatterns(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	monthsStr := r.URL.Query().Get("months")
	months := 12 // Default to 12 months
	
	if monthsStr != "" {
		var err error
		months, err = strconv.Atoi(monthsStr)
		if err != nil || months <= 0 {
			http.Error(w, "Invalid months parameter. Must be a positive integer", http.StatusBadRequest)
			return
		}
	}

	patterns, err := services.AnalyzeBudgetPatterns(userID, months)
	if err != nil {
		logger.Error("Error analyzing budget patterns: %v", err)
		http.Error(w, "Error analyzing budget patterns", http.StatusInternalServerError)
		return
	}

	// Convert time.Duration to seconds for JSON serialization
	changeFrequency := patterns["change_frequency"].([]time.Duration)
	changeFrequencySeconds := make([]int64, len(changeFrequency))
	for i, duration := range changeFrequency {
		changeFrequencySeconds[i] = int64(duration.Seconds())
	}

	response := BudgetHistoryPatternsResponse{
		NeedsChanges:           patterns["needs_changes"].([]float64),
		WantsChanges:           patterns["wants_changes"].([]float64),
		SavingsChanges:         patterns["savings_changes"].([]float64),
		ChangeFrequencySeconds: changeFrequencySeconds,
		AnalyzedPeriodMonths:   patterns["total_analyzed_period_months"].(int),
		AnalyzedChangesCount:   patterns["analyzed_changes_count"].(int),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
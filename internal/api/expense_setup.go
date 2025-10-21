package api

import (
	"encoding/json"
	"net/http"

	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// Response structures
type ExpenseTypeInfo struct {
	Value string `json:"value" example:"needs"`
	Name  string `json:"name" example:"Needs"`
}

type SystemOverviewResponse struct {
	ExpenseTypesCount int                    `json:"expense_types_count" example:"3"`
	ExpenseTypes      []ExpenseTypeInfo      `json:"expense_types"`
	SystemInfo        map[string]interface{} `json:"system_info"`
}

// @Summary Initialize expense system
// @Description Initialize the basic expense system with default expense types (Admin only)
// @Tags System Setup
// @Accept json
// @Produce json
// @Success 201 {object} SuccessResponse
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/setup/initialize [post]
func InitializeExpenseSystem(w http.ResponseWriter, r *http.Request) {
	err := services.InitializeExpenseSystem()
	if err != nil {
		logger.Error("Error initializing expense system: %v", err)
		http.Error(w, "Error initializing expense system", http.StatusInternalServerError)
		return
	}

	response := SuccessResponse{
		Message: "Expense system initialized successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// @Summary Setup new user
// @Description Create default categories for the authenticated user
// @Tags System Setup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} SuccessResponse
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/setup/user [post]
func SetupNewUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	err := services.SetupNewUser(userID)
	if err != nil {
		logger.Error("Error setting up new user: %v", err)
		http.Error(w, "Error setting up user", http.StatusInternalServerError)
		return
	}

	response := SuccessResponse{
		Message: "User setup completed successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get system overview
// @Description Get an overview of the expense system setup and configuration
// @Tags System Setup
// @Accept json
// @Produce json
// @Success 200 {object} SystemOverviewResponse
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/setup/overview [get]
func GetSystemOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := services.GetSystemOverview()
	if err != nil {
		logger.Error("Error getting system overview: %v", err)
		http.Error(w, "Error retrieving system overview", http.StatusInternalServerError)
		return
	}

	// Get expense types from overview (already formatted)
	expenseTypesRaw := overview["expense_types"].([]map[string]string)
	var expenseTypes []ExpenseTypeInfo
	for _, et := range expenseTypesRaw {
		expenseTypes = append(expenseTypes, ExpenseTypeInfo{
			Value: et["value"],
			Name:  et["name"],
		})
	}

	response := SystemOverviewResponse{
		ExpenseTypesCount: overview["expense_types_count"].(int),
		ExpenseTypes:      expenseTypes,
		SystemInfo:        overview["system_info"].(map[string]interface{}),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
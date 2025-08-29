package api

import (
	"encoding/json"
	"net/http"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// Response structures
type SimpleExpenseTypeResponse struct {
	ID          string                    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string                    `json:"name" example:"Needs"`
	Status      string                    `json:"status" example:"active"`
	CreatedAt   string                    `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   string                    `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	Categories  []UserCategoryResponse    `json:"categories,omitempty"`
}

type ExpenseTypesListResponse struct {
	ExpenseTypes []SimpleExpenseTypeResponse `json:"expense_types"`
	Count        int                         `json:"count" example:"3"`
}

// Helper function to convert model to response
func convertExpenseTypeToSimpleResponse(expenseType *models.ExpenseType) SimpleExpenseTypeResponse {
	response := SimpleExpenseTypeResponse{
		ID:        expenseType.ID.String(),
		Name:      expenseType.Name,
		Status:    string(expenseType.Status),
		CreatedAt: expenseType.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: expenseType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Convert categories if present
	if expenseType.Categories != nil {
		var categories []UserCategoryResponse
		for _, category := range expenseType.Categories {
			categoryResponse := UserCategoryResponse{
				ID:            category.ID.String(),
				Name:          category.Name,
				ExpenseTypeID: category.ExpenseTypeID.String(),
				Status:        string(category.Status),
				CreatedAt:     category.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:     category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}

			if category.StatusChangedAt != nil {
				statusChangedAt := category.StatusChangedAt.Format("2006-01-02T15:04:05Z")
				categoryResponse.StatusChangedAt = &statusChangedAt
			}

			categories = append(categories, categoryResponse)
		}
		response.Categories = categories
	}

	return response
}

// @Summary Get all expense types
// @Description Get all active expense types (Needs, Wants, Savings)
// @Tags Expense Types
// @Accept json
// @Produce json
// @Success 200 {object} ExpenseTypesListResponse
// @Failure 500 {string} string "Internal server error"
// @Router /expense-types [get]
func GetAllExpenseTypes(w http.ResponseWriter, r *http.Request) {
	expenseTypes, err := services.GetAllExpenseTypes()
	if err != nil {
		logger.Error("Error getting all expense types: %v", err)
		http.Error(w, "Error retrieving expense types", http.StatusInternalServerError)
		return
	}

	var responseTypes []SimpleExpenseTypeResponse
	for _, expenseType := range expenseTypes {
		responseTypes = append(responseTypes, convertExpenseTypeToSimpleResponse(&expenseType))
	}

	response := ExpenseTypesListResponse{
		ExpenseTypes: responseTypes,
		Count:        len(responseTypes),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get expense type by ID
// @Description Get a specific expense type by ID
// @Tags Expense Types
// @Accept json
// @Produce json
// @Param id path string true "Expense Type ID"
// @Success 200 {object} SimpleExpenseTypeResponse
// @Failure 400 {string} string "Expense type ID is required"
// @Failure 404 {string} string "Expense type not found"
// @Failure 500 {string} string "Internal server error"
// @Router /expense-types/{id} [get]
func GetExpenseTypeByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Expense type ID is required", http.StatusBadRequest)
		return
	}

	expenseType, err := services.GetExpenseTypeByID(id)
	if err != nil {
		logger.Error("Error getting expense type by ID: %v", err)
		http.Error(w, "Expense type not found", http.StatusNotFound)
		return
	}

	response := convertExpenseTypeToSimpleResponse(expenseType)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get expense type by name
// @Description Get a specific expense type by name (Needs, Wants, Savings)
// @Tags Expense Types
// @Accept json
// @Produce json
// @Param name path string true "Expense Type Name" Enums(Needs, Wants, Savings)
// @Success 200 {object} SimpleExpenseTypeResponse
// @Failure 400 {string} string "Expense type name is required"
// @Failure 404 {string} string "Expense type not found"
// @Failure 500 {string} string "Internal server error"
// @Router /expense-types/name/{name} [get]
func GetExpenseTypeByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	if name == "" {
		http.Error(w, "Expense type name is required", http.StatusBadRequest)
		return
	}

	expenseType, err := services.GetExpenseTypeByName(name)
	if err != nil {
		logger.Error("Error getting expense type by name: %v", err)
		http.Error(w, "Expense type not found", http.StatusNotFound)
		return
	}

	response := convertExpenseTypeToSimpleResponse(expenseType)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get expense types with user categories
// @Description Get all expense types with the authenticated user's categories loaded
// @Tags Expense Types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ExpenseTypesListResponse
// @Failure 500 {string} string "Internal server error"
// @Router /expense-types/with-categories [get]
func GetExpenseTypesWithUserCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	expenseTypes, err := services.GetExpenseTypesWithUserCategories(userID)
	if err != nil {
		logger.Error("Error getting expense types with user categories: %v", err)
		http.Error(w, "Error retrieving expense types with categories", http.StatusInternalServerError)
		return
	}

	var responseTypes []SimpleExpenseTypeResponse
	for _, expenseType := range expenseTypes {
		responseTypes = append(responseTypes, convertExpenseTypeToSimpleResponse(&expenseType))
	}

	response := ExpenseTypesListResponse{
		ExpenseTypes: responseTypes,
		Count:        len(responseTypes),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
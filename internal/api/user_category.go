package api

import (
	"encoding/json"
	"net/http"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// Request and response structures
type CreateUserCategoryRequest struct {
	Name        string `json:"name" example:"Groceries"`
	ExpenseType string `json:"expense_type" example:"needs" enums:"needs,wants,savings"`
}

type UpdateUserCategoryRequest struct {
	Name        *string `json:"name,omitempty" example:"Groceries Updated"`
	ExpenseType *string `json:"expense_type,omitempty" example:"needs" enums:"needs,wants,savings"`
}

type UserCategoryResponse struct {
	ID              string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name            string  `json:"name" example:"Groceries"`
	ExpenseType     string  `json:"expense_type" example:"needs" enums:"needs,wants,savings"`
	ExpenseTypeName string  `json:"expense_type_name" example:"Needs"`
	Status          string  `json:"status" example:"active"`
	StatusChangedAt *string `json:"status_changed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt       string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       string  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type UserCategoriesListResponse struct {
	Categories []UserCategoryResponse `json:"categories"`
	Count      int                    `json:"count" example:"15"`
}

type UserCategoriesGroupedResponse struct {
	GroupedCategories map[string][]UserCategoryResponse `json:"grouped_categories"`
	TotalCount        int                               `json:"total_count" example:"15"`
}

type UserCategoryStatsResponse struct {
	TotalCategories    int64            `json:"total_categories" example:"15"`
	CategoriesByType   map[string]int64 `json:"categories_by_type"`
	DeletedCategories  int64            `json:"deleted_categories" example:"2"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

// Helper functions to convert models to responses
func convertUserCategoryToResponse(category *models.Category) UserCategoryResponse {
	response := UserCategoryResponse{
		ID:              category.ID.String(),
		Name:            category.Name,
		ExpenseType:     string(category.ExpenseType),
		ExpenseTypeName: models.GetExpenseTypeName(category.ExpenseType),
		Status:          string(category.Status),
		CreatedAt:       category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if category.StatusChangedAt != nil {
		statusChangedAt := category.StatusChangedAt.Format("2006-01-02T15:04:05Z")
		response.StatusChangedAt = &statusChangedAt
	}

	return response
}

// @Summary Create user category
// @Description Create a new category for the authenticated user
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body CreateUserCategoryRequest true "Category data"
// @Success 201 {object} UserCategoryResponse
// @Failure 400 {string} string "Invalid request body or missing required fields"
// @Failure 409 {string} string "Category already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories [post]
func CreateUserCategory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var req CreateUserCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	if req.ExpenseType == "" {
		http.Error(w, "Expense type is required", http.StatusBadRequest)
		return
	}

	// Validate expense type
	if !models.IsValidExpenseType(req.ExpenseType) {
		http.Error(w, "Invalid expense type. Must be one of: needs, wants, savings", http.StatusBadRequest)
		return
	}

	category := &models.Category{
		Name:        req.Name,
		ExpenseType: models.ExpenseType(req.ExpenseType),
	}

	if err := services.CreateUserCategory(userID, category); err != nil {
		logger.Error("Error creating user category: %v", err)
		if err.Error() == "you already have a category with this name in this expense type" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err.Error() == "invalid expense type. Must be one of: needs, wants, savings" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Error creating category", http.StatusInternalServerError)
		return
	}

	// Get the created category with relations
	createdCategory, err := services.GetUserCategoryByID(userID, category.ID.String())
	if err != nil {
		logger.Error("Error retrieving created category: %v", err)
		http.Error(w, "Category created but error retrieving details", http.StatusInternalServerError)
		return
	}

	response := convertUserCategoryToResponse(createdCategory)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user category by ID
// @Description Get a specific user category by ID
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} UserCategoryResponse
// @Failure 400 {string} string "Category ID is required"
// @Failure 404 {string} string "Category not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/{id} [get]
func GetUserCategoryByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	category, err := services.GetUserCategoryByID(userID, id)
	if err != nil {
		logger.Error("Error getting user category by ID: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	response := convertUserCategoryToResponse(category)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get all user categories
// @Description Get all categories for the authenticated user
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param include_deleted query bool false "Include deleted categories" default:false
// @Success 200 {object} UserCategoriesListResponse
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories [get]
func GetUserCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	categories, err := services.GetUserCategories(userID, includeDeleted)
	if err != nil {
		logger.Error("Error getting user categories: %v", err)
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	var responseCategories []UserCategoryResponse
	for _, category := range categories {
		responseCategories = append(responseCategories, convertUserCategoryToResponse(&category))
	}

	response := UserCategoriesListResponse{
		Categories: responseCategories,
		Count:      len(responseCategories),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user categories by expense type
// @Description Get user categories for a specific expense type
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param expense_type path string true "Expense Type" enums:"needs,wants,savings"
// @Param include_deleted query bool false "Include deleted categories" default:false
// @Success 200 {object} UserCategoriesListResponse
// @Failure 400 {string} string "Expense type is required or invalid"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/expense-type/{expense_type} [get]
func GetUserCategoriesByExpenseType(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	expenseType := r.PathValue("expense_type")
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	if expenseType == "" {
		http.Error(w, "Expense type is required", http.StatusBadRequest)
		return
	}

	// Validate expense type
	if !models.IsValidExpenseType(expenseType) {
		http.Error(w, "Invalid expense type. Must be one of: needs, wants, savings", http.StatusBadRequest)
		return
	}

	categories, err := services.GetUserCategoriesByExpenseType(userID, expenseType, includeDeleted)
	if err != nil {
		logger.Error("Error getting user categories by expense type: %v", err)
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	var responseCategories []UserCategoryResponse
	for _, category := range categories {
		responseCategories = append(responseCategories, convertUserCategoryToResponse(&category))
	}

	response := UserCategoriesListResponse{
		Categories: responseCategories,
		Count:      len(responseCategories),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user categories by expense type name
// @Description Get user categories for a specific expense type by name (Needs, Wants, Savings)
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param expense_type_name path string true "Expense Type Name" Enums(Needs, Wants, Savings)
// @Success 200 {object} UserCategoriesListResponse
// @Failure 400 {string} string "Expense type name is required"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/expense-type-name/{expense_type_name} [get]
func GetUserCategoriesByExpenseTypeName(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	expenseTypeName := r.PathValue("expense_type_name")

	if expenseTypeName == "" {
		http.Error(w, "Expense type name is required", http.StatusBadRequest)
		return
	}

	categories, err := services.GetUserCategoriesByExpenseTypeName(userID, expenseTypeName)
	if err != nil {
		logger.Error("Error getting user categories by expense type name: %v", err)
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	var responseCategories []UserCategoryResponse
	for _, category := range categories {
		responseCategories = append(responseCategories, convertUserCategoryToResponse(&category))
	}

	response := UserCategoriesListResponse{
		Categories: responseCategories,
		Count:      len(responseCategories),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user categories grouped by type
// @Description Get user categories grouped by expense type
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserCategoriesGroupedResponse
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/grouped [get]
func GetUserCategoriesGroupedByType(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	groupedCategories, err := services.GetUserCategoriesGroupedByType(userID)
	if err != nil {
		logger.Error("Error getting user categories grouped by type: %v", err)
		http.Error(w, "Error retrieving grouped categories", http.StatusInternalServerError)
		return
	}

	responseGrouped := make(map[string][]UserCategoryResponse)
	totalCount := 0

	for typeName, categories := range groupedCategories {
		var responseCategories []UserCategoryResponse
		for _, category := range categories {
			responseCategories = append(responseCategories, convertUserCategoryToResponse(&category))
		}
		responseGrouped[typeName] = responseCategories
		totalCount += len(responseCategories)
	}

	response := UserCategoriesGroupedResponse{
		GroupedCategories: responseGrouped,
		TotalCount:        totalCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Update user category
// @Description Update an existing user category
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param category body UpdateUserCategoryRequest true "Updated category data"
// @Success 200 {object} UserCategoryResponse
// @Failure 400 {string} string "Invalid request or missing required fields"
// @Failure 404 {string} string "Category not found"
// @Failure 409 {string} string "Category name already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/{id} [put]
func UpdateUserCategory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	var req UpdateUserCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing category
	existingCategory, err := services.GetUserCategoryByID(userID, id)
	if err != nil {
		logger.Error("Category not found for update: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// Prepare updated category
	updatedCategory := &models.Category{
		Name:        existingCategory.Name,
		ExpenseType: existingCategory.ExpenseType,
	}

	if req.Name != nil {
		updatedCategory.Name = *req.Name
	}

	if req.ExpenseType != nil {
		// Validate expense type
		if !models.IsValidExpenseType(*req.ExpenseType) {
			http.Error(w, "Invalid expense type. Must be one of: needs, wants, savings", http.StatusBadRequest)
			return
		}
		updatedCategory.ExpenseType = models.ExpenseType(*req.ExpenseType)
	}

	updatedCategoryResult, err := services.UpdateUserCategory(userID, id, updatedCategory)
	if err != nil {
		logger.Error("Error updating user category: %v", err)
		if err.Error() == "you already have a category with this name in this expense type" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err.Error() == "invalid expense type. Must be one of: needs, wants, savings" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Error updating category", http.StatusInternalServerError)
		return
	}

	response := convertUserCategoryToResponse(updatedCategoryResult)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Delete user category
// @Description Soft delete a user category
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} UserCategoryResponse
// @Failure 400 {string} string "Category ID is required"
// @Failure 404 {string} string "Category not found"
// @Failure 409 {string} string "Category has active expenses"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/{id} [delete]
func SoftDeleteUserCategory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	err := services.SoftDeleteUserCategory(userID, id)
	if err != nil {
		logger.Error("Error soft deleting user category: %v", err)
		if err.Error() == "cannot delete category: you have active expenses in this category" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err.Error() == "category not found, already deleted, or access denied" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Error deleting category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Restore user category
// @Description Restore a deleted user category
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} UserCategoryResponse
// @Failure 400 {string} string "Invalid request or category can't be restored"
// @Failure 404 {string} string "Category not found"
// @Failure 409 {string} string "Category name conflict"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/{id}/restore [post]
func RestoreUserCategory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	restoredCategory, err := services.RestoreUserCategory(userID, id)
	if err != nil {
		logger.Error("Error restoring user category: %v", err)
		if err.Error() == "cannot restore: you already have a category with this name in this expense type" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err.Error() == "category not found, not deleted, or access denied" ||
		   err.Error() == "cannot restore category: expense type is not valid" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Error restoring category", http.StatusInternalServerError)
		return
	}

	response := convertUserCategoryToResponse(restoredCategory)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Create default user categories
// @Description Create default categories for the authenticated user
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} SuccessResponse
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/defaults [post]
func CreateDefaultUserCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	err := services.CreateDefaultUserCategories(userID)
	if err != nil {
		logger.Error("Error creating default user categories: %v", err)
		http.Error(w, "Error creating default categories", http.StatusInternalServerError)
		return
	}

	response := SuccessResponse{
		Message: "Default categories created successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user category statistics
// @Description Get statistics about user's categories
// @Tags User Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserCategoryStatsResponse
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/user-categories/stats [get]
func GetUserCategoryStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	stats, err := services.GetUserCategoryStats(userID)
	if err != nil {
		logger.Error("Error getting user category stats: %v", err)
		http.Error(w, "Error retrieving category statistics", http.StatusInternalServerError)
		return
	}

	response := UserCategoryStatsResponse{
		TotalCategories:   stats["total_categories"].(int64),
		CategoriesByType:  stats["categories_by_type"].(map[string]int64),
		DeletedCategories: stats["deleted_categories"].(int64),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
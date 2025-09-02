package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

func createGoal(userID string, goal models.Goal) (*models.Goal, error) {
	// Force basic fields
	goal.UserID = uuid.MustParse(userID)
	goal.Status = models.StatusActive
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()

	result := db.DB.Create(&goal)
	if result.Error != nil {
		logger.Error("Error creating goal: %v", result.Error)
		return nil, errors.New("error creating goal")
	}

	return &goal, nil
}

func getGoalByID(userID string, goalID string) (*models.Goal, error) {
	var goal models.Goal
	result := db.DB.Where("user_id = ? AND id = ?", userID, goalID).First(&goal)
	if result.Error != nil {
		logger.Error("Error getting goal: %v", result.Error)
		return nil, errors.New("error getting goal")
	}

	return &goal, nil
}

func GetGoals(userID string, includeDeleted bool) ([]models.Goal, error) {
	var goals []models.Goal
	query := db.DB.Where("user_id = ?", userID)

	if !includeDeleted {
		query = query.Where("status = ?", models.StatusActive)
	}

	result := query.Find(&goals)

	if result.Error != nil {
		logger.Error("Error getting goals: %v", result.Error)
		return nil, errors.New("error getting goals")
	}

	return goals, nil
}

func updateGoal(userID string, goalID string, updates models.Goal) (*models.Goal, error) {
	// Verificar que el goal existe y pertenece al usuario
	existingGoal, err := getGoalByID(userID, goalID)
	if err != nil {
		return nil, err
	}

	// Preparar campos para actualizar
	updateData := map[string]interface{}{
		"updated_at": time.Now(),
	}

	// Solo actualizar campos que no estén vacíos
	if updates.Name != "" {
		updateData["name"] = updates.Name
	}
	if updates.TotalAmount > 0 {
		updateData["total_amount"] = updates.TotalAmount
	}
	if updates.SavedAmount >= 0 {
		updateData["saved_amount"] = updates.SavedAmount
	}

	// Actualizar en la base de datos
	result := db.DB.Model(existingGoal).Updates(updateData)
	if result.Error != nil {
		logger.Error("Error updating goal: %v", result.Error)
		return nil, errors.New("error updating goal")
	}

	// Obtener el goal actualizado
	updatedGoal, err := getGoalByID(userID, goalID)
	if err != nil {
		return nil, err
	}

	return updatedGoal, nil
}

func deleteGoal(userID string, goalID string) error {
	// Verificar que el goal existe y pertenece al usuario
	existingGoal, err := getGoalByID(userID, goalID)
	if err != nil {
		return err
	}

	// Soft delete - cambiar status a deleted
	now := time.Now()
	result := db.DB.Model(existingGoal).Updates(map[string]interface{}{
		"status":             models.StatusDeleted,
		"status_changed_at": &now,
		"updated_at":        now,
	})

	if result.Error != nil {
		logger.Error("Error deleting goal: %v", result.Error)
		return errors.New("error deleting goal")
	}

	return nil
}

func restoreGoal(userID string, goalID string) (*models.Goal, error) {
	// Verificar que el goal existe y pertenece al usuario
	var goal models.Goal
	result := db.DB.Where("user_id = ? AND id = ?", userID, goalID).First(&goal)
	if result.Error != nil {
		logger.Error("Error finding goal to restore: %v", result.Error)
		return nil, errors.New("goal not found")
	}

	// Restaurar - cambiar status a active
	now := time.Now()
	result = db.DB.Model(&goal).Updates(map[string]interface{}{
		"status":             models.StatusActive,
		"status_changed_at": &now,
		"updated_at":        now,
	})

	if result.Error != nil {
		logger.Error("Error restoring goal: %v", result.Error)
		return nil, errors.New("error restoring goal")
	}

	return &goal, nil
}

func changeGoalStatus(userID string, goalID string, newStatus models.Status) (*models.Goal, error) {
	// Verificar que el goal existe y pertenece al usuario
	existingGoal, err := getGoalByID(userID, goalID)
	if err != nil {
		return nil, err
	}

	// Actualizar status
	now := time.Now()
	result := db.DB.Model(existingGoal).Updates(map[string]interface{}{
		"status":             newStatus,
		"status_changed_at": &now,
		"updated_at":        now,
	})

	if result.Error != nil {
		logger.Error("Error changing goal status: %v", result.Error)
		return nil, errors.New("error changing goal status")
	}

	// Obtener el goal actualizado
	updatedGoal, err := getGoalByID(userID, goalID)
	if err != nil {
		return nil, err
	}

	return updatedGoal, nil
}

// Funciones públicas (exportadas)
func CreateGoal(userID string, goal models.Goal) (*models.Goal, error) {
	return createGoal(userID, goal)
}

func GetGoalByID(userID string, goalID string) (*models.Goal, error) {
	return getGoalByID(userID, goalID)
}

func UpdateGoal(userID string, goalID string, updates models.Goal) (*models.Goal, error) {
	return updateGoal(userID, goalID, updates)
}

func DeleteGoal(userID string, goalID string) error {
	return deleteGoal(userID, goalID)
}

func RestoreGoal(userID string, goalID string) (*models.Goal, error) {
	return restoreGoal(userID, goalID)
}

func ChangeGoalStatus(userID string, goalID string, newStatus models.Status) (*models.Goal, error) {
	return changeGoalStatus(userID, goalID, newStatus)
}

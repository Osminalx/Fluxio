package services

import (
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateFixedExpense creates a new fixed expense
func CreateFixedExpense(userID string,fixedExpense models.FixedExpense)(*models.FixedExpense,error){
	// Force basic Fields
	fixedExpense.UserID = uuid.MustParse(userID)
	fixedExpense.Status = models.StatusActive
	fixedExpense.CreatedAt = time.Now()
	fixedExpense.UpdatedAt = time.Now()

	result := db.DB.Create(&fixedExpense)
	if result.Error != nil {
		logger.Error("Error creating fixed expense: %v", result.Error)
		return nil,errors.New("error creating fixed expense")
	}

	return &fixedExpense,nil
}

// GetFixedExpenseByID returns a fixed expense by its ID
func GetFixedExpenseByID(userID string,id string)(*models.FixedExpense,error){
	var fixedExpense models.FixedExpense
	result := db.DB.Where("user_id = ? AND id = ?",userID,id).First(&fixedExpense)
	if result.Error != nil {
		logger.Error("Error getting fixed expense: %v", result.Error)
		return nil,errors.New("error getting fixed expense")
	}

	return &fixedExpense,nil
}

func GetFixedExpenses(userID string,includeDeleted bool)([]models.FixedExpense,error){
	var fixedExpenses []models.FixedExpense
	query := db.DB.Where("user_id = ?",userID)

	if !includeDeleted{
		query = query.Where("status = ?",models.StatusActive)
	}

	result := query.Find(&fixedExpenses)
	if result.Error != nil {
		logger.Error("Error getting fixed expenses: %v", result.Error)
		return nil,errors.New("error getting fixed expenses")
	}

	return fixedExpenses,nil
}

func UpdateFixedExpense(userID string,id string,fixedExpense models.FixedExpense)(*models.FixedExpense,error){
	var existingFixedExpense models.FixedExpense
	result := db.DB.Where("user_id = ? AND id = ?",userID,id).First(&existingFixedExpense)
	if result.Error != nil {
		logger.Error("Error getting fixed expense: %v", result.Error)
		return nil,errors.New("error getting fixed expense")
	}

	if existingFixedExpense.Status.IsDeleted(){
		logger.Error("Fixed expense is deleted")
		return nil,errors.New("fixed expense is deleted")
	}

	existingFixedExpense.Name = fixedExpense.Name
	existingFixedExpense.Amount = fixedExpense.Amount
	existingFixedExpense.DueDate = fixedExpense.DueDate
	existingFixedExpense.UpdatedAt = time.Now()

	result = db.DB.Save(&existingFixedExpense)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound{
		logger.Error("Error updating fixed expense: %v", result.Error)
		return nil,errors.New("error updating fixed expense")
	}

	return &existingFixedExpense,nil
}

func DeleteFixedExpense(userID string,id string)(*models.FixedExpense,error){
	var existingFixedExpense models.FixedExpense
	result := db.DB.Where("user_id = ? AND id = ?",userID,id).First(&existingFixedExpense)
	if result.Error != nil {
		logger.Error("Error getting fixed expense: %v", result.Error)
		return nil,errors.New("error getting fixed expense")
	}

	if existingFixedExpense.Status.IsDeleted(){
		logger.Error("Fixed expense is deleted")
		return nil,errors.New("fixed expense is deleted")
	}

	result = db.DB.Model(&existingFixedExpense).Update("status",models.StatusDeleted)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound{
		logger.Error("Error deleting fixed expense: %v", result.Error)
		return nil,errors.New("error deleting fixed expense")
	}

	return &existingFixedExpense,nil
}
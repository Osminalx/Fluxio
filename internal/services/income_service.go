package services

import (
	"errors"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

func CreateIncome(income *models.Income) error {
	result := db.DB.Create(income)
	if result.Error != nil{
		logger.Error("Error creating income: %v", result.Error)
		return result.Error
	}
	logger.Info("Income created successfully: %+v", income)
	return nil
}

func GetIncomeByID(id string) (*models.Income, error) {
	var income models.Income
	result := db.DB.Where("id = ?", id).First(&income)
	if result.Error != nil{
		logger.Error("Error getting income by id: %v", result.Error)
		return nil, result.Error
	}
	logger.Info("Income retrieved successfully: %+v", income)
	return &income, nil
}

func PatchIncome(id string, income *models.Income) (*models.Income, error) {
	var updatedIncome models.Income
	result := db.DB.Model(&models.Income{}).Where("id = ?", id).Updates(income)
	if result.Error != nil{
		logger.Error("Error patching income: %v", result.Error)
		return nil, result.Error
	}
	if result.RowsAffected == 0{
		logger.Error("Income not found")
		return nil, errors.New("income not found")
	}
	logger.Info("Income patched successfully: %+v", updatedIncome)
	return &updatedIncome, nil
}

func DeleteIncome(id string) error {
	result := db.DB.Delete(&models.Income{}, id)
	if result.Error != nil{
		logger.Error("Error deleting income: %v", result.Error)
		return result.Error
	}
	logger.Info("Income deleted successfully, id: %s", id)
	return nil
}
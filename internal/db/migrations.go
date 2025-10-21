package db

import (
	"fmt"
	"strings"

	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"gorm.io/gorm"
)

// createEnumTypes creates all PostgreSQL enum types needed by the application
// This must run BEFORE GORM auto-migration
func createEnumTypes(db *gorm.DB) error {
	// Create expense_type_enum
	if err := db.Exec("CREATE TYPE expense_type_enum AS ENUM ('needs', 'wants', 'savings')").Error; err != nil {
		// Check if error is because type already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("error creating expense_type_enum: %w", err)
		}
		logger.Info("✅ expense_type_enum already exists")
	} else {
		logger.Info("✅ Created PostgreSQL enum type: expense_type_enum")
	}
	
	return nil
}

// MigrateExpenseTypeToEnum migrates the categories table from ExpenseTypeID (UUID) to ExpenseType (enum)
// This should be run ONCE during deployment
func MigrateExpenseTypeToEnum(db *gorm.DB) error {
	logger.Info("🔄 Checking for ExpenseType data migration...")

	// Step 1: Check if migration is needed
	var needsMigration bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'categories' AND column_name = 'expense_type_id')").Scan(&needsMigration).Error; err != nil {
		return fmt.Errorf("error checking if migration is needed: %w", err)
	}

	if !needsMigration {
		logger.Info("✅ Migration not needed - expense_type_id column doesn't exist")
		// Still check if we need to convert varchar to enum
		return convertVarcharToEnum(db)
	}

	logger.Info("📝 Migration needed - proceeding with migration...")

	// Step 2: Add new expense_type column if it doesn't exist (temporarily as varchar)
	if err := db.Exec("ALTER TABLE categories ADD COLUMN IF NOT EXISTS expense_type VARCHAR(20)").Error; err != nil {
		return fmt.Errorf("error adding expense_type column: %w", err)
	}
	logger.Info("✅ Added expense_type column")

	// Step 3: Migrate data from expense_type_id to expense_type
	// This assumes you have the expense_types table with names: "Needs", "Wants", "Savings"
	query := `
		UPDATE categories c
		SET expense_type = CASE 
			WHEN LOWER(et.name) = 'needs' THEN 'needs'
			WHEN LOWER(et.name) = 'wants' THEN 'wants'
			WHEN LOWER(et.name) = 'savings' THEN 'savings'
			WHEN et.name = 'Needs' THEN 'needs'
			WHEN et.name = 'Wants' THEN 'wants'
			WHEN et.name = 'Savings' THEN 'savings'
			ELSE 'needs'
		END
		FROM expense_types et
		WHERE c.expense_type_id = et.id AND c.expense_type IS NULL
	`
	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("error migrating data: %w", err)
	}
	logger.Info("✅ Migrated data from expense_type_id to expense_type")

	// Step 4: Set default for any null values (shouldn't happen, but safety)
	if err := db.Exec("UPDATE categories SET expense_type = 'needs' WHERE expense_type IS NULL").Error; err != nil {
		return fmt.Errorf("error setting defaults: %w", err)
	}

	// Step 5: Make expense_type NOT NULL
	if err := db.Exec("ALTER TABLE categories ALTER COLUMN expense_type SET NOT NULL").Error; err != nil {
		return fmt.Errorf("error setting NOT NULL constraint: %w", err)
	}
	logger.Info("✅ Set expense_type as NOT NULL")

	// Step 6: Drop old foreign key constraint
	if err := db.Exec("ALTER TABLE categories DROP CONSTRAINT IF EXISTS fk_categories_expense_type").Error; err != nil {
		logger.Warn("Warning dropping foreign key constraint: %v (may not exist)", err)
	}

	// Step 7: Drop old expense_type_id column
	if err := db.Exec("ALTER TABLE categories DROP COLUMN IF EXISTS expense_type_id").Error; err != nil {
		return fmt.Errorf("error dropping expense_type_id column: %w", err)
	}
	logger.Info("✅ Dropped expense_type_id column")

	// Step 8: Convert varchar column to enum type
	if err := convertVarcharToEnum(db); err != nil {
		return fmt.Errorf("error converting to enum: %w", err)
	}

	// Step 9: Create index for better performance (optional)
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_categories_expense_type ON categories(expense_type)").Error; err != nil {
		logger.Warn("Warning creating index: %v", err)
	} else {
		logger.Info("✅ Created index on expense_type")
	}

	logger.Info("🎉 ExpenseType migration completed successfully!")
	return nil
}

// convertVarcharToEnum converts the expense_type column from varchar to PostgreSQL enum
func convertVarcharToEnum(db *gorm.DB) error {
	// Check if column is already enum type
	var dataType string
	err := db.Raw(`
		SELECT data_type 
		FROM information_schema.columns 
		WHERE table_name = 'categories' 
		AND column_name = 'expense_type'
	`).Scan(&dataType).Error
	
	if err != nil {
		return fmt.Errorf("error checking expense_type column type: %w", err)
	}

	if dataType == "USER-DEFINED" {
		logger.Info("✅ expense_type column is already a PostgreSQL enum")
		return nil
	}

	logger.Info("🔄 Converting expense_type from varchar to PostgreSQL enum...")

	// Convert varchar to enum
	if err := db.Exec(`
		ALTER TABLE categories 
		ALTER COLUMN expense_type 
		TYPE expense_type_enum 
		USING expense_type::expense_type_enum
	`).Error; err != nil {
		return fmt.Errorf("error converting expense_type to enum: %w", err)
	}

	logger.Info("✅ Converted expense_type column to PostgreSQL enum")
	return nil
}

// DropExpenseTypesTable drops the old expense_types table
// WARNING: This is destructive! Only run after confirming the migration worked
func DropExpenseTypesTable(db *gorm.DB) error {
	logger.Warn("⚠️  Dropping expense_types table...")
	
	// Check if table exists
	var exists bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'expense_types')").Scan(&exists).Error; err != nil {
		return fmt.Errorf("error checking if expense_types table exists: %w", err)
	}

	if !exists {
		logger.Info("✅ expense_types table doesn't exist - nothing to drop")
		return nil
	}

	if err := db.Exec("DROP TABLE IF EXISTS expense_types CASCADE").Error; err != nil {
		return fmt.Errorf("error dropping expense_types table: %w", err)
	}
	
	logger.Info("✅ Dropped expense_types table")
	return nil
}

// RunAllMigrations runs auto-migration for all models and custom migrations
func RunAllMigrations(db *gorm.DB) error {
	logger.Info("🔄 Running database migrations...")

	// Step 1: Create PostgreSQL enum type FIRST (before GORM needs it)
	logger.Info("Creating PostgreSQL enum types...")
	if err := createEnumTypes(db); err != nil {
		return fmt.Errorf("error creating enum types: %w", err)
	}

	// Step 2: Run GORM auto-migration for all models
	logger.Info("Running GORM auto-migration...")
	if err := db.AutoMigrate(models.GetAllModels()...); err != nil {
		return fmt.Errorf("error running auto-migration: %w", err)
	}
	logger.Info("✅ GORM auto-migration completed")

	// Step 3: Run custom migration for ExpenseType (data migration from old structure)
	logger.Info("Running custom ExpenseType migration...")
	if err := MigrateExpenseTypeToEnum(db); err != nil {
		return fmt.Errorf("error running ExpenseType migration: %w", err)
	}

	// Step 3: Optionally drop old expense_types table
	// Uncomment the lines below ONLY after verifying the migration worked correctly
	// logger.Info("Dropping old expense_types table...")
	// if err := DropExpenseTypesTable(db); err != nil {
	//     logger.Warn("Warning dropping expense_types table: %v", err)
	// }

	logger.Info("🎉 All migrations completed successfully!")
	return nil
}


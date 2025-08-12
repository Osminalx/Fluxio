package db

import (
	"fmt"
	"log"
	"os"

	"github.com/Osminalx/fluxio/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=fluxio port=5432 sslmode=disable"
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Auto migrate models
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Error migrating database:", err)
	}

	fmt.Println("✅ Conectado a Postgres con GORM")
}
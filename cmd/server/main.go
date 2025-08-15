// Package main Fluxio API Server
//
// Servidor de API para autenticación y gestión de usuarios
//
//	Schemes: http
//	Host: localhost:8080
//	BasePath: /
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- bearer
//
// swagger:meta
package main

import (
	"log"
	"net/http"

	"github.com/Osminalx/fluxio/docs"
	"github.com/Osminalx/fluxio/internal/api"
	"github.com/Osminalx/fluxio/internal/auth"
	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/middleware"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Fluxio API
// @version 1.0
// @description API de autenticación y gestión de usuarios con GORM y JWT
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
// @description Ingresa "Bearer" seguido de un espacio y el token JWT

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Swagger docs
	docs.SwaggerInfo.Title = "Fluxio API"
	docs.SwaggerInfo.Description = "API de autenticación y gestión de usuarios con GORM y JWT"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Connect to database
	logger.Info("🗄️  Conectando a la base de datos...")
	db.Connect()
	logger.Info("✅ Conectado a Postgres con GORM")

	// Create main router
	mux := http.NewServeMux()
	
	// We'll wrap the entire mux with logging middleware at the end

	// API v1 routes - PUBLIC (no authentication required)
	mux.HandleFunc("/api/v1/hello", api.HelloHandler)
	mux.HandleFunc("/api/v1/auth/login", api.LoginHandler)
	mux.HandleFunc("/api/v1/auth/register", api.RegisterHandler)

	// API v1 routes - PROTECTED (require authentication)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/protected", api.ProtectedHandler)
	
	// Apply auth middleware to protected API v1 routes
	mux.Handle("/api/v1/protected/", auth.AuthMiddleware(protectedMux))

	// Swagger UI (public access - no versioning needed)
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))
	
	// Health check endpoint (no versioning)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"1.0"}`))
	})

	logger.Info("🚀 Servidor iniciando en puerto 8080")
	logger.Info("📝 API v1 Endpoints disponibles:")
	logger.Info("  GET  /api/v1/hello - Endpoint público")
	logger.Info("  POST /api/v1/auth/login - Login")
	logger.Info("  POST /api/v1/auth/register - Registro")
	logger.Info("  GET  /api/v1/protected - Endpoint protegido (requiere JWT)")
	logger.Info("📚 Otros endpoints:")
	logger.Info("  GET  /health - Health check")
	logger.Info("  GET  /swagger/index.html - Swagger UI")

	// Apply logging middleware to all routes
	handler := middleware.LoggingMiddleware(mux)
	
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		logger.Fatal("Error al iniciar el servidor: %v", err)
	}
}
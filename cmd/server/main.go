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
	"fmt"
	"log"
	"net/http"

	"github.com/Osminalx/fluxio/internal/api"
	"github.com/Osminalx/fluxio/internal/auth"
	"github.com/Osminalx/fluxio/internal/db"
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

	// Connect to database
	db.Connect()

	// Create router
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/hello", api.HelloHandler)
	mux.HandleFunc("/auth/login", api.LoginHandler)
	mux.HandleFunc("/auth/register", api.RegisterHandler)

	// Swagger UI
	mux.HandleFunc("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Protected routes (require authentication)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/protected", api.ProtectedHandler)
	
	// Apply auth middleware to protected routes
	mux.Handle("/", auth.AuthMiddleware(protectedMux))

	fmt.Println("🚀 Server is running on port 8080")
	fmt.Println("📝 Available endpoints:")
	fmt.Println("  GET  /hello - Public endpoint")
	fmt.Println("  POST /auth/login - Login")
	fmt.Println("  POST /auth/register - Register")
	fmt.Println("  GET  /protected - Protected endpoint (requires JWT)")
	fmt.Println("  📚 Swagger UI: http://localhost:8080/swagger/index.html")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
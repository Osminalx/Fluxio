package api

import (
	"encoding/json"
	"net/http"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/Osminalx/fluxio/internal/services"
)

type LoginRequest struct {
	Email    string `json:"email" example:"usuario@ejemplo.com"`
	Password string `json:"password" example:"contraseña123"`
}

type RegisterRequest struct {
	Email    string `json:"email" example:"usuario@ejemplo.com"`
	Password string `json:"password" example:"contraseña123"`
	Name     string `json:"name" example:"Juan Pérez"`
}

type AuthResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  models.User `json:"user"`
}

// LoginHandler godoc
// @Summary Iniciar sesión
// @Description Autentica un usuario y devuelve un token JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciales de login"
// @Success 200 {object} AuthResponse
// @Failure 400 {string} string "Cuerpo de solicitud inválido"
// @Failure 401 {string} string "Credenciales inválidas"
// @Failure 500 {string} string "Error interno del servidor"
// @Router /auth/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := services.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !services.CheckPassword(req.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := services.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterHandler godoc
// @Summary Registrar usuario
// @Description Crea una nueva cuenta de usuario
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Datos del usuario"
// @Success 200 {object} AuthResponse
// @Failure 400 {string} string "Cuerpo de solicitud inválido"
// @Failure 409 {string} string "Usuario ya existe"
// @Failure 500 {string} string "Error interno del servidor"
// @Router /auth/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, _ := services.GetUserByEmail(req.Email)
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	result := db.DB.Create(&user)
	if result.Error != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	token, err := services.GenerateToken(&user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

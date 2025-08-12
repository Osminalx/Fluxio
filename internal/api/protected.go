package api

import (
	"encoding/json"
	"net/http"

	"github.com/Osminalx/fluxio/internal/auth"
	"github.com/Osminalx/fluxio/internal/services"
)

type ProtectedResponse struct {
	Message string `json:"message" example:"¡Acceso autorizado! Esta es una ruta protegida."`
	User    UserClaims `json:"user"`
}

// UserClaims es una versión simplificada para Swagger
type UserClaims struct {
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email  string `json:"email" example:"usuario@ejemplo.com"`
}

// ProtectedHandler godoc
// @Summary Endpoint protegido
// @Description Endpoint que requiere autenticación JWT
// @Tags protected
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} ProtectedResponse
// @Failure 401 {string} string "No autorizado"
// @Failure 500 {string} string "Error interno del servidor"
// @Router /protected [get]
func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by middleware)
	userClaims, ok := r.Context().Value(auth.UserContextKey).(*services.Claims)
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	// Convert to simplified type for response
	response := ProtectedResponse{
		Message: "¡Acceso autorizado! Esta es una ruta protegida.",
		User: UserClaims{
			UserID: userClaims.UserID,
			Email:  userClaims.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

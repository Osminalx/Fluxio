package api

import (
	"encoding/json"
	"net/http"

	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
	"github.com/google/uuid"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenHandler godoc
// @Summary Refresh access token
// @Description Generates a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} services.TokenPair
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Invalid or expired refresh token"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/auth/refresh [post]
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding refresh token request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Create refresh token service instance
	refreshTokenService := services.NewRefreshTokenService()
	
	// Validate refresh token and get user
	user, err := refreshTokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		logger.Warn("Failed refresh token attempt: %v", err)
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new token pair using auth service
	tokenPair, err := services.GenerateTokenPair(user)
	if err != nil {
		logger.Error("Error generating new token pair: %v", err)
		http.Error(w, "Error generating tokens", http.StatusInternalServerError)
		return
	}

	// Revoke the old refresh token
	if err := refreshTokenService.RevokeRefreshToken(req.RefreshToken); err != nil {
		logger.Warn("Failed to revoke old refresh token: %v", err)
		// Continue anyway, as the new token was generated successfully
	}

	logger.Info("Token refreshed successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenPair)
}

// LogoutHandler godoc
// @Summary Logout user
// @Description Revokes the refresh token (logout from current device)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token to revoke"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/auth/logout [post]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Error decoding logout request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Create refresh token service instance
	refreshTokenService := services.NewRefreshTokenService()
	
	// Revoke the refresh token
	if err := refreshTokenService.RevokeRefreshToken(req.RefreshToken); err != nil {
		logger.Error("Error revoking refresh token: %v", err)
		http.Error(w, "Error during logout", http.StatusInternalServerError)
		return
	}

	logger.Info("User logged out successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// LogoutAllHandler godoc
// @Summary Logout from all devices
// @Description Revokes all refresh tokens for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security bearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/auth/logout-all [post]
func LogoutAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get userID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse userID to UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid userID format: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Create refresh token service instance
	refreshTokenService := services.NewRefreshTokenService()
	
	// Revoke all refresh tokens for this user
	if err := refreshTokenService.RevokeAllUserRefreshTokens(userID); err != nil {
		logger.Error("Error revoking all user tokens: %v", err)
		http.Error(w, "Error during logout", http.StatusInternalServerError)
		return
	}

	logger.Info("User logged out from all devices: %s", userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out from all devices successfully",
	})
}

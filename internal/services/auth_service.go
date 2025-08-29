package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair represents both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}

func GenerateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // Short-lived access token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(user *models.User) (*TokenPair, error) {
	// Generate access token (short-lived)
	accessToken, err := GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (long-lived)
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := db.DB.Create(refreshTokenModel).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil
}

// generateRefreshToken creates a secure random refresh token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := db.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// RefreshAccessToken validates a refresh token and generates a new access token
func RefreshAccessToken(refreshTokenStr string) (*TokenPair, error) {
	// Find the refresh token in database
	var refreshToken models.RefreshToken
	if err := db.DB.Where("token = ? AND is_revoked = false", refreshTokenStr).First(&refreshToken).Error; err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if token is expired
	if !refreshToken.IsValid() {
		return nil, errors.New("refresh token expired")
	}

	// Get the user
	var user models.User
	if err := db.DB.Where("id = ?", refreshToken.UserID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Revoke old refresh token
	refreshToken.IsRevoked = true
	db.DB.Save(&refreshToken)

	// Generate new token pair
	return GenerateTokenPair(&user)
}

// RevokeRefreshToken revokes a refresh token (for logout)
func RevokeRefreshToken(refreshTokenStr string) error {
	var refreshToken models.RefreshToken
	if err := db.DB.Where("token = ?", refreshTokenStr).First(&refreshToken).Error; err != nil {
		return errors.New("refresh token not found")
	}

	refreshToken.IsRevoked = true
	return db.DB.Save(&refreshToken).Error
}

// RevokeAllUserTokens revokes all refresh tokens for a user (for logout all devices)
func RevokeAllUserTokens(userID string) error {
	return db.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Update("is_revoked", true).Error
}

// CleanupExpiredTokens removes expired refresh tokens from database
func CleanupExpiredTokens() error {
	return db.DB.Where("expires_at < ? OR is_revoked = true", time.Now()).
		Delete(&models.RefreshToken{}).Error
}

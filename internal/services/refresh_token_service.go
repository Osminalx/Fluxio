package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Osminalx/fluxio/internal/db"
	"github.com/Osminalx/fluxio/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenService struct {
	db *gorm.DB
}

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{
		db: db.DB,
	}
}

// TODO: Check if this service uses follows the best practices for security
// generateSecureToken generates a cryptographically secure random token
func (s *RefreshTokenService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateRefreshToken creates a new refresh token for a user
func (s *RefreshTokenService) CreateRefreshToken(userID uuid.UUID, expirationDays int) (*models.RefreshToken, error) {
	if expirationDays <= 0 {
		expirationDays = 30 // Default 30 days
	}

	// Generate secure token
	tokenString, err := s.generateSecureToken()
	if err != nil {
		return nil, err
	}

	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: time.Now().AddDate(0, 0, expirationDays),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsRevoked: false,
	}

	if err := s.db.Create(refreshToken).Error; err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// GetRefreshTokenByToken retrieves a refresh token by its token string
func (s *RefreshTokenService) GetRefreshTokenByToken(tokenString string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := s.db.Preload("User").Where("token = ?", tokenString).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	return &refreshToken, nil
}

// GetRefreshTokenByID retrieves a refresh token by its ID
func (s *RefreshTokenService) GetRefreshTokenByID(tokenID uuid.UUID) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := s.db.Preload("User").Where("id = ?", tokenID).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	return &refreshToken, nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user
func (s *RefreshTokenService) ValidateRefreshToken(tokenString string) (*models.User, error) {
	refreshToken, err := s.GetRefreshTokenByToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if token is valid (not revoked and not expired)
	if !refreshToken.IsValid() {
		return nil, errors.New("refresh token is invalid or expired")
	}

	// Check if user is still active
	if !refreshToken.User.IsAccessible() {
		return nil, errors.New("user account is not accessible")
	}

	return &refreshToken.User, nil
}

// RevokeRefreshToken revokes a refresh token by token string
func (s *RefreshTokenService) RevokeRefreshToken(tokenString string) error {
	updates := map[string]interface{}{
		"is_revoked": true,
		"updated_at": time.Now(),
	}

	result := s.db.Model(&models.RefreshToken{}).Where("token = ?", tokenString).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

// RevokeRefreshTokenByID revokes a refresh token by its ID
func (s *RefreshTokenService) RevokeRefreshTokenByID(tokenID uuid.UUID) error {
	updates := map[string]interface{}{
		"is_revoked": true,
		"updated_at": time.Now(),
	}

	result := s.db.Model(&models.RefreshToken{}).Where("id = ?", tokenID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a specific user
func (s *RefreshTokenService) RevokeAllUserRefreshTokens(userID uuid.UUID) error {
	updates := map[string]interface{}{
		"is_revoked": true,
		"updated_at": time.Now(),
	}

	return s.db.Model(&models.RefreshToken{}).Where("user_id = ?", userID).Updates(updates).Error
}

// GetUserRefreshTokens retrieves all refresh tokens for a user
func (s *RefreshTokenService) GetUserRefreshTokens(userID uuid.UUID, includeRevoked bool, limit, offset int) ([]*models.RefreshToken, error) {
	query := s.db.Where("user_id = ?", userID)

	if !includeRevoked {
		query = query.Where("is_revoked = ?", false)
	}

	query = query.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var tokens []*models.RefreshToken
	if err := query.Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

// CleanupExpiredTokens removes expired refresh tokens from the database
func (s *RefreshTokenService) CleanupExpiredTokens() error {
	now := time.Now()
	return s.db.Where("expires_at < ?", now).Delete(&models.RefreshToken{}).Error
}

// CleanupRevokedTokens removes revoked refresh tokens older than specified days
func (s *RefreshTokenService) CleanupRevokedTokens(olderThanDays int) error {
	if olderThanDays <= 0 {
		olderThanDays = 7 // Default 7 days
	}

	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)
	return s.db.Where("is_revoked = ? AND updated_at < ?", true, cutoffDate).Delete(&models.RefreshToken{}).Error
}

// RotateRefreshToken revokes the old token and creates a new one
func (s *RefreshTokenService) RotateRefreshToken(oldTokenString string, expirationDays int) (*models.RefreshToken, error) {
	// Validate the old token first
	user, err := s.ValidateRefreshToken(oldTokenString)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Revoke the old token
	updates := map[string]interface{}{
		"is_revoked": true,
		"updated_at": time.Now(),
	}

	if err := tx.Model(&models.RefreshToken{}).Where("token = ?", oldTokenString).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create new token
	if expirationDays <= 0 {
		expirationDays = 30
	}

	tokenString, err := s.generateSecureToken()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	newRefreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().AddDate(0, 0, expirationDays),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsRevoked: false,
	}

	if err := tx.Create(newRefreshToken).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return newRefreshToken, nil
}

// GetRefreshTokenStats returns statistics about refresh tokens
func (s *RefreshTokenService) GetRefreshTokenStats(userID *uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	query := s.db.Model(&models.RefreshToken{})

	// Filter by user if specified
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// Total tokens
	var totalCount int64
	query.Count(&totalCount)
	stats["total_tokens"] = totalCount

	// Active tokens (not revoked and not expired)
	var activeCount int64
	now := time.Now()
	query.Where("is_revoked = ? AND expires_at > ?", false, now).Count(&activeCount)
	stats["active_tokens"] = activeCount

	// Revoked tokens
	var revokedCount int64
	s.db.Model(&models.RefreshToken{}).Where("is_revoked = ?", true).Count(&revokedCount)
	if userID != nil {
		s.db.Model(&models.RefreshToken{}).Where("user_id = ? AND is_revoked = ?", *userID, true).Count(&revokedCount)
	}
	stats["revoked_tokens"] = revokedCount

	// Expired tokens
	var expiredCount int64
	expiredQuery := s.db.Model(&models.RefreshToken{}).Where("expires_at <= ?", now)
	if userID != nil {
		expiredQuery = expiredQuery.Where("user_id = ?", *userID)
	}
	expiredQuery.Count(&expiredCount)
	stats["expired_tokens"] = expiredCount

	// Tokens expiring soon (within 7 days)
	var expiringSoonCount int64
	soonDate := now.AddDate(0, 0, 7)
	expiringSoonQuery := s.db.Model(&models.RefreshToken{}).Where("is_revoked = ? AND expires_at > ? AND expires_at <= ?", false, now, soonDate)
	if userID != nil {
		expiringSoonQuery = expiringSoonQuery.Where("user_id = ?", *userID)
	}
	expiringSoonQuery.Count(&expiringSoonCount)
	stats["expiring_soon_tokens"] = expiringSoonCount

	return stats, nil
}

// ExtendRefreshToken extends the expiration date of a refresh token
func (s *RefreshTokenService) ExtendRefreshToken(tokenString string, additionalDays int) (*models.RefreshToken, error) {
	if additionalDays <= 0 {
		return nil, errors.New("additional days must be greater than zero")
	}

	refreshToken, err := s.GetRefreshTokenByToken(tokenString)
	if err != nil {
		return nil, err
	}

	if refreshToken.IsRevoked {
		return nil, errors.New("cannot extend revoked token")
	}

	newExpirationDate := refreshToken.ExpiresAt.AddDate(0, 0, additionalDays)
	updates := map[string]interface{}{
		"expires_at": newExpirationDate,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(refreshToken).Updates(updates).Error; err != nil {
		return nil, err
	}

	refreshToken.ExpiresAt = newExpirationDate
	refreshToken.UpdatedAt = time.Now()

	return refreshToken, nil
}

// GetTokensExpiringWithin returns tokens expiring within the specified number of days
func (s *RefreshTokenService) GetTokensExpiringWithin(days int, userID *uuid.UUID) ([]*models.RefreshToken, error) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	query := s.db.Preload("User").
		Where("is_revoked = ? AND expires_at > ? AND expires_at <= ?", false, now, futureDate).
		Order("expires_at ASC")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	var tokens []*models.RefreshToken
	if err := query.Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

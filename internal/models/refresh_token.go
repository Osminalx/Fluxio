package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Token     string    `json:"token" gorm:"type:varchar(512);not null;unique;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	IsRevoked bool      `json:"is_revoked" gorm:"default:false"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) (err error) {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return
}

// IsValid checks if the refresh token is still valid
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && time.Now().Before(rt.ExpiresAt)
}

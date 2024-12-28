package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey"`
	Username     string         `gorm:"uniqueIndex;not null"`
	Email        string         `gorm:"uniqueIndex;not null"`
	PasswordHash string         `gorm:"not null"`
	Role         string         `gorm:"not null;default:'user'"`
	Balance      Balance        `gorm:"foreignKey:UserID"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type Transaction struct {
	ID         uint           `gorm:"primaryKey"`
	FromUserID uint           `gorm:"index;not null"`
	FromUser   User           `gorm:"foreignKey:FromUserID"`
	ToUserID   uint           `gorm:"index;not null"`
	ToUser     User           `gorm:"foreignKey:ToUserID"`
	Amount     float64        `gorm:"not null"`
	Type       string         `gorm:"not null"` // e.g., "transfer", "deposit", "withdrawal"
	Status     string         `gorm:"not null"` // e.g., "pending", "completed", "failed"
	CreatedAt  time.Time      `gorm:"not null"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type Balance struct {
	UserID        uint           `gorm:"primaryKey"`
	Amount        float64        `gorm:"not null;default:0"`
	LastUpdatedAt time.Time      `gorm:"not null"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type AuditLog struct {
	ID         uint           `gorm:"primaryKey"`
	EntityType string         `gorm:"index;not null"` // e.g., "user", "transaction", "balance"
	EntityID   uint           `gorm:"index;not null"`
	Action     string         `gorm:"not null"` // e.g., "create", "update", "delete"
	Details    string         `gorm:"type:text"`
	CreatedAt  time.Time      `gorm:"not null"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

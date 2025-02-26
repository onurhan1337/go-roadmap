package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BalanceHistory struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	UserID    uint            `gorm:"index;not null" json:"user_id"`
	OldAmount decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"old_amount"`
	NewAmount decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"new_amount"`
	CreatedAt time.Time       `gorm:"not null" json:"created_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (h *BalanceHistory) TableName() string {
	return "balance_history"
}

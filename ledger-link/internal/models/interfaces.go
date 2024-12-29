package models

import (
	"context"
	"sync"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id uint) (*Transaction, error)
	GetByUserID(ctx context.Context, userID uint) ([]Transaction, error)
	Update(ctx context.Context, tx *Transaction) error
}

type BalanceRepository interface {
	Create(ctx context.Context, balance *Balance) error
	GetByUserID(ctx context.Context, userID uint) (*Balance, error)
	Update(ctx context.Context, balance *Balance) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByEntityID(ctx context.Context, entityType string, entityID uint) ([]AuditLog, error)
}

type UserService interface {
	Register(ctx context.Context, user *User) error
	Authenticate(ctx context.Context, email, password string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

type TransactionService interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	ProcessTransaction(ctx context.Context, tx *Transaction) error
	GetUserTransactions(ctx context.Context, userID uint) ([]Transaction, error)
	SubmitTransaction(ctx context.Context, tx *Transaction) error
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID uint) (*Balance, error)
	UpdateBalance(ctx context.Context, userID uint, amount float64) error
	LockBalance(ctx context.Context, userID uint) (*sync.Mutex, error)
}

type AuditService interface {
	LogAction(ctx context.Context, entityType string, entityID uint, action string, details string) error
	GetEntityAuditLog(ctx context.Context, entityType string, entityID uint) ([]AuditLog, error)
}

type TransactionProcessor interface {
	Start(ctx context.Context) error
	Stop()
	Submit(tx *Transaction) error
}

package models

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetUsers(ctx context.Context) ([]*User, error)
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
	GetBalanceHistory(ctx context.Context, userID uint, limit int) ([]BalanceHistory, error)
	CreateBalanceHistory(ctx context.Context, history *BalanceHistory) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByEntityID(ctx context.Context, entityType string, entityID uint) ([]AuditLog, error)
}

type UserService interface {
	Register(ctx context.Context, user *User) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetUsers(ctx context.Context) ([]*User, error)
	Delete(ctx context.Context, id uint) error
	IsAdmin(user *User) bool
	CanAccessUser(requestingUser *User, targetUserID uint) bool
}

type TransactionService interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	ProcessTransaction(ctx context.Context, tx *Transaction) error
	GetUserTransactions(ctx context.Context, userID uint) ([]Transaction, error)
	GetTransaction(ctx context.Context, transactionID uint) (*Transaction, error)
	SubmitTransaction(ctx context.Context, tx *Transaction) error
	Credit(ctx context.Context, userID uint, amount decimal.Decimal, notes string) error
	Debit(ctx context.Context, userID uint, amount decimal.Decimal, notes string) error
	Transfer(ctx context.Context, fromUserID, toUserID uint, amount decimal.Decimal, notes string) error
	Start(ctx context.Context) error
	Stop()
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID uint) (*Balance, error)
	UpdateBalance(ctx context.Context, userID uint, amount decimal.Decimal) error
	LockBalance(ctx context.Context, userID uint) (*sync.Mutex, error)
	GetBalanceHistory(ctx context.Context, userID uint, limit int) ([]BalanceHistory, error)
	GetBalanceAtTime(ctx context.Context, userID uint, timestamp time.Time) (*Balance, error)
	CreateInitialBalance(ctx context.Context, balance *Balance) error
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

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password, username string) (string, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
	RefreshToken(ctx context.Context, oldToken string) (string, error)
}

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Service() AuthService
}

type UserHandler interface {
	GetUsers(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
}

type TransactionHandler interface {
	HandleCredit(w http.ResponseWriter, r *http.Request)
	HandleDebit(w http.ResponseWriter, r *http.Request)
	HandleTransfer(w http.ResponseWriter, r *http.Request)
	HandleGetTransaction(w http.ResponseWriter, r *http.Request)
	HandleGetTransactionHistory(w http.ResponseWriter, r *http.Request)
}

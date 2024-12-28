package models

import (
	"encoding/json"
	"errors"
	"net/mail"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("username must be between 3 and 30 characters")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
	ErrInvalidStatus   = errors.New("invalid transaction status")
	ErrInvalidType     = errors.New("invalid transaction type")
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"` // "-" excludes from JSON
	Role         string         `gorm:"not null;default:'user'" json:"role"`
	Balance      Balance        `gorm:"foreignKey:UserID" json:"balance"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) Validate() error {
	if len(u.Username) < 3 || len(u.Username) > 30 {
		return ErrInvalidUsername
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return ErrInvalidEmail
	}
	if len(u.PasswordHash) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User // Use type alias to avoid recursion
	return json.Marshal(&struct {
		*Alias
		PasswordHash string `json:"-"` // Explicitly exclude password
	}{
		Alias: (*Alias)(u),
	})
}

type TransactionStatus string
type TransactionType string

const (
	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusFailed    TransactionStatus = "failed"

	TypeTransfer   TransactionType = "transfer"
	TypeDeposit    TransactionType = "deposit"
	TypeWithdrawal TransactionType = "withdrawal"
)

type Transaction struct {
	ID         uint              `gorm:"primaryKey" json:"id"`
	FromUserID uint              `gorm:"index;not null" json:"from_user_id"`
	FromUser   User              `gorm:"foreignKey:FromUserID" json:"from_user"`
	ToUserID   uint              `gorm:"index;not null" json:"to_user_id"`
	ToUser     User              `gorm:"foreignKey:ToUserID" json:"to_user"`
	Amount     float64           `gorm:"not null" json:"amount"`
	Type       TransactionType   `gorm:"not null" json:"type"`
	Status     TransactionStatus `gorm:"not null" json:"status"`
	CreatedAt  time.Time         `gorm:"not null" json:"created_at"`
	DeletedAt  gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return ErrInvalidAmount
	}

	validStatus := map[TransactionStatus]bool{
		StatusPending:   true,
		StatusCompleted: true,
		StatusFailed:    true,
	}
	if !validStatus[t.Status] {
		return ErrInvalidStatus
	}

	validType := map[TransactionType]bool{
		TypeTransfer:   true,
		TypeDeposit:    true,
		TypeWithdrawal: true,
	}
	if !validType[t.Type] {
		return ErrInvalidType
	}

	return nil
}

func (t *Transaction) UpdateStatus(status TransactionStatus) error {
	if !t.IsValidStatus(status) {
		return ErrInvalidStatus
	}
	t.Status = status
	return nil
}

func (t *Transaction) IsValidStatus(status TransactionStatus) bool {
	validStatus := map[TransactionStatus]bool{
		StatusPending:   true,
		StatusCompleted: true,
		StatusFailed:    true,
	}
	return validStatus[status]
}

type Balance struct {
	UserID        uint           `gorm:"primaryKey" json:"user_id"`
	amount        int64          `gorm:"-" json:"-"`                       // Internal atomic counter
	Amount        float64        `gorm:"not null;default:0" json:"amount"` // For GORM and JSON
	LastUpdatedAt time.Time      `gorm:"not null" json:"last_updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	mu            sync.RWMutex   `gorm:"-" json:"-"` // For complex operations
}

func (b *Balance) AfterFind(tx *gorm.DB) error {
	atomic.StoreInt64(&b.amount, int64(b.Amount*1e8)) // Store as fixed-point number
	return nil
}

func (b *Balance) BeforeSave(tx *gorm.DB) error {
	b.Amount = float64(atomic.LoadInt64(&b.amount)) / 1e8
	return nil
}

func (b *Balance) SafeAmount() float64 {
	return float64(atomic.LoadInt64(&b.amount)) / 1e8
}

func (b *Balance) UpdateAmount(amount float64) {
	atomic.StoreInt64(&b.amount, int64(amount*1e8))
	b.mu.Lock()
	b.LastUpdatedAt = time.Now()
	b.mu.Unlock()
}

func (b *Balance) AddAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	amountInt := int64(amount * 1e8)
	for {
		current := atomic.LoadInt64(&b.amount)
		if atomic.CompareAndSwapInt64(&b.amount, current, current+amountInt) {
			b.mu.Lock()
			b.LastUpdatedAt = time.Now()
			b.mu.Unlock()
			return nil
		}
	}
}

func (b *Balance) SubtractAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	amountInt := int64(amount * 1e8)
	for {
		current := atomic.LoadInt64(&b.amount)
		if current < amountInt {
			return errors.New("insufficient balance")
		}
		if atomic.CompareAndSwapInt64(&b.amount, current, current-amountInt) {
			b.mu.Lock()
			b.LastUpdatedAt = time.Now()
			b.mu.Unlock()
			return nil
		}
	}
}

func (b *Balance) MarshalJSON() ([]byte, error) {
	type Alias Balance
	return json.Marshal(&struct {
		*Alias
		Amount float64 `json:"amount"`
	}{
		Alias:  (*Alias)(b),
		Amount: b.SafeAmount(),
	})
}

type AuditLog struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	EntityType string         `gorm:"index;not null" json:"entity_type"`
	EntityID   uint           `gorm:"index;not null" json:"entity_id"`
	Action     string         `gorm:"not null" json:"action"`
	Details    string         `gorm:"type:text" json:"details"`
	CreatedAt  time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *AuditLog) Validate() error {
	if a.EntityType == "" {
		return errors.New("entity type is required")
	}
	if a.EntityID == 0 {
		return errors.New("entity ID is required")
	}
	if a.Action == "" {
		return errors.New("action is required")
	}
	return nil
}

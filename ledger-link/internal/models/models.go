package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("username must be between 3 and 30 characters")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
	ErrInvalidStatus   = errors.New("invalid transaction status")
	ErrInvalidType     = errors.New("invalid transaction type")
	ErrInvalidRole     = errors.New("invalid user role")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"

	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusFailed    TransactionStatus = "failed"
	StatusCancelled TransactionStatus = "cancelled"

	TypeTransfer    TransactionType = "transfer"
	TypeDeposit     TransactionType = "deposit"
	TypeWithdrawal  TransactionType = "withdrawal"
	TypeAdjustment  TransactionType = "adjustment"

	EntityTypeUser        = "user"
	EntityTypeTransaction = "transaction"
	EntityTypeBalance     = "balance"

	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(30);uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Role         string         `gorm:"not null;default:'user'" json:"role"`
	Balance      Balance        `gorm:"foreignKey:UserID" json:"balance"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) ValidateUsername() error {
	username := strings.TrimSpace(u.Username)
	if len(username) < 3 || len(username) > 30 {
		return ErrInvalidUsername
	}
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", username)
	if err != nil || !matched {
		return errors.New("username can only contain letters, numbers, underscores, and hyphens")
	}
	return nil
}

func (u *User) ValidateEmail() error {
	email := strings.TrimSpace(u.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func (u *User) ValidateRole() error {
	switch u.Role {
	case RoleUser, RoleAdmin:
		return nil
	default:
		return ErrInvalidRole
	}
}

func (u *User) Validate() error {
	if err := u.ValidateUsername(); err != nil {
		return err
	}
	if err := u.ValidateEmail(); err != nil {
		return err
	}
	if err := u.ValidateRole(); err != nil {
		return err
	}
	if len(u.PasswordHash) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = RoleUser
	}
	return u.Validate()
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	return u.Validate()
}

func (u *User) SafeCopy() *User {
	copy := &User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.Balance.UserID != 0 {
		copy.Balance = Balance{
			UserID:        u.Balance.UserID,
			Amount:        u.Balance.SafeAmount(),
			LastUpdatedAt: u.Balance.LastUpdatedAt,
		}
	}
	return copy
}

func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		*Alias
		PasswordHash string `json:"-"`
	}{
		Alias: (*Alias)(u),
	})
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

type TransactionStatus string
type TransactionType string

type Transaction struct {
	ID         uint              `gorm:"primaryKey" json:"id"`
	FromUserID uint              `gorm:"index;not null" json:"from_user_id"`
	FromUser   User              `gorm:"foreignKey:FromUserID" json:"from_user"`
	ToUserID   uint              `gorm:"index;not null" json:"to_user_id"`
	ToUser     User              `gorm:"foreignKey:ToUserID" json:"to_user"`
	Amount     float64           `gorm:"not null" json:"amount"`
	Type       TransactionType   `gorm:"not null" json:"type"`
	Status     TransactionStatus `gorm:"not null" json:"status"`
	Notes      string            `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt  time.Time         `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time         `gorm:"not null" json:"updated_at"`
	DeletedAt  gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return ErrInvalidAmount
	}

	if !t.IsValidType(t.Type) {
		return ErrInvalidType
	}

	if !t.IsValidStatus(t.Status) {
		return ErrInvalidStatus
	}

	if t.Type == TypeTransfer && (t.FromUserID == 0 || t.ToUserID == 0) {
		return errors.New("transfer requires both from and to users")
	}

	return nil
}

func (t *Transaction) IsValidType(txType TransactionType) bool {
	switch txType {
	case TypeTransfer, TypeDeposit, TypeWithdrawal, TypeAdjustment:
		return true
	default:
		return false
	}
}

func (t *Transaction) IsValidStatus(status TransactionStatus) bool {
	switch status {
	case StatusPending, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

func (t *Transaction) UpdateStatus(status TransactionStatus) error {
	if !t.IsValidStatus(status) {
		return ErrInvalidStatus
	}
	t.Status = status
	return nil
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(t),
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	t.CreatedAt, err = time.Parse(time.RFC3339, aux.CreatedAt)
	if err != nil {
		return err
	}

	t.UpdatedAt, err = time.Parse(time.RFC3339, aux.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

type Balance struct {
	UserID        uint           `gorm:"primaryKey" json:"user_id"`
	amount        int64          `gorm:"-" json:"-"`
	Amount        float64        `gorm:"not null;default:0" json:"amount"`
	LastUpdatedAt time.Time      `gorm:"not null" json:"last_updated_at"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updated_at"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	mu            sync.RWMutex   `gorm:"-" json:"-"`
}

func (b *Balance) Validate() error {
	if b.Amount < 0 {
		return errors.New("balance cannot be negative")
	}
	if b.UserID == 0 {
		return errors.New("user ID is required")
	}
	return nil
}

func (b *Balance) BeforeCreate(tx *gorm.DB) error {
	if err := b.Validate(); err != nil {
		return err
	}
	b.LastUpdatedAt = time.Now()
	return nil
}

func (b *Balance) BeforeUpdate(tx *gorm.DB) error {
	if err := b.Validate(); err != nil {
		return err
	}
	b.LastUpdatedAt = time.Now()
	return nil
}

func (b *Balance) AfterFind(tx *gorm.DB) error {
	atomic.StoreInt64(&b.amount, int64(b.Amount*100))
	return nil
}

func (b *Balance) BeforeSave(tx *gorm.DB) error {
	b.Amount = float64(atomic.LoadInt64(&b.amount)) / 100
	return nil
}

func (b *Balance) SafeAmount() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return float64(atomic.LoadInt64(&b.amount)) / 100
}

func (b *Balance) UpdateAmount(amount float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	atomic.StoreInt64(&b.amount, int64(amount*100))
	b.Amount = amount
	b.LastUpdatedAt = time.Now()
}

func (b *Balance) AddAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	currentAmount := float64(atomic.LoadInt64(&b.amount)) / 100
	newAmount := currentAmount + amount
	atomic.StoreInt64(&b.amount, int64(newAmount*100))
	b.Amount = newAmount
	b.LastUpdatedAt = time.Now()

	return nil
}

func (b *Balance) SubtractAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	currentAmount := float64(atomic.LoadInt64(&b.amount)) / 100
	if currentAmount < amount {
		return errors.New("insufficient balance")
	}

	newAmount := currentAmount - amount
	atomic.StoreInt64(&b.amount, int64(newAmount*100))
	b.Amount = newAmount
	b.LastUpdatedAt = time.Now()

	return nil
}

func (b *Balance) MarshalJSON() ([]byte, error) {
	type Alias Balance
	return json.Marshal(&struct {
		*Alias
		Amount        string `json:"amount"`
		LastUpdatedAt string `json:"last_updated_at"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}{
		Alias:         (*Alias)(b),
		Amount:        fmt.Sprintf("%.2f", b.SafeAmount()),
		LastUpdatedAt: b.LastUpdatedAt.Format(time.RFC3339),
		CreatedAt:     b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     b.UpdatedAt.Format(time.RFC3339),
	})
}

func (b *Balance) UnmarshalJSON(data []byte) error {
	type Alias Balance
	aux := &struct {
		*Alias
		Amount        string `json:"amount"`
		LastUpdatedAt string `json:"last_updated_at"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(aux.Amount, 64)
	if err != nil {
		return err
	}
	b.UpdateAmount(amount)

	b.LastUpdatedAt, err = time.Parse(time.RFC3339, aux.LastUpdatedAt)
	if err != nil {
		return err
	}

	b.CreatedAt, err = time.Parse(time.RFC3339, aux.CreatedAt)
	if err != nil {
		return err
	}

	b.UpdatedAt, err = time.Parse(time.RFC3339, aux.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

type AuditLog struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	EntityType string         `gorm:"index;not null" json:"entity_type"`
	EntityID   uint          `gorm:"index;not null" json:"entity_id"`
	Action     string         `gorm:"not null" json:"action"`
	Details    string         `gorm:"type:text" json:"details"`
	UserID     uint          `gorm:"index;not null" json:"user_id"`
	User       User           `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt  time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *AuditLog) Validate() error {
	if a.EntityID == 0 {
		return errors.New("entity ID is required")
	}
	if a.EntityType == "" {
		return errors.New("entity type is required")
	}
	if a.Action == "" {
		return errors.New("action is required")
	}
	if a.UserID == 0 {
		return errors.New("user ID is required")
	}

	switch a.EntityType {
	case EntityTypeUser, EntityTypeTransaction, EntityTypeBalance:
		// valid entity type
	default:
		return errors.New("invalid entity type")
	}

	switch a.Action {
	case ActionCreate, ActionUpdate, ActionDelete:
		// valid action
	default:
		return errors.New("invalid action")
	}

	return nil
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	return a.Validate()
}

func (a *AuditLog) BeforeUpdate(tx *gorm.DB) error {
	return a.Validate()
}

func (a *AuditLog) MarshalJSON() ([]byte, error) {
	type Alias AuditLog
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(a),
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	})
}

func (a *AuditLog) UnmarshalJSON(data []byte) error {
	type Alias AuditLog
	aux := &struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	a.CreatedAt, err = time.Parse(time.RFC3339, aux.CreatedAt)
	if err != nil {
		return err
	}

	a.UpdatedAt, err = time.Parse(time.RFC3339, aux.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

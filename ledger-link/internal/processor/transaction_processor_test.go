package processor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"
)

// Mock implementations
type MockTransactionRepo struct {
	mock.Mock
}

func (m *MockTransactionRepo) Create(ctx context.Context, tx *models.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockTransactionRepo) Update(ctx context.Context, tx *models.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockTransactionRepo) GetByID(ctx context.Context, id uint) (*models.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetByUserID(ctx context.Context, userID uint) ([]models.Transaction, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) GetBalance(ctx context.Context, userID uint) (*models.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Balance), args.Error(1)
}

func (m *MockBalanceService) UpdateBalance(ctx context.Context, userID uint, amount decimal.Decimal) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockBalanceService) LockBalance(ctx context.Context, userID uint) (*sync.Mutex, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*sync.Mutex), args.Error(1)
}

func (m *MockBalanceService) GetBalanceHistory(ctx context.Context, userID uint, limit int) ([]models.BalanceHistory, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]models.BalanceHistory), args.Error(1)
}

func (m *MockBalanceService) GetBalanceAtTime(ctx context.Context, userID uint, timestamp time.Time) (*models.Balance, error) {
	args := m.Called(ctx, userID, timestamp)
	return args.Get(0).(*models.Balance), args.Error(1)
}

func (m *MockBalanceService) CreateInitialBalance(ctx context.Context, balance *models.Balance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAction(ctx context.Context, entityType string, entityID uint, action string, details string) error {
	args := m.Called(ctx, entityType, entityID, action, details)
	return args.Error(0)
}

func (m *MockAuditService) GetEntityAuditLog(ctx context.Context, entityType string, entityID uint) ([]models.AuditLog, error) {
	args := m.Called(ctx, entityType, entityID)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func TestBatchProcessing(t *testing.T) {
	// Create mocks
	repo := new(MockTransactionRepo)
	balanceSvc := new(MockBalanceService)
	auditSvc := new(MockAuditService)
	logger := logger.New("info")

	// Create processor with test configuration
	processor := NewTransactionProcessor(repo, balanceSvc, auditSvc, logger)
	processor.batchConfig = BatchConfig{
		MaxBatchSize:    5,
		BatchTimeout:    100 * time.Millisecond,
		WorkerCount:     2,
		QueueBufferSize: 10,
	}

	// Start the processor
	ctx := context.Background()
	err := processor.Start(ctx)
	assert.NoError(t, err)
	defer processor.Stop()

	// Test cases
	t.Run("Process multiple deposits for same user", func(t *testing.T) {
		userID := uint(1)
		initialBalance := decimal.NewFromInt(1000)

		// Setup mock expectations
		balance := &models.Balance{UserID: userID, Amount: initialBalance}
		balanceSvc.On("GetBalance", mock.Anything, userID).Return(balance, nil).Times(2) // Expect multiple calls

		// Calculate expected final balance
		expectedTotal := initialBalance
		transactions := make([]*models.Transaction, 3)

		// Track running balance for each batch
		runningBalance := initialBalance

		for i := 0; i < 3; i++ {
			amount := decimal.NewFromInt(int64(100 * (i + 1)))
			expectedTotal = expectedTotal.Add(amount)

			tx := &models.Transaction{
				ID:       uint(i + 1),
				ToUserID: userID,
				Amount:   amount,
				Type:     models.TypeDeposit,
				Status:   models.StatusPending,
			}
			transactions[i] = tx

			// Setup update expectations
			repo.On("Update", mock.Anything, mock.MatchedBy(func(t *models.Transaction) bool {
				return t.ID == tx.ID && t.Status == models.StatusCompleted
			})).Return(nil)

			// For each transaction, expect a potential balance update
			runningBalance = runningBalance.Add(amount)
			balanceSvc.On("UpdateBalance", mock.Anything, userID, mock.MatchedBy(func(amount decimal.Decimal) bool {
				return amount.GreaterThanOrEqual(initialBalance) && amount.LessThanOrEqual(expectedTotal)
			})).Return(nil).Maybe()
		}

		// Setup audit log expectations
		for _, tx := range transactions {
			auditSvc.On("LogAction", mock.Anything, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, mock.Anything).Return(nil)
		}

		// Submit transactions
		var wg sync.WaitGroup
		for _, tx := range transactions {
			wg.Add(1)
			go func(tx *models.Transaction) {
				defer wg.Done()
				err := processor.SubmitForBatchProcessing(tx)
				assert.NoError(t, err)
			}(tx)
		}

		// Wait for all submissions
		wg.Wait()

		// Wait for batch processing
		time.Sleep(200 * time.Millisecond)

		// Verify expectations
		mock.AssertExpectationsForObjects(t, repo, balanceSvc, auditSvc)
	})

	t.Run("Process deposits for different users", func(t *testing.T) {
		users := []uint{1, 2}
		initialBalance := decimal.NewFromInt(1000)

		// Setup expectations for each user
		for _, userID := range users {
			balance := &models.Balance{UserID: userID, Amount: initialBalance}
			balanceSvc.On("GetBalance", mock.Anything, userID).Return(balance, nil)

			amount := decimal.NewFromInt(500)
			expectedTotal := initialBalance.Add(amount)

			tx := &models.Transaction{
				ID:       uint(userID + 10), // Unique IDs
				ToUserID: userID,
				Amount:   amount,
				Type:     models.TypeDeposit,
				Status:   models.StatusPending,
			}

			// Expect balance update
			balanceSvc.On("UpdateBalance", mock.Anything, userID, expectedTotal).Return(nil)

			// Expect transaction update
			repo.On("Update", mock.Anything, mock.MatchedBy(func(t *models.Transaction) bool {
				return t.ID == tx.ID && t.Status == models.StatusCompleted
			})).Return(nil)

			// Expect audit log
			auditSvc.On("LogAction", mock.Anything, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, mock.Anything).Return(nil)

			// Submit transaction
			err := processor.SubmitForBatchProcessing(tx)
			assert.NoError(t, err)
		}

		// Wait for batch processing
		time.Sleep(200 * time.Millisecond)

		// Verify expectations
		mock.AssertExpectationsForObjects(t, repo, balanceSvc, auditSvc)
	})

	t.Run("Handle batch timeout", func(t *testing.T) {
		userID := uint(3)
		initialBalance := decimal.NewFromInt(1000)
		balance := &models.Balance{UserID: userID, Amount: initialBalance}

		// Setup mock expectations
		balanceSvc.On("GetBalance", mock.Anything, userID).Return(balance, nil)

		amount := decimal.NewFromInt(100)
		expectedTotal := initialBalance.Add(amount)

		tx := &models.Transaction{
			ID:       uint(20),
			ToUserID: userID,
			Amount:   amount,
			Type:     models.TypeDeposit,
			Status:   models.StatusPending,
		}

		// Expect balance update
		balanceSvc.On("UpdateBalance", mock.Anything, userID, expectedTotal).Return(nil)

		// Expect transaction update
		repo.On("Update", mock.Anything, mock.MatchedBy(func(t *models.Transaction) bool {
			return t.ID == tx.ID && t.Status == models.StatusCompleted
		})).Return(nil)

		// Expect audit log
		auditSvc.On("LogAction", mock.Anything, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, mock.Anything).Return(nil)

		// Submit single transaction
		err := processor.SubmitForBatchProcessing(tx)
		assert.NoError(t, err)

		// Wait for timeout-based processing
		time.Sleep(150 * time.Millisecond)

		// Verify expectations
		mock.AssertExpectationsForObjects(t, repo, balanceSvc, auditSvc)
	})

	t.Run("Non-deposit transactions are processed immediately", func(t *testing.T) {
		userID := uint(4)
		initialBalance := decimal.NewFromInt(1000)
		balance := &models.Balance{UserID: userID, Amount: initialBalance}

		// Setup mock expectations
		balanceSvc.On("GetBalance", mock.Anything, userID).Return(balance, nil)

		amount := decimal.NewFromInt(100)
		expectedTotal := initialBalance.Sub(amount)

		tx := &models.Transaction{
			ID:         uint(30),
			FromUserID: userID,
			Amount:     amount,
			Type:       models.TypeWithdrawal,
			Status:     models.StatusPending,
		}

		// Expect balance update
		balanceSvc.On("UpdateBalance", mock.Anything, userID, expectedTotal).Return(nil)

		// Expect transaction update
		repo.On("Update", mock.Anything, mock.MatchedBy(func(t *models.Transaction) bool {
			return t.ID == tx.ID && t.Status == models.StatusCompleted
		})).Return(nil)

		// Expect audit log
		auditSvc.On("LogAction", mock.Anything, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, mock.Anything).Return(nil)

		// Submit withdrawal transaction
		err := processor.SubmitForBatchProcessing(tx)
		assert.NoError(t, err)

		// Verify expectations immediately
		mock.AssertExpectationsForObjects(t, repo, balanceSvc, auditSvc)
	})
}

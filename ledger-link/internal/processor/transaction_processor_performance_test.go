package processor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"
)

// setupTestProcessor creates a processor for performance testing
func setupTestProcessor(useBatch bool) (*TransactionProcessor, *MockTransactionRepo, *MockBalanceService, *MockAuditService) {
	repo := new(MockTransactionRepo)
	balanceSvc := new(MockBalanceService)
	auditSvc := new(MockAuditService)
	logger := logger.New("error") // Use error level to reduce noise

	processor := NewTransactionProcessor(repo, balanceSvc, auditSvc, logger)
	if useBatch {
		processor.batchConfig = BatchConfig{
			MaxBatchSize:    100,
			BatchTimeout:    50 * time.Millisecond,
			WorkerCount:     2,
			QueueBufferSize: 1000,
		}
	}

	return processor, repo, balanceSvc, auditSvc
}

// setupTestMocks prepares mock services for testing
func setupTestMocks(repo *MockTransactionRepo, balanceSvc *MockBalanceService, auditSvc *MockAuditService, simulateLatency bool) {
	balance := &models.Balance{
		UserID: 1,
		Amount: decimal.NewFromInt(1000000),
	}

	// Simulate database/network latency if requested
	if simulateLatency {
		balanceSvc.On("GetBalance", mock.Anything, uint(1)).Return(balance, nil).
			Run(func(args mock.Arguments) {
				time.Sleep(5 * time.Millisecond) // Simulate DB read latency
			}).Maybe()

		balanceSvc.On("UpdateBalance", mock.Anything, uint(1), mock.Anything).Return(nil).
			Run(func(args mock.Arguments) {
				time.Sleep(10 * time.Millisecond) // Simulate DB write latency
			}).Maybe()

		repo.On("Update", mock.Anything, mock.Anything).Return(nil).
			Run(func(args mock.Arguments) {
				time.Sleep(5 * time.Millisecond) // Simulate DB write latency
			}).Maybe()

		auditSvc.On("LogAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).
			Run(func(args mock.Arguments) {
				time.Sleep(2 * time.Millisecond) // Simulate audit log write
			}).Maybe()
	} else {
		balanceSvc.On("GetBalance", mock.Anything, uint(1)).Return(balance, nil).Maybe()
		balanceSvc.On("UpdateBalance", mock.Anything, uint(1), mock.Anything).Return(nil).Maybe()
		repo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
		auditSvc.On("LogAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	}
}

// generateTestTransactions creates test transactions
func generateTestTransactions(count int) []*models.Transaction {
	txs := make([]*models.Transaction, count)
	for i := 0; i < count; i++ {
		txs[i] = &models.Transaction{
			ID:       uint(i + 1),
			ToUserID: 1,
			Amount:   decimal.NewFromInt(100),
			Type:     models.TypeDeposit,
			Status:   models.StatusPending,
			Notes:    fmt.Sprintf("Test transaction %d", i+1),
		}
	}
	return txs
}

func TestPerformanceComparison(t *testing.T) {
	fmt.Println("\n🚀 Transaction Processing Performance Test")
	fmt.Println("==========================================")
	fmt.Println("Simulating real-world conditions with database latency:")
	fmt.Println("- Database read: 5ms")
	fmt.Println("- Database write: 10ms")
	fmt.Println("- Transaction update: 5ms")
	fmt.Println("- Audit log: 2ms")
	fmt.Println("==========================================")

	testCases := []struct {
		name     string
		txCount  int
		useBatch bool
	}{
		{"Small Batch (10 transactions)", 10, true},
		{"Without Batch (10 transactions)", 10, false},
		{"Medium Batch (100 transactions)", 100, true},
		{"Without Batch (100 transactions)", 100, false},
		{"Large Batch (500 transactions)", 500, true},
		{"Without Batch (500 transactions)", 500, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			processor, repo, balanceSvc, auditSvc := setupTestProcessor(tc.useBatch)
			setupTestMocks(repo, balanceSvc, auditSvc, true) // Enable latency simulation
			transactions := generateTestTransactions(tc.txCount)

			if tc.useBatch {
				err := processor.Start(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer processor.Stop()
			}

			// Run multiple iterations to get average
			iterations := 3
			var totalDuration time.Duration

			for i := 0; i < iterations; i++ {
				start := time.Now()
				completed := make(chan struct{})

				go func() {
					if tc.useBatch {
						for _, tx := range transactions {
							_ = processor.SubmitForBatchProcessing(tx)
						}
					} else {
						for _, tx := range transactions {
							_ = processor.ProcessTransaction(ctx, tx)
						}
					}
					completed <- struct{}{}
				}()

				// Wait for processing to complete
				<-completed
				if tc.useBatch {
					// Wait for the last batch to be processed
					time.Sleep(100 * time.Millisecond)
				}

				totalDuration += time.Since(start)
			}

			// Calculate metrics
			avgDuration := totalDuration / time.Duration(iterations)
			txPerSecond := float64(tc.txCount) / avgDuration.Seconds()
			avgPerTx := avgDuration / time.Duration(tc.txCount)

			// Calculate theoretical DB operations
			dbOpsWithoutBatch := tc.txCount * 4 // Get balance, update balance, update tx, audit log
			dbOpsWithBatch := tc.txCount + 3    // One balance update per batch + tx updates + audit logs

			dbOps := dbOpsWithoutBatch
			if tc.useBatch && tc.txCount > 1 {
				dbOps = dbOpsWithBatch
			}

			// Print results in a clear format
			fmt.Printf("\n📊 %s\n", tc.name)
			fmt.Printf("   ├─ Transactions: %d\n", tc.txCount)
			fmt.Printf("   ├─ Total Time: %v\n", avgDuration.Round(time.Millisecond))
			fmt.Printf("   ├─ Speed: %d transactions/second\n", int(txPerSecond))
			fmt.Printf("   ├─ Average per Transaction: %v\n", avgPerTx.Round(time.Microsecond))
			fmt.Printf("   ├─ Database Operations: %d\n", dbOps)
			fmt.Printf("   └─ Batch Processing: %v\n", tc.useBatch)
		})
	}

	fmt.Println("\n✨ Performance Test Complete")
	fmt.Println("==========================================")
}

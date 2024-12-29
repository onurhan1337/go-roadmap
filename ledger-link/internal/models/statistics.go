package models

import (
	"sync/atomic"
	"time"
)

type TransactionStats struct {
	TotalTransactions     atomic.Uint64
	SuccessfulTransactions atomic.Uint64
	FailedTransactions    atomic.Uint64
	TotalAmount          atomic.Uint64
	LastUpdateTime       atomic.Int64
}

func NewTransactionStats() *TransactionStats {
	return &TransactionStats{}
}

func (s *TransactionStats) IncrementTotal() uint64 {
	s.updateTimestamp()
	return s.TotalTransactions.Add(1)
}

func (s *TransactionStats) IncrementSuccessful() uint64 {
	s.updateTimestamp()
	return s.SuccessfulTransactions.Add(1)
}

func (s *TransactionStats) IncrementFailed() uint64 {
	s.updateTimestamp()
	return s.FailedTransactions.Add(1)
}

func (s *TransactionStats) AddAmount(amount float64) {
	amountCents := uint64(amount * 100)
	s.TotalAmount.Add(amountCents)
	s.updateTimestamp()
}

func (s *TransactionStats) GetTotalAmount() float64 {
	return float64(s.TotalAmount.Load()) / 100
}

func (s *TransactionStats) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_transactions":      s.TotalTransactions.Load(),
		"successful_transactions": s.SuccessfulTransactions.Load(),
		"failed_transactions":     s.FailedTransactions.Load(),
		"total_amount":           s.GetTotalAmount(),
		"last_update":            time.Unix(s.LastUpdateTime.Load(), 0),
	}
}

func (s *TransactionStats) updateTimestamp() {
	s.LastUpdateTime.Store(time.Now().Unix())
}

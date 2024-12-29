package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ledger-link/pkg/logger"
)

// Task represents a unit of work to be processed
type Task interface {
	Process(ctx context.Context) error
	ID() string
}

// ProcessorStats tracks the processor's performance metrics
type ProcessorStats struct {
	mu            sync.RWMutex
	totalTasks    int64
	successTasks  int64
	failedTasks   int64
	processingTime time.Duration
}

// BatchProcessor is a concurrent task processor for batch operations
type BatchProcessor struct {
	workerCount int
	jobQueue    chan Task
	logger      *logger.Logger
	wg          sync.WaitGroup
	done        chan struct{}
	stats       *ProcessorStats
}

// Config holds the configuration for the BatchProcessor
type Config struct {
	WorkerCount int
	QueueSize   int
	Logger      *logger.Logger
}

// NewBatchProcessor creates a new BatchProcessor instance
func NewBatchProcessor(cfg Config) *BatchProcessor {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 5
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}

	return &BatchProcessor{
		workerCount: cfg.WorkerCount,
		jobQueue:    make(chan Task, cfg.QueueSize),
		logger:      cfg.Logger,
		done:        make(chan struct{}),
		stats:       &ProcessorStats{},
	}
}

// Start initializes the worker pool and begins processing tasks
func (p *BatchProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting batch processor", "workers", p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	go func() {
		<-ctx.Done()
		p.logger.Info("shutting down batch processor")
		close(p.jobQueue)
		close(p.done)
	}()

	return nil
}

// Stop waits for all workers to finish and stops the processor
func (p *BatchProcessor) Stop() {
	p.wg.Wait()
	p.logger.Info("batch processor stopped")
}

// Submit adds a new task to the processing queue
func (p *BatchProcessor) Submit(task Task) error {
	select {
	case p.jobQueue <- task:
		return nil
	case <-p.done:
		return fmt.Errorf("batch processor is shutting down")
	default:
		return fmt.Errorf("task queue is full")
	}
}

// worker processes tasks from the job queue
func (p *BatchProcessor) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	p.logger.Info("starting worker", "worker_id", id)

	for {
		select {
		case task, ok := <-p.jobQueue:
			if !ok {
				p.logger.Info("worker shutting down", "worker_id", id)
				return
			}

			start := time.Now()
			if err := task.Process(ctx); err != nil {
				p.logger.Error("failed to process task",
					"error", err,
					"task_id", task.ID(),
					"worker_id", id,
				)
				p.incrementFailedTasks()
			} else {
				p.incrementSuccessfulTasks()
			}
			p.addProcessingTime(time.Since(start))

		case <-ctx.Done():
			p.logger.Info("worker context cancelled", "worker_id", id)
			return
		}
	}
}

// GetStats returns the current processing statistics
func (p *BatchProcessor) GetStats() map[string]interface{} {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	return map[string]interface{}{
		"total_tasks":     p.stats.totalTasks,
		"successful_tasks": p.stats.successTasks,
		"failed_tasks":    p.stats.failedTasks,
		"processing_time": p.stats.processingTime.String(),
		"avg_time_per_task": time.Duration(int64(p.stats.processingTime) / p.stats.totalTasks).String(),
	}
}

func (p *BatchProcessor) incrementSuccessfulTasks() {
	p.stats.mu.Lock()
	p.stats.successTasks++
	p.stats.totalTasks++
	p.stats.mu.Unlock()
}

func (p *BatchProcessor) incrementFailedTasks() {
	p.stats.mu.Lock()
	p.stats.failedTasks++
	p.stats.totalTasks++
	p.stats.mu.Unlock()
}

func (p *BatchProcessor) addProcessingTime(duration time.Duration) {
	p.stats.mu.Lock()
	p.stats.processingTime += duration
	p.stats.mu.Unlock()
}
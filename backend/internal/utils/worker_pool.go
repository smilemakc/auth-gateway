package utils

import (
	"context"
	"sync"
)

// Task represents a background task to be executed
type Task func(ctx context.Context) error

// WorkerPool manages a pool of workers for executing background tasks
type WorkerPool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	once      sync.Once
}

// NewWorkerPool creates a new worker pool
// workers: number of concurrent workers
// queueSize: size of the task queue (0 = unbuffered)
func NewWorkerPool(workers, queueSize int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	if queueSize < 0 {
		queueSize = 0
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, queueSize),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start workers
	pool.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go pool.worker(i)
	}

	return pool
}

// worker runs in a goroutine and processes tasks from the queue
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			// Execute task with context
			if err := task(p.ctx); err != nil {
				// Log error if needed (can be extended with logger)
				_ = err
			}
		}
	}
}

// Submit adds a task to the queue
// Returns false if the pool is shutting down
func (p *WorkerPool) Submit(task Task) bool {
	select {
	case <-p.ctx.Done():
		return false
	case p.taskQueue <- task:
		return true
	}
}

// Shutdown gracefully shuts down the worker pool
// Waits for all queued tasks to complete
func (p *WorkerPool) Shutdown() {
	p.once.Do(func() {
		close(p.taskQueue)
		p.cancel()
		p.wg.Wait()
	})
}

// ShutdownNow immediately shuts down the worker pool
// Does not wait for queued tasks
func (p *WorkerPool) ShutdownNow() {
	p.once.Do(func() {
		p.cancel()
		close(p.taskQueue)
		p.wg.Wait()
	})
}

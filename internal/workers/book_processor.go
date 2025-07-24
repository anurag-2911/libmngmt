package workers

import (
	"context"
	"fmt"
	"libmngmt/internal/models"
	"log"
	"sync"
	"time"
)

// BookProcessor handles concurrent book operations
type BookProcessor struct {
	workers    int
	jobQueue   chan BookJob
	resultChan chan BookResult
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// BookJob represents a job to be processed
type BookJob struct {
	ID         string
	Type       JobType
	BookData   *models.CreateBookRequest
	UpdateData *models.UpdateBookRequest
	Callback   func(BookResult)
}

// JobType defines the type of operation
type JobType int

const (
	JobTypeValidate JobType = iota
	JobTypeProcess
	JobTypeNotify
)

// BookResult represents the result of a job
type BookResult struct {
	JobID   string
	Success bool
	Error   error
	Book    *models.Book
	Message string
}

// NewBookProcessor creates a new book processor with specified number of workers
func NewBookProcessor(workers int, bufferSize int) *BookProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &BookProcessor{
		workers:    workers,
		jobQueue:   make(chan BookJob, bufferSize),
		resultChan: make(chan BookResult, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins processing jobs with the specified number of workers
func (bp *BookProcessor) Start() {
	log.Printf("Starting BookProcessor with %d workers", bp.workers)

	// Start worker goroutines
	for i := 0; i < bp.workers; i++ {
		bp.wg.Add(1)
		go bp.worker(i)
	}

	// Start result processor
	bp.wg.Add(1)
	go bp.resultProcessor()
}

// Stop gracefully shuts down the processor
func (bp *BookProcessor) Stop() {
	log.Println("Stopping BookProcessor...")
	bp.cancel()
	close(bp.jobQueue)
	bp.wg.Wait()
	close(bp.resultChan)
	log.Println("BookProcessor stopped")
}

// SubmitJob submits a job for processing
func (bp *BookProcessor) SubmitJob(job BookJob) error {
	select {
	case bp.jobQueue <- job:
		return nil
	case <-bp.ctx.Done():
		return fmt.Errorf("processor is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// worker processes jobs from the job queue
func (bp *BookProcessor) worker(id int) {
	defer bp.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case job, ok := <-bp.jobQueue:
			if !ok {
				log.Printf("Worker %d: job queue closed, exiting", id)
				return
			}

			log.Printf("Worker %d processing job %s of type %d", id, job.ID, job.Type)
			result := bp.processJob(job)

			// Send result to result channel
			select {
			case bp.resultChan <- result:
			case <-bp.ctx.Done():
				return
			}

		case <-bp.ctx.Done():
			log.Printf("Worker %d: context cancelled, exiting", id)
			return
		}
	}
}

// processJob processes a single job
func (bp *BookProcessor) processJob(job BookJob) BookResult {
	result := BookResult{
		JobID:   job.ID,
		Success: true,
	}

	// Simulate processing time
	processingTime := time.Millisecond * time.Duration(50+job.Type*10)
	time.Sleep(processingTime)

	switch job.Type {
	case JobTypeValidate:
		result.Message = "Book validation completed"
		if job.BookData != nil && job.BookData.Title == "" {
			result.Success = false
			result.Error = fmt.Errorf("title is required")
		}

	case JobTypeProcess:
		result.Message = "Book processing completed"
		// Simulate some processing logic
		if job.BookData != nil {
			result.Book = &models.Book{
				Title:  job.BookData.Title,
				Author: job.BookData.Author,
				ISBN:   job.BookData.ISBN,
			}
		}

	case JobTypeNotify:
		result.Message = "Notification sent"
		// Simulate notification logic

	default:
		result.Success = false
		result.Error = fmt.Errorf("unknown job type: %d", job.Type)
	}

	return result
}

// resultProcessor handles job results
func (bp *BookProcessor) resultProcessor() {
	defer bp.wg.Done()

	log.Println("Result processor started")

	for {
		select {
		case result, ok := <-bp.resultChan:
			if !ok {
				log.Println("Result processor: result channel closed, exiting")
				return
			}

			log.Printf("Processing result for job %s: success=%t, message=%s",
				result.JobID, result.Success, result.Message)

			// Execute callback if provided
			if result.Success {
				log.Printf("Job %s completed successfully: %s", result.JobID, result.Message)
			} else {
				log.Printf("Job %s failed: %v", result.JobID, result.Error)
			}

		case <-bp.ctx.Done():
			log.Println("Result processor: context cancelled, exiting")
			return
		}
	}
}

// Shutdown gracefully shuts down the worker pool
func (bp *BookProcessor) Shutdown(ctx context.Context) error {
	bp.cancel() // Cancel the context to stop workers

	// Wait for all workers to finish with context timeout
	done := make(chan struct{})
	go func() {
		bp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetMetrics returns worker pool metrics
func (bp *BookProcessor) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"workers":            bp.workers,
		"queue_capacity":     cap(bp.jobQueue),
		"current_queue_size": len(bp.jobQueue),
	}
}

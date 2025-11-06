package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// executeJob executes a job with simulated work
// In Phase 2, this will integrate with the cache service
func (w *Worker) executeJob(jobID, payload string) error {
	log.Printf("[EXEC] Worker %s executing job %s (payload: %s)", w.ID, jobID, payload)

	// Simulate job execution time (1-3 seconds)
	executionTime := time.Duration(1+rand.Intn(3)) * time.Second

	// Sleep to simulate work
	time.Sleep(executionTime)

	// Simulate 10% failure rate for testing retry logic
	if rand.Float32() < 0.1 {
		return fmt.Errorf("simulated job failure")
	}

	log.Printf("[DONE] Worker %s completed job %s in %v", w.ID, jobID, executionTime)
	return nil
}

// executeJobWithTimeout executes a job with a timeout
func (w *Worker) executeJobWithTimeout(ctx context.Context, jobID, payload string) error {
	done := make(chan error, 1)

	// Execute job in goroutine
	go func() {
		done <- w.executeJob(jobID, payload)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("job timed out after %v", ctx.Err())
	}
}

// ExecuteJobWithRetry executes a job with exponential backoff retry logic
func (w *Worker) ExecuteJobWithRetry(jobID, payload string, maxRetries int) error {
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt seconds
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("[RETRY] Worker %s retrying job %s (attempt %d/%d) after %v",
				w.ID, jobID, attempt, maxRetries, backoff)
			time.Sleep(backoff)
		}

		err = w.executeJob(jobID, payload)
		if err == nil {
			return nil // Success!
		}

		log.Printf("[WARN] Worker %s job %s attempt %d failed: %v",
			w.ID, jobID, attempt+1, err)
	}

	return fmt.Errorf("job failed after %d attempts: %w", maxRetries+1, err)
}

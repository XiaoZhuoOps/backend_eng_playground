package OOD

import (
	"context"
	"errors"
	"time"
)

// TaskExecutor defines an interface for executing tasks with a timeout.
type TaskExecutor interface {
	// Execute runs the given task after the specified delay.
	// If the task execution exceeds the duration, it is canceled.
	Execute(task func(ctx context.Context) error, delay time.Duration, timeout time.Duration) error
}

// TaskExecutorImpl is the implementation of the TaskExecutor interface.
type TaskExecutorImpl struct{}

// Execute implements the TaskExecutor interface.
func (te *TaskExecutorImpl) Execute(task func(ctx context.Context) error, delay time.Duration, timeout time.Duration) error {
	// Create a context that will be canceled after the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a timer for the delay.
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		// Start the task in a separate goroutine.
		done := make(chan error, 1)
		go func() {
			done <- task(ctx)
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return errors.New("task execution exceeded timeout and was canceled")
		}
	case <-ctx.Done():
		return errors.New("execution canceled before task could start")
	}
}

// Example usage
func ExampleTaskExecutor() {
	foo := func() error {
		time.Sleep(2 * time.Second) // simulate a long-running task such as wait for a response from a remote server
		return nil
	}

	task := func(ctx context.Context) error {
		// Simulate a long-running task
		result := make(chan error, 1)
		go func() {
			result <- foo()
		}()
		select {
		// case <-time.After(2 * time.Second):
		case <-result:
			// Task completed
			return nil
		case <-ctx.Done():
			// Task was canceled
			return ctx.Err()
		}
	}

	executor := &TaskExecutorImpl{}

	err := executor.Execute(task, 1*time.Second, 3*time.Second)
	if err != nil {
		// Handle error
	}

	err = executor.Execute(task, 1*time.Second, 1*time.Second)
	if err != nil {
		// Handle timeout error
	}
}

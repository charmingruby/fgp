package task_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gustavodias/fgp/task"
)

func ExampleTask_retryTimeout() {
	fetch := task.From(func(_ context.Context) (string, error) {
		return "payload", nil
	})
	retries := task.Retry(fetch, task.RetryConfig{
		Attempts: 3,
		Delay:    10 * time.Millisecond,
		ShouldRetry: func(err error) bool {
			return !errors.Is(err, context.Canceled)
		},
	})
	deadline := task.Timeout(retries, 100*time.Millisecond)
	value, _ := deadline(context.Background())
	fmt.Println(value)
	// Output:
	// payload
}

# Worker with Retry, Backoff, and Cancellation

## Problem
Run an idempotent background worker that polls jobs, retries transient failures with backoff, and stops cleanly on shutdown signals.

## Functional Design
```go
var poll = task.From(func(ctx context.Context) (Job, error) {
	return queue.Next(ctx)
})

process := func(job Job) task.Task[struct{}] {
	return task.From(func(ctx context.Context) (struct{}, error) {
		return struct{}{}, handler.Handle(ctx, job)
	})
}

var runJob = task.FlatMap(poll, process)

var resilient = task.Retry(runJob, task.RetryConfig{
	Attempts: 5,
	Delay:    100 * time.Millisecond,
	Backoff: func(attempt int, err error) time.Duration {
		return time.Duration(attempt) * 100 * time.Millisecond
	},
	ShouldRetry: func(err error) bool {
		return errors.Is(err, ErrTransient)
	},
})

func worker(ctx context.Context) error {
	for {
		if _, err := resilient(ctx); err != nil {
			return err // ctx.Err or permanent failure
		}
	}
}
```

## Imperative Comparison
Conventional loops often inline retry logic with mutable counters, manual timer management, and ad-hoc context checks. Modeling the pipeline as `task.Retry(task.FlatMap(...))` makes cancellation propagation automatic, and the backoff strategy stays declarative.

## Performance Notes
- `Retry` allocates once for the wrapper closure; attempts reuse the same job Task.
- Using `timeutil.Sleep` avoids goroutine leaks because cancellation wakes sleepers immediately.
- No channels or goroutines are spawned unless the worker body does so explicitly.

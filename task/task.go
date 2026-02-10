// Package task defines context-aware effectful computations and combinators.
//
// Example:
//
//	getUser := task.From(func(ctx context.Context) (User, error) {
//		return repo.Load(ctx)
//	})
//	nameTask := task.Map(getUser, func(u User) string { return u.Name })
package task

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gustavodias/fgp/internal/timeutil"
	"github.com/gustavodias/fgp/option"
	"github.com/gustavodias/fgp/result"
)

// Task represents a computation that can be executed with a context.
//
// Example:
//
//	var fetchUser Task[User] = func(ctx context.Context) (User, error) {
//		return repo.Load(ctx)
//	}
type Task[T any] func(ctx context.Context) (T, error)

// From wraps an arbitrary context-aware function into a Task.
//
// Example:
//
//	fetch := From(repo.Load)
//	user, err := fetch(ctx)
func From[T any](fn func(ctx context.Context) (T, error)) Task[T] {
	return func(ctx context.Context) (T, error) {
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, err
		}
		return fn(ctx)
	}
}

// Pure lifts a value into a Task that respects cancellation.
//
// Example:
//
//	unit := Pure("ready")
//	value, _ := unit(ctx)
func Pure[T any](value T) Task[T] {
	return func(ctx context.Context) (T, error) {
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, err
		}
		return value, nil
	}
}

// Fail creates a Task that immediately fails with err (or context error if
// err is nil).
//
// Example:
//
//	failing := Fail[string](errors.New("boom"))
//	_, err := failing(ctx)
func Fail[T any](err error) Task[T] {
	failureErr := err
	if failureErr == nil {
		failureErr = errors.New("task: nil error")
	}
	return func(ctx context.Context) (T, error) {
		var zero T
		if ctxErr := ctx.Err(); ctxErr != nil {
			return zero, ctxErr
		}
		return zero, failureErr
	}
}

// Map transforms the Task result when it succeeds.
//
// Example:
//
//	getName := Map(fetchUser, func(u User) string { return u.Name })
func Map[T any, U any](t Task[T], fn func(T) U) Task[U] {
	return func(ctx context.Context) (U, error) {
		val, err := t(ctx)
		if err != nil {
			var zero U
			return zero, err
		}
		if err := ctx.Err(); err != nil {
			var zero U
			return zero, err
		}
		return fn(val), nil
	}
}

// FlatMap chains two Tasks.
//
// Example:
//
//	full := FlatMap(fetchUser, func(u User) Task[Profile] {
//		return fetchProfile(u.ID)
//	})
func FlatMap[T any, U any](t Task[T], fn func(T) Task[U]) Task[U] {
	return func(ctx context.Context) (U, error) {
		val, err := t(ctx)
		if err != nil {
			var zero U
			return zero, err
		}
		if err := ctx.Err(); err != nil {
			var zero U
			return zero, err
		}
		return fn(val)(ctx)
	}
}

// Tap executes fn on success and passes the value through unchanged.
//
// Example:
//
//	logged := Tap(fetchUser, func(u User) {
//		log.Println("loaded", u.ID)
//	})
func Tap[T any](t Task[T], fn func(T)) Task[T] {
	return func(ctx context.Context) (T, error) {
		val, err := t(ctx)
		if err == nil {
			fn(val)
		}
		return val, err
	}
}

// TapErr executes fn when the Task fails.
//
// Example:
//
//	withMetrics := TapErr(fetchUser, func(err error) {
//		metrics.Count("user.fail")
//	})
func TapErr[T any](t Task[T], fn func(error)) Task[T] {
	return func(ctx context.Context) (T, error) {
		val, err := t(ctx)
		if err != nil {
			fn(err)
		}
		return val, err
	}
}

// Ensure runs fn after the task completes, regardless of success.
//
// Example:
//
//	withCleanup := Ensure(fetchUser, func() { span.End() })
func Ensure[T any](t Task[T], fn func()) Task[T] {
	return func(ctx context.Context) (T, error) {
		val, err := t(ctx)
		fn()
		return val, err
	}
}

// Bracket ensures that release runs after use, even when errors occur.
//
// Example:
//
//	withConn := Bracket(acquireConn,
//		func(conn *sql.Conn) Task[Result] { return useConn(conn) },
//		func(ctx context.Context, conn *sql.Conn, err error) error { return conn.Close() },
//	)
func Bracket[A any, B any](
	acquire Task[A],
	use func(A) Task[B],
	release func(context.Context, A, error) error,
) Task[B] {
	return func(ctx context.Context) (B, error) {
		resource, err := acquire(ctx)
		if err != nil {
			var zero B
			return zero, err
		}
		value, useErr := use(resource)(ctx)
		releaseErr := release(ctx, resource, useErr)
		if releaseErr != nil {
			if useErr != nil {
				return value, errors.Join(useErr, releaseErr)
			}
			var zero B
			return zero, releaseErr
		}
		if useErr != nil {
			return value, useErr
		}
		return value, nil
	}
}

// Timeout bounds the execution time of a Task.
//
// Example:
//
//	fast := Timeout(fetchUser, 500*time.Millisecond)
func Timeout[T any](t Task[T], d time.Duration) Task[T] {
	if d <= 0 {
		return t
	}
	return func(ctx context.Context) (T, error) {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return t(ctxWithTimeout)
	}
}

// RetryConfig defines retry behavior for Retry.
//
// Example:
//
//	cfg := RetryConfig{Attempts: 3, Delay: 100 * time.Millisecond}
type RetryConfig struct { //nolint:govet // fieldalignment: keep numeric fields grouped for readability
	Attempts    int
	Delay       time.Duration
	Backoff     func(attempt int, err error) time.Duration
	ShouldRetry func(error) bool
}

// Retry re-executes the task according to cfg when it fails.
//
// Example:
//
//	withRetry := Retry(fetchUser, RetryConfig{Attempts: 5, Delay: time.Second})
func Retry[T any](t Task[T], cfg RetryConfig) Task[T] { //nolint:gocognit // branching handles retry policies
	return func(ctx context.Context) (T, error) {
		attempts := cfg.Attempts
		if attempts <= 0 {
			attempts = 1
		}
		var lastErr error
		var value T
		for attempt := 1; attempt <= attempts; attempt++ {
			if err := ctx.Err(); err != nil {
				var zero T
				return zero, err
			}
			value, lastErr = t(ctx)
			if lastErr == nil {
				return value, nil
			}
			if cfg.ShouldRetry != nil && !cfg.ShouldRetry(lastErr) {
				break
			}
			if attempt == attempts {
				break
			}
			delay := cfg.Delay
			if cfg.Backoff != nil {
				delay = cfg.Backoff(attempt, lastErr)
			}
			if delay < 0 {
				delay = 0
			}
			if !timeutil.Sleep(ctx, delay) {
				var zero T
				return zero, ctx.Err()
			}
		}
		var zero T
		return zero, lastErr
	}
}

// Sequence runs tasks sequentially.
//
// Example:
//
//	all := Sequence([]Task[string]{taskA, taskB})
func Sequence[T any](tasks []Task[T]) Task[[]T] {
	return func(ctx context.Context) ([]T, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		results := make([]T, 0, len(tasks))
		for _, t := range tasks {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			val, err := t(ctx)
			if err != nil {
				return nil, err
			}
			results = append(results, val)
		}
		return results, nil
	}
}

// SequencePar executes all tasks concurrently, failing fast on the first error.
//
// Example:
//
//	parallel := SequencePar([]Task[int]{taskA, taskB})
func SequencePar[T any](tasks []Task[T]) Task[[]T] {
	return TraverseParN(tasks, len(tasks), func(t Task[T]) Task[T] {
		return t
	})
}

// TraversePar executes fn for each input element concurrently.
//
// Example:
//
//	tasks := TraversePar(ids, func(id int) Task[User] { return fetchUserByID(id) })
func TraversePar[A any, B any](items []A, fn func(A) Task[B]) Task[[]B] {
	return TraverseParN(items, len(items), fn)
}

// TraverseParN is a bounded parallel traversal that limits concurrency to n.
//
// Example:
//
//	bounded := TraverseParN(urls, 4, func(url string) Task[*http.Response] {
//		return fetchURL(url)
//	})
func TraverseParN[A any, B any](items []A, n int, fn func(A) Task[B]) Task[[]B] {
	return func(ctx context.Context) ([]B, error) {
		if len(items) == 0 {
			return []B{}, nil
		}
		workers := clampParallelism(len(items), n)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		results := make([]B, len(items))
		jobs := make(chan workItem[A], len(items))
		errCh := make(chan error, 1)
		var wg sync.WaitGroup

		worker := func() {
			defer wg.Done()
			for job := range jobs {
				val, err := fn(job.item)(ctx)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					cancel()
					return
				}
				results[job.index] = val
			}
		}

		wg.Add(workers)
		for range workers {
			go worker()
		}

		enqueueWork(ctx, jobs, items)
		close(jobs)
		wg.Wait()

		if err := pullError(errCh); err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return results, nil
	}
}

type workItem[T any] struct { //nolint:govet // fieldalignment: generic payload size dominates; keep simple layout
	index int
	item  T
}

func clampParallelism(total, requested int) int {
	if requested <= 0 {
		return 1
	}
	if requested > total {
		return total
	}
	return requested
}

func enqueueWork[A any](ctx context.Context, jobs chan<- workItem[A], items []A) {
	for idx, item := range items {
		select {
		case <-ctx.Done():
			return
		case jobs <- workItem[A]{index: idx, item: item}:
		}
	}
}

func pullError(errCh <-chan error) error {
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// FromResult lifts an existing Result into a Task. Context cancellation takes
// precedence over the stored error.
//
// Example:
//
//	t := FromResult(result.Ok(42))
//	value, _ := t(ctx)
func FromResult[T any](res result.Result[T]) Task[T] {
	return func(ctx context.Context) (T, error) {
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, err
		}
		return res.Unwrap()
	}
}

// FromOption lifts an Option into a Task. When the Option is None, errFactory is
// used to produce the failure error; if nil, a descriptive error is used.
//
// Example:
//
//	t := FromOption(opt, func() error { return errors.New("missing user") })
func FromOption[T any](opt option.Option[T], errFactory func() error) Task[T] {
	return func(ctx context.Context) (T, error) {
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, err
		}
		if value, ok := opt.Get(); ok {
			return value, nil
		}
		var err error
		if errFactory != nil {
			err = errFactory()
		}
		if err == nil {
			err = errors.New("task: option is none")
		}
		var zero T
		return zero, err
	}
}

// ToResultTask converts a Task into one that never fails (except for context
// cancellation) and instead wraps the outcome in a Result.
//
// Example:
//
//	wrapped := ToResultTask(fetchUser)
//	res, err := wrapped(ctx)
//	if err != nil {
//		return err // context cancellation
//	}
func ToResultTask[T any](t Task[T]) Task[result.Result[T]] {
	return func(ctx context.Context) (result.Result[T], error) {
		val, err := t(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, ctxErr) {
				return result.Result[T]{}, err
			}
			return result.Err[T](err), nil
		}
		return result.Ok(val), nil
	}
}

package task_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charmingruby/fgp/option"
	"github.com/charmingruby/fgp/result"
	"github.com/charmingruby/fgp/task"
)

func TestRetryEventuallySucceeds(t *testing.T) {
	var attempts atomic.Int32
	work := task.From(func(_ context.Context) (int, error) {
		if attempts.Add(1) < 3 {
			return 0, errors.New("not yet")
		}
		return 7, nil
	})
	retried := task.Retry(work, task.RetryConfig{Attempts: 5, Delay: time.Millisecond})
	value, err := retried(context.Background())
	if err != nil || value != 7 {
		t.Fatalf("unexpected retry result %v %v", value, err)
	}
}

func TestTimeout(t *testing.T) {
	work := task.From(func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return 1, nil
		}
	})
	to := task.Timeout(work, 10*time.Millisecond)
	_, err := to(context.Background())
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestSequenceParCancelsOnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tasks := []task.Task[int]{
		task.Pure(1),
		task.Fail[int](errors.New("boom")),
		task.From(func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		}),
	}
	_, err := task.SequencePar(tasks)(ctx)
	if err == nil {
		t.Fatalf("expected error from sequence par")
	}
}

func TestTraverseParNRespectsLimit(t *testing.T) {
	var current atomic.Int32
	var peak atomic.Int32
	items := []int{1, 2, 3, 4, 5}
	fn := func(v int) task.Task[int] {
		return task.From(func(_ context.Context) (int, error) {
			n := current.Add(1)
			updatePeak(&peak, n)
			time.Sleep(5 * time.Millisecond)
			current.Add(-1)
			return v * 2, nil
		})
	}
	limit := 2
	values, err := task.TraverseParN(items, limit, fn)(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if peak.Load() > int32(limit) {
		t.Fatalf("expected concurrency <= %d, got %d", limit, peak.Load())
	}
	if len(values) != len(items) {
		t.Fatalf("unexpected length")
	}
}

func TestTraverseParNPreservesOrder(t *testing.T) {
	items := []int{1, 2, 3, 4}
	fn := func(v int) task.Task[int] {
		return task.From(func(_ context.Context) (int, error) {
			time.Sleep(time.Duration(5-v) * time.Millisecond)
			return v * 3, nil
		})
	}
	values, err := task.TraverseParN(items, 2, fn)(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, v := range items {
		if values[i] != v*3 {
			t.Fatalf("order mismatch at %d", i)
		}
	}
}

func TestRaceAndParZip(t *testing.T) {
	fast := task.From(func(context.Context) (int, error) { return 1, nil })
	slow := task.From(func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return 2, nil
		}
	})
	winner, err := task.Race(slow, fast)(context.Background())
	if err != nil || winner != 1 {
		t.Fatalf("expected fast task to win, got %v %v", winner, err)
	}
	_, err = task.Race[int]()(context.Background())
	if err == nil {
		t.Fatalf("expected error when no tasks provided")
	}
	combo := task.ParZip(fast, task.Pure("user"))
	pair, err := combo(context.Background())
	if err != nil || pair.First != 1 || pair.Second != "user" {
		t.Fatalf("unexpected parzip result %v %v", pair, err)
	}
	failing := task.ParZip(task.Fail[int](errors.New("boom")), task.Pure("ok"))
	if _, err := failing(context.Background()); err == nil {
		t.Fatalf("expected parzip failure on error")
	}
}

func TestParMapNAndBoth(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	var peak atomic.Int32
	var inFlight atomic.Int32
	fn := func(ctx context.Context, v int) (int, error) {
		n := inFlight.Add(1)
		updatePeak(&peak, n)
		select {
		case <-ctx.Done():
			inFlight.Add(-1)
			return 0, ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
		inFlight.Add(-1)
		return v * 10, nil
	}
	values, err := task.ParMapN(items, 2, fn)(context.Background())
	if err != nil {
		t.Fatalf("unexpected parmap error: %v", err)
	}
	if peak.Load() > 2 {
		t.Fatalf("expected concurrency <= 2, got %d", peak.Load())
	}
	for i, v := range items {
		if values[i] != v*10 {
			t.Fatalf("unexpected parmap output")
		}
	}
	tuple, err := task.Both(task.Pure(1), task.Pure("x"))(context.Background())
	if err != nil || tuple.First != 1 || tuple.Second != "x" {
		t.Fatalf("unexpected both result %v %v", tuple, err)
	}
}

func TestDelayAndAttempt(t *testing.T) {
	start := time.Now()
	_, err := task.Delay(10 * time.Millisecond)(context.Background())
	if err != nil {
		t.Fatalf("unexpected delay error: %v", err)
	}
	if time.Since(start) < 5*time.Millisecond {
		t.Fatalf("delay returned too early")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := task.Delay(time.Second)(ctx); err == nil {
		t.Fatalf("expected delay cancellation")
	}
	safe := task.Attempt(task.Task[int](func(context.Context) (int, error) {
		panic("boom")
	}))
	if _, err := safe(context.Background()); err == nil {
		t.Fatalf("expected panic converted to error")
	}
}

func TestSequenceRespectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	tasks := []task.Task[int]{
		task.From(func(_ context.Context) (int, error) {
			count++
			cancel()
			return 1, nil
		}),
		task.From(func(_ context.Context) (int, error) {
			count++
			return 2, nil
		}),
	}
	_, err := task.Sequence(tasks)(ctx)
	if err == nil {
		t.Fatalf("expected cancellation error")
	}
	if count != 1 {
		t.Fatalf("expected second task skipped")
	}
}

func TestBracketJoinsErrors(t *testing.T) {
	acquire := task.Pure(1)
	useErr := errors.New("use failed")
	use := func(_ int) task.Task[int] {
		return task.Fail[int](useErr)
	}
	releaseErr := errors.New("release failed")
	release := func(_ context.Context, _ int, _ error) error {
		return releaseErr
	}
	_, err := task.Bracket(acquire, use, release)(context.Background())
	if err == nil || !errors.Is(err, useErr) || !errors.Is(err, releaseErr) {
		t.Fatalf("expected joined errors containing both, got %v", err)
	}
}

func TestRetryNegativeDelay(t *testing.T) {
	var attempts atomic.Int32
	work := task.From(func(_ context.Context) (int, error) {
		if attempts.Add(1) >= 2 {
			return 9, nil
		}
		return 0, errors.New("boom")
	})
	retry := task.Retry(work, task.RetryConfig{Attempts: 3, Delay: -time.Second})
	value, err := retry(context.Background())
	if err != nil || value != 9 {
		t.Fatalf("unexpected retry output")
	}
}

func TestInteropHelpers(t *testing.T) {
	resTask := task.FromResult(result.Ok(5))
	value, err := resTask(context.Background())
	if err != nil || value != 5 {
		t.Fatalf("unexpected from result output")
	}
	opTask := task.FromOption(option.Some("x"), nil)
	opValue, err := opTask(context.Background())
	if err != nil || opValue != "x" {
		t.Fatalf("unexpected from option output")
	}
	_, err = task.FromOption(option.None[string](), func() error { return errors.New("missing") })(context.Background())
	if err == nil {
		t.Fatalf("expected error when option none")
	}
	wrap := task.ToResultTask(task.Fail[int](errors.New("oops")))
	wrapped, err := wrap(context.Background())
	if err != nil {
		t.Fatalf("expected nil task error, got %v", err)
	}
	if wrapped.IsOk() {
		t.Fatalf("expected wrapped err")
	}
}

func updatePeak(peak *atomic.Int32, value int32) {
	for {
		old := peak.Load()
		if value <= old {
			return
		}
		if peak.CompareAndSwap(old, value) {
			return
		}
	}
}

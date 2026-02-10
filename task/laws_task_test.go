package task_test

import (
	"context"
	"errors"
	"testing"
	"testing/quick"

	"github.com/charmingruby/fgp/task"
)

func TestTaskMapIdentityLaw(t *testing.T) {
	identity := func(x int) int { return x }
	check := func(value int) bool {
		base := task.Pure(value)
		mapped := task.Map(base, identity)
		return equalTasks(base, mapped)
	}
	if err := quick.Check(check, nil); err != nil {
		t.Fatalf("map identity law failed: %v", err)
	}
}

func TestTaskFlatMapAssociativity(t *testing.T) {
	f := func(x int) task.Task[int] {
		return task.From(func(ctx context.Context) (int, error) {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			if x%2 == 0 {
				return x / 2, nil
			}
			return 0, errors.New("odd")
		})
	}
	g := func(x int) task.Task[int] {
		return task.Map(task.Pure(x), func(v int) int { return v + 3 })
	}
	check := func(value int) bool {
		left := task.FlatMap(task.FlatMap(task.Pure(value), f), g)
		right := task.FlatMap(task.Pure(value), func(v int) task.Task[int] {
			return task.FlatMap(f(v), g)
		})
		return equalTasks(left, right)
	}
	if err := quick.Check(check, nil); err != nil {
		t.Fatalf("flatmap associativity failed: %v", err)
	}
}

func TestTaskContextErrorsTakePrecedence(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	failing := task.Fail[int](errors.New("boom"))
	_, err := failing(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	mapped := task.Map(failing, func(v int) int { return v })
	_, err = mapped(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("map should propagate context cancellation")
	}
}

func equalTasks[T comparable](a task.Task[T], b task.Task[T]) bool {
	ctx := context.Background()
	av, aerr := a(ctx)
	bv, berr := b(ctx)
	if (aerr == nil) != (berr == nil) {
		return false
	}
	if aerr != nil && berr != nil {
		return true
	}
	return av == bv
}

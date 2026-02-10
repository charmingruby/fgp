package result_test

import (
	"errors"
	"testing"
	"testing/quick"

	"github.com/charmingruby/fgp/result"
)

func TestResultFunctorLaws(t *testing.T) {
	id := func(x int) int { return x }
	inc := func(x int) int { return x + 1 }
	dbl := func(x int) int { return x * 2 }

	check := func(value int, ok bool) bool {
		var res result.Result[int]
		if ok {
			res = result.Ok(value)
		} else {
			res = result.Err[int](errors.New("boom"))
		}
		left := result.Map(result.Map(res, inc), dbl)
		right := result.Map(res, func(v int) int { return dbl(inc(v)) })
		return equalResult(res, result.Map(res, id)) && equalResult(left, right)
	}

	if err := quick.Check(check, nil); err != nil {
		t.Fatalf("functor laws failed: %v", err)
	}
}

func TestResultMonadLaws(t *testing.T) {
	f := func(x int) result.Result[int] {
		if x%2 == 0 {
			return result.Ok(x / 2)
		}
		return result.Err[int](errors.New("odd"))
	}
	g := func(x int) result.Result[int] {
		return result.Ok(x + 3)
	}

	leftIdentity := func(x int) bool {
		return equalResult(result.FlatMap(result.Ok(x), f), f(x))
	}
	if err := quick.Check(leftIdentity, nil); err != nil {
		t.Fatalf("left identity failed: %v", err)
	}

	rightIdentity := func(value int, ok bool) bool {
		var res result.Result[int]
		if ok {
			res = result.Ok(value)
		} else {
			res = result.Err[int](errors.New("fail"))
		}
		return equalResult(result.FlatMap(res, result.Ok[int]), res)
	}
	if err := quick.Check(rightIdentity, nil); err != nil {
		t.Fatalf("right identity failed: %v", err)
	}

	associativity := func(value int) bool {
		left := result.FlatMap(result.FlatMap(result.Ok(value), f), g)
		right := result.FlatMap(result.Ok(value), func(v int) result.Result[int] {
			return result.FlatMap(f(v), g)
		})
		return equalResult(left, right)
	}
	if err := quick.Check(associativity, nil); err != nil {
		t.Fatalf("associativity failed: %v", err)
	}
}

func equalResult[T comparable](a, b result.Result[T]) bool {
	if a.IsOk() != b.IsOk() {
		return false
	}
	if !a.IsOk() {
		return true
	}
	return a.UnwrapOr(zero[T]()) == b.UnwrapOr(zero[T]())
}

func zero[T any]() T {
	var z T
	return z
}

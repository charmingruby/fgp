package option_test

import (
	"errors"
	"testing"
	"testing/quick"

	"github.com/charmingruby/fgp/option"
)

func TestSomeNilBehavior(t *testing.T) {
	var value any
	opt := option.Some(value)
	if opt.IsNone() {
		t.Fatalf("expected Some(nil) to be considered present")
	}
	got, ok := opt.Get()
	if !ok || got != nil {
		t.Fatalf("expected stored nil, got %v present %v", got, ok)
	}
}

func TestZeroValueIsNone(t *testing.T) {
	var zero option.Option[int]
	if !zero.IsNone() {
		t.Fatalf("zero value should be None")
	}
	if zero.ToPtr() != nil {
		t.Fatalf("zero value should not yield pointer")
	}
}

func TestOptionToResult(t *testing.T) {
	opt := option.Some(42)
	res := opt.ToResult(func() error { return errors.New("missing") })
	if res.IsErr() {
		t.Fatalf("expected Ok, got err %v", res.Err())
	}
	if got := res.UnwrapOr(0); got != 42 {
		t.Fatalf("unexpected value %d", got)
	}

	none := option.None[int]()
	res = none.ToResult(func() error { return errors.New("boom") })
	if res.IsOk() {
		t.Fatalf("expected Err result")
	}
}

func TestOptionFunctorLaws(t *testing.T) {
	identity := func(x int) int { return x }
	composition := func(x int) int { return x + 1 }
	other := func(x int) int { return x * 2 }

	check := func(value int, present bool) bool {
		var opt option.Option[int]
		if present {
			opt = option.Some(value)
		} else {
			opt = option.None[int]()
		}
		idMapped := option.Map(opt, identity)
		compMapped := option.Map(option.Map(opt, composition), other)
		composed := option.Map(opt, func(x int) int { return other(composition(x)) })
		return equalOption(opt, idMapped) && equalOption(compMapped, composed)
	}

	if err := quick.Check(check, nil); err != nil {
		t.Fatalf("functor law failed: %v", err)
	}
}

func TestOptionMonadLaws(t *testing.T) {
	f := func(x int) option.Option[int] {
		if x%2 == 0 {
			return option.Some(x / 2)
		}
		return option.None[int]()
	}
	g := func(x int) option.Option[int] {
		return option.Some(x + 3)
	}
	leftIdentity := func(x int) bool {
		return equalOption(option.FlatMap(option.Some(x), f), f(x))
	}
	if err := quick.Check(leftIdentity, nil); err != nil {
		t.Fatalf("left identity failed: %v", err)
	}

	rightIdentity := func(present bool, x int) bool {
		var opt option.Option[int]
		if present {
			opt = option.Some(x)
		} else {
			opt = option.None[int]()
		}
		return equalOption(option.FlatMap(opt, option.Some[int]), opt)
	}
	if err := quick.Check(rightIdentity, nil); err != nil {
		t.Fatalf("right identity failed: %v", err)
	}

	associativity := func(x int) bool {
		left := option.FlatMap(option.FlatMap(option.Some(x), f), g)
		right := option.FlatMap(option.Some(x), func(v int) option.Option[int] {
			return option.FlatMap(f(v), g)
		})
		return equalOption(left, right)
	}
	if err := quick.Check(associativity, nil); err != nil {
		t.Fatalf("associativity failed: %v", err)
	}
}

func TestOptionFilter(t *testing.T) {
	opt := option.Some(10)
	if opt.Filter(func(v int) bool { return v > 10 }).IsSome() {
		t.Fatalf("expected filter to drop value")
	}
	if !opt.Filter(func(v int) bool { return v == 10 }).IsSome() {
		t.Fatalf("expected filter to keep value")
	}
}

func TestOptionInterop(t *testing.T) {
	opt := option.FromOk(5, true)
	ptr := opt.ToPtr()
	if ptr == nil || *ptr != 5 {
		t.Fatalf("expected pointer copy")
	}
	fromPtr := option.FromPtr(ptr)
	if fromPtr.IsNone() {
		t.Fatalf("expected value from pointer")
	}
	none := option.FromPtr[int](nil)
	if none.IsSome() {
		t.Fatalf("expected none from nil ptr")
	}
	fromOkNone := option.FromOk(1, false)
	if fromOkNone.IsSome() {
		t.Fatalf("expected none from ok=false")
	}
}

func equalOption[T comparable](a, b option.Option[T]) bool {
	av, aok := a.Get()
	bv, bok := b.Get()
	if aok != bok {
		return false
	}
	if !aok {
		return true
	}
	return av == bv
}

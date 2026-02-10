package option_test

import (
	"testing"
	"testing/quick"

	"github.com/charmingruby/fgp/option"
)

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

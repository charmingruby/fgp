package option_test

import (
	"errors"
	"testing"

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

func TestOptionFilter(t *testing.T) {
	opt := option.Some(10)
	if opt.Filter(func(v int) bool { return v > 10 }).IsSome() {
		t.Fatalf("expected filter to drop value")
	}
	if !opt.Filter(func(v int) bool { return v == 10 }).IsSome() {
		t.Fatalf("expected filter to keep value")
	}
}

func TestOptionTap(t *testing.T) {
	calls := 0
	opt := option.Tap(option.Some(5), func(v int) {
		if v != 5 {
			t.Fatalf("unexpected value %d", v)
		}
		calls++
	})
	if opt.IsNone() {
		t.Fatalf("expected tap to keep value")
	}
	if calls != 1 {
		t.Fatalf("expected single call, got %d", calls)
	}
	none := option.Tap(option.None[int](), func(int) { calls++ })
	if none.IsSome() {
		t.Fatalf("expected none to stay none")
	}
	if calls != 1 {
		t.Fatalf("tap should not run for none")
	}
}

func TestOptionZipTraverseSequence(t *testing.T) {
	zip := option.Zip(option.Some("a"), option.Some(2))
	if zip.IsNone() {
		t.Fatalf("expected zipped value")
	}
	pair, _ := zip.Get()
	if pair.First != "a" || pair.Second != 2 {
		t.Fatalf("unexpected pair %+v", pair)
	}
	zipNone := option.Zip(option.Some(1), option.None[int]())
	if zipNone.IsSome() {
		t.Fatalf("zip should short circuit")
	}
	seq := option.Sequence([]option.Option[int]{option.Some(1), option.Some(2)})
	if seq.IsNone() {
		t.Fatalf("expected successful sequence")
	}
	values, _ := seq.Get()
	if len(values) != 2 || values[1] != 2 {
		t.Fatalf("unexpected values: %v", values)
	}
	traverse := option.Traverse([]int{1, 2, 3}, func(v int) option.Option[int] {
		if v == 2 {
			return option.None[int]()
		}
		return option.Some(v * 2)
	})
	if traverse.IsSome() {
		t.Fatalf("expected traverse failure on drop")
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

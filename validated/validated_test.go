package validated_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/charmingruby/fgp/result"
	"github.com/charmingruby/fgp/validated"
)

func TestValidatedBasics(t *testing.T) {
	v := validated.Valid[string](10)
	mapped := validated.Map[string](v, func(n int) int { return n * 2 })
	if !mapped.IsValid() || mapped.UnsafeValue() != 20 {
		t.Fatalf("expected mapped value")
	}
	inv := validated.Invalid[string, int]("a", "b")
	if inv.IsValid() || len(inv.Errors()) != 2 {
		t.Fatalf("expected invalid state with errors")
	}
}

func TestZipSequenceTraverse(t *testing.T) {
	a := validated.Valid[string](1)
	b := validated.Valid[string](2)
	zip := validated.Zip(a, b)
	if !zip.IsValid() || zip.UnsafeValue().First != 1 || zip.UnsafeValue().Second != 2 {
		t.Fatalf("zip should combine values")
	}
	combined := validated.Zip(validated.Invalid[string, int]("err1"), b)
	if combined.IsValid() || len(combined.Errors()) != 1 {
		t.Fatalf("zip should accumulate errors")
	}
	seq := validated.Sequence([]validated.Validated[string, int]{
		validated.Valid[string](1),
		validated.Valid[string](2),
	})
	if !seq.IsValid() || len(seq.UnsafeValue()) != 2 {
		t.Fatalf("sequence should produce values")
	}
	seqErr := validated.Sequence([]validated.Validated[string, int]{
		validated.Valid[string](1),
		validated.Invalid[string, int]("boom"),
	})
	if seqErr.IsValid() {
		t.Fatalf("sequence should surface errors")
	}
	trav := validated.Traverse([]int{1, 2, 3}, func(v int) validated.Validated[string, int] {
		if v%2 == 0 {
			return validated.Invalid[string, int]("even")
		}
		return validated.Valid[string](v)
	})
	if trav.IsValid() || !reflect.DeepEqual(trav.Errors(), []string{"even"}) {
		t.Fatalf("expected traversal error")
	}
}

func TestResultInterop(t *testing.T) {
	res := validated.FromResult(result.Ok(5))
	if !res.IsValid() {
		t.Fatalf("expected valid from result")
	}
	failure := validated.FromResult(result.Err[int](errors.New("boom")))
	if failure.IsValid() {
		t.Fatalf("expected invalid state")
	}
	back := validated.ToResult(failure)
	if back.IsOk() {
		t.Fatalf("expected error result")
	}
}

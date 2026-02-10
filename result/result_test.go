package result_test

import (
	"errors"
	"testing"

	"github.com/charmingruby/fgp/result"
)

func TestZipAndSequence(t *testing.T) {
	left := result.Ok(1)
	right := result.Ok(2)
	zip := result.Zip2(left, right)
	if zip.IsErr() {
		t.Fatalf("expected zip ok: %v", zip.Err())
	}
	if zip.UnwrapOr(result.Tuple2[int, int]{}).First != 1 {
		t.Fatalf("unexpected first value")
	}
	seq := result.Sequence([]result.Result[int]{result.Ok(1), result.Ok(2)})
	if seq.IsErr() {
		t.Fatalf("sequence failed: %v", seq.Err())
	}
	values := seq.UnwrapOr(nil)
	if len(values) != 2 {
		t.Fatalf("unexpected length")
	}
}

func TestResultFlatMapErrAndCollect(t *testing.T) {
	boom := errors.New("boom")
	res := result.Err[int](boom)
	recovered := result.FlatMapErr(res, func(err error) result.Result[int] {
		if !errors.Is(err, boom) {
			t.Fatalf("unexpected err %v", err)
		}
		return result.Ok(10)
	})
	if recovered.IsErr() {
		t.Fatalf("expected recovery: %v", recovered.Err())
	}
	unchanged := result.FlatMapErr(result.Ok(1), func(err error) result.Result[int] {
		t.Fatalf("should not run: %v", err)
		return result.Err[int](err)
	})
	if unchanged.UnwrapOr(0) != 1 {
		t.Fatalf("expected unchanged ok value")
	}
	results := []result.Result[int]{result.Ok(1), result.Err[int](boom), result.Ok(2)}
	values := result.Collect(results)
	if len(values) != 2 {
		t.Fatalf("expected 2 successes got %d", len(values))
	}
	ok, errs := result.PartitionResults(results)
	if len(ok) != 2 || len(errs) != 1 {
		t.Fatalf("unexpected partition output %v %v", ok, errs)
	}
	if !errors.Is(errs[0], boom) {
		t.Fatalf("unexpected error slice: %v", errs)
	}
}

func TestTraverse(t *testing.T) {
	items := []int{1, 2, 3}
	res := result.Traverse(items, func(v int) result.Result[int] {
		if v == 2 {
			return result.Err[int](errors.New("stop"))
		}
		return result.Ok(v * 2)
	})
	if res.IsOk() {
		t.Fatalf("expected error traversal")
	}
}

func TestTupleInterop(t *testing.T) {
	res := result.FromTuple(10, nil)
	if res.IsErr() {
		t.Fatalf("expected ok result")
	}
	value, err := res.ToTuple()
	if err != nil || value != 10 {
		t.Fatalf("unexpected tuple back %v %v", value, err)
	}
	failed := result.FromTuple(0, errors.New("boom"))
	if failed.IsOk() {
		t.Fatalf("expected error result")
	}
}

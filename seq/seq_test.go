package seq_test

import (
	"reflect"
	"testing"

	"github.com/charmingruby/fgp/seq"
)

func TestMapFilterReduce(t *testing.T) {
	src := []int{1, 2, 3, 4}
	mapped := seq.Map(src, func(v int) int { return v * v })
	if mapped[0] != 1 || mapped[3] != 16 {
		t.Fatalf("unexpected map output")
	}
	filtered := seq.Filter(mapped, func(v int) bool { return v%2 == 0 })
	if !reflect.DeepEqual(filtered, []int{4, 16}) {
		t.Fatalf("unexpected filter output %v", filtered)
	}
	red, ok := seq.Reduce(filtered, func(acc, next int) int { return acc + next })
	if !ok || red != 20 {
		t.Fatalf("unexpected reduce result")
	}
}

func TestGroupDistinctPartition(t *testing.T) {
	people := []struct {
		Name string
		City string
	}{
		{"Ana", "SP"},
		{"Joao", "RJ"},
		{"Bia", "SP"},
	}
	groups := seq.GroupBy(people, func(p struct {
		Name string
		City string
	}) string {
		return p.City
	})
	if len(groups["SP"]) != 2 {
		t.Fatalf("expected two in SP")
	}
	distinct := seq.DistinctBy([]string{"a", "b", "a"}, func(s string) string { return s })
	if len(distinct) != 2 {
		t.Fatalf("expected unique slice")
	}
	a, b := seq.Partition([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	if len(a) != 2 || len(b) != 2 {
		t.Fatalf("partition mismatch")
	}
}

func TestIteratorPipeline(t *testing.T) {
	it := seq.FromSlice([]int{1, 2, 3, 4})
	it = seq.Drop(it, 1)
	it = seq.Take(seq.MapIter(it, func(v int) int { return v * 10 }), 2)
	values := seq.ToSlice(it)
	if !reflect.DeepEqual(values, []int{20, 30}) {
		t.Fatalf("unexpected iterator output %v", values)
	}
}

func TestIteratorHelpers(t *testing.T) {
	values := seq.ToSlice(seq.Take(seq.Range(0, 5), 3))
	if !reflect.DeepEqual(values, []int{0, 1, 2}) {
		t.Fatalf("range values mismatch %v", values)
	}
	repeater := seq.Take(seq.Repeat("go"), 2)
	if got := seq.ToSlice(repeater); !reflect.DeepEqual(got, []string{"go", "go"}) {
		t.Fatalf("repeat mismatch %v", got)
	}
	iter := seq.TakeWhile(seq.Iterate(1, func(v int) int { return v * 2 }), func(v int) bool { return v < 10 })
	if got := seq.ToSlice(iter); !reflect.DeepEqual(got, []int{1, 2, 4, 8}) {
		t.Fatalf("iterate/takewhile mismatch %v", got)
	}
	dropped := seq.ToSlice(seq.DropWhile(seq.FromSlice([]int{0, 0, 3, 4}), func(v int) bool { return v == 0 }))
	if !reflect.DeepEqual(dropped, []int{3, 4}) {
		t.Fatalf("dropwhile mismatch %v", dropped)
	}
}

func TestChunkWindowScanCollect(t *testing.T) {
	given := []int{1, 2, 3, 4, 5}
	chunked := seq.Chunk(given, 2)
	if len(chunked) != 3 || !reflect.DeepEqual(chunked[0], []int{1, 2}) {
		t.Fatalf("chunk unexpected %v", chunked)
	}
	window := seq.Window([]int{1, 2, 3}, 2)
	if !reflect.DeepEqual(window, [][]int{{1, 2}, {2, 3}}) {
		t.Fatalf("window unexpected %v", window)
	}
	scan := seq.ScanLeft([]int{1, 2, 3}, 0, func(acc, v int) int { return acc + v })
	if !reflect.DeepEqual(scan, []int{0, 1, 3, 6}) {
		t.Fatalf("scan mismatch %v", scan)
	}
	collected := seq.Collect([]int{1, 2, 3, 4}, func(v int) (int, bool) {
		if v%2 == 0 {
			return v * v, true
		}
		return 0, false
	})
	if !reflect.DeepEqual(collected, []int{4, 16}) {
		t.Fatalf("collect mismatch %v", collected)
	}
}

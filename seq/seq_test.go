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

package fp_test

import (
	"testing"

	"github.com/charmingruby/fgp/fp"
)

func TestPipeComposeCurry(t *testing.T) {
	sum := func(a, b int) int { return a + b }
	curried := fp.Curry(sum)
	if curried(2)(3) != 5 {
		t.Fatalf("unexpected curry result")
	}
	pipeline := fp.Compose(
		func(i int) int { return i * 2 },
		func(i int) int { return i + 1 },
	)
	if pipeline(3) != 8 {
		t.Fatalf("compose result mismatch")
	}
	final := fp.Pipe(1, func(i int) int { return i + 1 }, func(i int) int { return i * 5 })
	if final != 10 {
		t.Fatalf("pipe result mismatch")
	}
}

func TestMaybe(t *testing.T) {
	trueBranchCalls := 0
	falseBranchCalls := 0
	value := fp.Maybe(true,
		func() int {
			trueBranchCalls++
			return 1
		},
		func() int {
			falseBranchCalls++
			return 2
		},
	)
	if value != 1 {
		t.Fatalf("expected true branch value")
	}
	if trueBranchCalls != 1 || falseBranchCalls != 0 {
		t.Fatalf("unexpected branch execution")
	}

	value = fp.Maybe(false,
		func() int {
			trueBranchCalls++
			return 3
		},
		func() int {
			falseBranchCalls++
			return 4
		},
	)
	if value != 4 {
		t.Fatalf("expected false branch value")
	}
	if trueBranchCalls != 1 || falseBranchCalls != 1 {
		t.Fatalf("unexpected branch counts after false path")
	}
}

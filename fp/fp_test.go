package fp_test

import (
	"testing"

	"github.com/gustavodias/fgp/fp"
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

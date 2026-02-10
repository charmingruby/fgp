package seq_test

import (
	"fmt"

	"github.com/gustavodias/fgp/seq"
)

func ExampleIterator_pipeline() {
	values := []int{1, 2, 3, 4}
	it := seq.FromSlice(values)
	it = seq.MapIter(it, func(v int) int { return v * 2 })
	it = seq.Take(it, 3)
	fmt.Println(seq.ToSlice(it))
	// Output:
	// [2 4 6]
}

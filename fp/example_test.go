package fp_test

import (
	"fmt"

	"github.com/gustavodias/fgp/fp"
)

func ExamplePipe() {
	add := func(v int) int { return v + 1 }
	mul := func(v int) int { return v * 2 }
	fmt.Println(fp.Pipe(2, add, mul))
	// Output:
	// 6
}

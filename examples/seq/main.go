// Package main demonstra iteradores lazy de seq.
package main

import (
	"fmt"

	"github.com/charmingruby/fgp/seq"
)

func main() {
	values := []int{1, 2, 3, 4}
	it := seq.FromSlice(values)
	it = seq.MapIter(it, func(v int) int { return v * 2 })
	it = seq.Take(it, 3)
	fmt.Println(seq.ToSlice(it)) //nolint:forbidigo // exemplos precisam imprimir
}

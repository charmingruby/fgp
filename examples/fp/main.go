// Package main demonstra Pipe/Compose do pacote fp.
package main

import (
	"fmt"

	"github.com/charmingruby/fgp/fp"
)

func main() {
	add := func(v int) int { return v + 1 }
	mul := func(v int) int { return v * 2 }
	fmt.Println(fp.Pipe(2, add, mul)) //nolint:forbidigo // exemplos precisam escrever no stdout
}

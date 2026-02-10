// Package main demonstra o uso de result.Traverse.
package main

import (
	"errors"
	"fmt"

	"github.com/gustavodias/fgp/result"
)

func main() {
	items := []int{1, 2, 3}
	res := result.Traverse(items, func(v int) result.Result[string] {
		if v == 2 {
			return result.Err[string](errors.New("downstream unavailable"))
		}
		return result.Ok(fmt.Sprintf("user-%d", v))
	})

	if res.IsOk() {
		fmt.Println("users:", res.UnwrapOr(nil)) //nolint:forbidigo // exemplos precisam imprimir
	} else {
		fmt.Println("error:", res.Err()) //nolint:forbidigo // exemplos precisam imprimir
	}
}

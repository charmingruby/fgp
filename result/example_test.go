package result_test

import (
	"errors"
	"fmt"

	"github.com/gustavodias/fgp/result"
)

func ExampleTraverse() {
	ids := []int{1, 2, 3}
	op := result.Traverse(ids, func(id int) result.Result[string] {
		if id == 2 {
			return result.Err[string](errors.New("downstream unavailable"))
		}
		return result.Ok(fmt.Sprintf("user-%d", id))
	})
	if op.IsOk() {
		fmt.Println(op.UnwrapOr(nil))
	} else {
		fmt.Println(op.Err())
	}
	// Output:
	// downstream unavailable
}

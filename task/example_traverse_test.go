package task_test

import (
	"context"
	"fmt"
	"time"

	"github.com/gustavodias/fgp/task"
)

func ExampleTraverseParN() {
	urls := []string{"a", "b", "c"}
	fetch := func(u string) task.Task[string] {
		return task.From(func(_ context.Context) (string, error) {
			time.Sleep(5 * time.Millisecond)
			return "ok:" + u, nil
		})
	}
	combined := task.TraverseParN(urls, 2, fetch)
	values, _ := combined(context.Background())
	fmt.Println(values)
	// Output:
	// [ok:a ok:b ok:c]
}

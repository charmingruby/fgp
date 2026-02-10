// Package main demonstra Tasks com retry/timeout e TraverseParN.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gustavodias/fgp/task"
)

func main() {
	demoRetryTimeout()
	demoTraverseParN()
}

func demoRetryTimeout() {
	fetch := task.From(func(context.Context) (string, error) {
		return "payload", nil
	})
	retries := task.Retry(fetch, task.RetryConfig{
		Attempts:    3,
		Delay:       10 * time.Millisecond,
		ShouldRetry: func(err error) bool { return !errors.Is(err, context.Canceled) },
	})
	deadline := task.Timeout(retries, 100*time.Millisecond)
	value, err := deadline(context.Background())
	fmt.Println("retry+timeout:", value, err) //nolint:forbidigo // exemplos precisam imprimir
}

func demoTraverseParN() {
	urls := []string{"a", "b", "c"}
	fetch := func(u string) task.Task[string] {
		return task.From(func(context.Context) (string, error) {
			time.Sleep(5 * time.Millisecond)
			return "ok:" + u, nil
		})
	}
	combined := task.TraverseParN(urls, 2, fetch)
	values, err := combined(context.Background())
	fmt.Println("traverseParN:", values, err) //nolint:forbidigo // exemplos precisam imprimir
}

# Batch Processing with TraverseParN

## Problem
Transform thousands of records in batches without overwhelming downstream services, while preserving ordering and surfacing the first failure immediately.

## Functional Design
```go
type Input struct { ID int }
type Output struct { ID int; Payload string }

func fetch(item Input) task.Task[Output] {
	return task.From(func(ctx context.Context) (Output, error) {
		payload, err := upstream.Fetch(ctx, item.ID)
		if err != nil {
			return Output{}, err
		}
		return Output{ID: item.ID, Payload: payload}, nil
	})
}

func process(ctx context.Context, inputs []Input) ([]Output, error) {
	return task.TraverseParN(inputs, 16, fetch)(ctx)
}
```

## Imperative Comparison
A hand-written worker pool typically manages wait groups, job channels, and manual cancellation. `TraverseParN` already enforces bounded concurrency, propagates cancellation when any job fails, and returns results in the original order, reducing surface area for leaks or races.

## Performance Notes
- Work queues size themselves to the number of items, so there are zero dynamic allocations per iteration beyond result storage.
- The helper uses a `sync.WaitGroup` and early cancellation to shrink wasted work.
- Passing `n` equal to the number of CPU-bound workers keeps goroutine count predictable.

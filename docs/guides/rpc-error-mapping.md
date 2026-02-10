# RPC Boundary Error Mapping

## Problem
Expose RPC methods that must distinguish between domain failures and effect failures (timeouts, cancellations) while keeping JSON-RPC responses deterministic.

## Functional Design
```go
type RPCError struct {
	Code    string
	Message string
}

func toRPCError(err error) RPCError {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return RPCError{Code: "timeout", Message: err.Error()}
	case errors.Is(err, ErrNotFound):
		return RPCError{Code: "not_found", Message: err.Error()}
	default:
		return RPCError{Code: "internal", Message: "unexpected failure"}
	}
}

func handler(ctx context.Context, req Request) Response {
	resTask := task.ToResultTask(
		task.Map(fetchUser(req.ID), hydrateResponse),
	)
	res, err := resTask(ctx)
	if err != nil {
		return Response{Error: toRPCError(err)}
	}
	return result.Fold(res,
		func(effectErr error) Response { return Response{Error: toRPCError(effectErr)} },
		func(value Payload) Response { return Response{Data: value} },
	)
}
```

## Imperative Comparison
Manual RPC handlers frequently mix domain and transport errors, leading to duplicated switch statements. Using `task.ToResultTask` captures effect errors (context cancellations) outside the `result.Result`, so clients can tell whether the request was canceled versus rejected.

## Performance Notes
- `ToResultTask` only allocates when wrapping an actual error, so success paths stay zero-alloc.
- Mapping errors once at the boundary avoids repeated string building deeper in the call stack.
- Context timeouts stop downstream RPCs early, preventing goroutine leaks.

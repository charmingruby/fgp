# HTTP Handler Pipeline with Task + Result

## Problem
Expose an HTTP endpoint that loads a user profile, enriches it, and renders JSON while guaranteeing request-scoped cancellation and deterministic error handling.

## Functional Design
```go
var fetchUser = task.From(func(ctx context.Context) (User, error) {
	return repo.Load(ctx)
})

var fetchPermissions = func(user User) task.Task[[]Permission] {
	return task.From(func(ctx context.Context) ([]Permission, error) {
		return acl.Load(ctx, user.ID)
	})
}

var enrich = task.FlatMap(fetchUser, func(user User) task.Task[result.Result[EnrichedUser]] {
	return task.ToResultTask(task.Map(fetchPermissions(user), func(perms []Permission) EnrichedUser {
		return EnrichedUser{User: user, Permissions: perms}
	}))
})

func HandleProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	res, err := enrich(ctx)
	if err != nil {
		writeError(w, err)
		return
	}
	summary := result.Map(res, toSummary)
	writeJSON(w, summary.UnwrapOrElse(func(e error) Summary {
		return Summary{Error: e.Error()}
	}))
}
```

## Imperative Comparison
A traditional handler would mix context checks, nested if/err, and manual logging. The combination of `task.FlatMap` and `result.Result` keeps the control flow explicit: context cancellation short-circuits before repo work, and JSON rendering never sees partial state.

## Performance Notes
- Tasks avoid goroutine-per-request by running inline and only allocating closures on composition.
- `task.ToResultTask` prevents extra allocations by reusing stack values when wrapping errors.
- `result.Map` is zero-alloc for success paths so the hot handler path is dominated by repository calls, not plumbing.

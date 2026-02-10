// Package result provides a success/error abstraction similar to Go's (T, error).
//
// Example:
//
//	res := result.Ok("done")
//	value, err := res.Unwrap()
//	_ = value
//
// Result combinators uphold Functor/Monad laws (see laws_result_test.go) to make
// transformations predictable even across retries and RPC boundaries.
package result

import "errors"

// Result represents the outcome of a computation that may succeed with a value
// or fail with an error. It never panics except in Unsafe helpers.
//
// Example:
//
//	res := result.Ok("token")
//	value, err := res.Unwrap()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(value)
type Result[T any] struct {
	value T
	err   error
}

// Ok constructs a successful Result carrying value.
//
// Example:
//
//	res := result.Ok(200)
//	fmt.Println(res.IsOk()) // true
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err constructs a failed Result. Passing a nil error automatically converts it
// into a descriptive placeholder to avoid silent successes.
//
// Example:
//
//	res := result.Err[int](errors.New("boom"))
//	_, err := res.Unwrap()
//	fmt.Println(err)
func Err[T any](err error) Result[T] {
	if err == nil {
		err = errors.New("result: nil error")
	}
	return Result[T]{err: err}
}

// FromTuple converts a standard Go (value, error) pair to a Result.
//
// Example:
//
//	value, err := repo.Load()
//	res := result.FromTuple(value, err)
//	return res
func FromTuple[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// IsOk reports whether the Result represents success.
//
// Example:
//
//	if res.IsOk() {
//		fmt.Println("success")
//	}
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr reports whether the Result represents failure.
//
// Example:
//
//	if res.IsErr() {
//		log.Println(res.Err())
//	}
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Err returns the stored error, if any.
//
// Example:
//
//	if err := res.Err(); err != nil {
//		return err
//	}
func (r Result[T]) Err() error {
	return r.err
}

// UnsafeUnwrap returns the underlying value or panics if the Result is an error.
//
// Example:
//
//	func mustConfig(res result.Result[Config]) Config {
//		return res.UnsafeUnwrap()
//	}
func (r Result[T]) UnsafeUnwrap() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.value
}

// Unwrap returns the value and error, mirroring standard Go semantics.
//
// Example:
//
//	value, err := res.Unwrap()
//	if err != nil {
//		return err
//	}
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// ToTuple exposes the underlying (value, error) pair, matching idiomatic Go
// callers that expect tuple returns.
//
// Example:
//
//	value, err := res.ToTuple()
func (r Result[T]) ToTuple() (T, error) {
	return r.value, r.err
}

// UnwrapOr returns the value when ok, otherwise returns fallback.
//
// Example:
//
//	code := res.UnwrapOr(http.StatusInternalServerError)
func (r Result[T]) UnwrapOr(fallback T) T {
	if r.err == nil {
		return r.value
	}
	return fallback
}

// UnwrapOrElse lazily computes a fallback using fn when the Result is an error.
//
// Example:
//
//	value := res.UnwrapOrElse(func(err error) string {
//		return "error: " + err.Error()
//	})
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	if r.err == nil {
		return r.value
	}
	return fn(r.err)
}

// Map transforms the value on success.
//
// Example:
//
//	length := result.Map(res, func(s string) int { return len(s) })
func Map[T any, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err == nil {
		return Ok(fn(r.value))
	}
	return Err[U](r.err)
}

// FlatMap chains computations, propagating the first error.
//
// Example:
//
//	res := result.FlatMap(loadUser(), fetchProfile)
func FlatMap[T any, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err == nil {
		return fn(r.value)
	}
	return Err[U](r.err)
}

// FlatMapErr chains error handlers, allowing recovery paths that still return Results.
//
// Example:
//
//	recovered := result.FlatMapErr(load(), func(err error) result.Result[Config] {
//		return loadFromFallback()
//	})
func FlatMapErr[T any](r Result[T], fn func(error) Result[T]) Result[T] {
	if r.err == nil {
		return r
	}
	if fn == nil {
		return r
	}
	return fn(r.err)
}

// MapErr transforms the stored error when present.
//
// Example:
//
//	res := result.MapErr(load(), func(err error) error {
//		return fmt.Errorf("wrap: %w", err)
//	})
func MapErr[T any](r Result[T], fn func(error) error) Result[T] {
	if fn == nil {
		return r
	}
	if r.err == nil {
		return r
	}
	return Err[T](fn(r.err))
}

// Recover converts an error Result into success using fn while keeping success
// values untouched.
//
// Example:
//
//	res := result.Recover(loadConfig(), func(err error) Config {
//		return defaultConfig
//	})
func Recover[T any](r Result[T], fn func(error) T) Result[T] {
	if r.err == nil {
		return r
	}
	return Ok(fn(r.err))
}

// Fold collapses the Result into a single value.
//
// Example:
//
//	message := result.Fold(res,
//		func(err error) string { return "failed: " + err.Error() },
//		func(val string) string { return "ok: " + val },
//	)
func Fold[T any, U any](r Result[T], onErr func(error) U, onOk func(T) U) U {
	if r.err == nil {
		return onOk(r.value)
	}
	return onErr(r.err)
}

// Tap executes fn when the Result is Ok and returns the original Result.
//
// Example:
//
//	_ = result.Tap(saveUser(), func(u User) {
//		metrics.Count("user_saved")
//	})
func Tap[T any](r Result[T], fn func(T)) Result[T] {
	if r.err == nil {
		fn(r.value)
	}
	return r
}

// TapErr executes fn when the Result is Err and returns the original Result.
//
// Example:
//
//	_ = result.TapErr(load(), func(err error) {
//		log.Println("load failed", err)
//	})
func TapErr[T any](r Result[T], fn func(error)) Result[T] {
	if r.err != nil {
		fn(r.err)
	}
	return r
}

// Collect gathers the successful values from the provided Results, ignoring failures.
// The returned slice never shares the backing array with inputs.
//
// Example:
//
//	values := result.Collect([]result.Result[int]{result.Ok(1), result.Err[int](err)})
func Collect[T any](results []Result[T]) []T {
	if len(results) == 0 {
		return []T{}
	}
	values := make([]T, 0, len(results))
	for _, r := range results {
		if r.err == nil {
			values = append(values, r.value)
		}
	}
	return values
}

// PartitionResults splits the input slice into successful values and collected errors.
//
// Example:
//
//	vals, errs := result.PartitionResults(results)
func PartitionResults[T any](results []Result[T]) ([]T, []error) {
	if len(results) == 0 {
		return []T{}, []error{}
	}
	values := make([]T, 0, len(results))
	errs := make([]error, 0, len(results))
	for _, r := range results {
		if r.err == nil {
			values = append(values, r.value)
			continue
		}
		errs = append(errs, r.err)
	}
	return values, errs
}

// Zip2 combines two results into one containing a pair of values.
//
// Example:
//
//	combined := result.Zip2(loadUser(), loadProfile())
func Zip2[A any, B any](ra Result[A], rb Result[B]) Result[Tuple2[A, B]] {
	if ra.err != nil {
		return Err[Tuple2[A, B]](ra.err)
	}
	if rb.err != nil {
		return Err[Tuple2[A, B]](rb.err)
	}
	return Ok(Tuple2[A, B]{First: ra.value, Second: rb.value})
}

// Zip3 combines three results into one containing a triple of values.
//
// Example:
//
//	combined := result.Zip3(loadUser(), loadProfile(), loadSettings())
func Zip3[A any, B any, C any](ra Result[A], rb Result[B], rc Result[C]) Result[Tuple3[A, B, C]] {
	if ra.err != nil {
		return Err[Tuple3[A, B, C]](ra.err)
	}
	if rb.err != nil {
		return Err[Tuple3[A, B, C]](rb.err)
	}
	if rc.err != nil {
		return Err[Tuple3[A, B, C]](rc.err)
	}
	return Ok(Tuple3[A, B, C]{First: ra.value, Second: rb.value, Third: rc.value})
}

// Sequence converts a slice of Results into a Result containing a slice of
// values, failing fast on the first error.
//
// Example:
//
//	res := result.Sequence([]result.Result[int]{loadA(), loadB()})
func Sequence[T any](results []Result[T]) Result[[]T] {
	values := make([]T, 0, len(results))
	for _, r := range results {
		if r.err != nil {
			return Err[[]T](r.err)
		}
		values = append(values, r.value)
	}
	return Ok(values)
}

// Traverse maps input values to Results and sequences them.
//
// Example:
//
//	res := result.Traverse(ids, func(id int) result.Result[User] {
//		return loadUser(id)
//	})
func Traverse[A any, B any](items []A, fn func(A) Result[B]) Result[[]B] {
	values := make([]B, 0, len(items))
	for _, item := range items {
		res := fn(item)
		if res.err != nil {
			return Err[[]B](res.err)
		}
		values = append(values, res.value)
	}
	return Ok(values)
}

// Tuple2 represents a pair of values.
//
// Example:
//
//	p := result.Tuple2[int, string]{First: 1, Second: "a"}
type Tuple2[A any, B any] struct {
	First  A
	Second B
}

// Tuple3 represents three values.
//
// Example:
//
//	t := result.Tuple3[int, string, bool]{First: 1, Second: "a", Third: true}
type Tuple3[A any, B any, C any] struct {
	First  A
	Second B
	Third  C
}

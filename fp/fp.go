// Package fp provides lightweight functional composition helpers for Go.
//
// Example:
//
//	value := fp.Pipe("go",
//		func(s string) string { return strings.ToUpper(s) },
//		func(s string) string { return s + "!" },
//	)
package fp

// Identity returns the supplied value unchanged.
//
// Example:
//
//	value := Identity(42)
func Identity[T any](v T) T {
	return v
}

// Constant returns a function that always returns v.
//
// Example:
//
//	getDefault := Constant(time.Minute)
//	fmt.Println(getDefault())
func Constant[T any](v T) func() T {
	return func() T {
		return v
	}
}

// Pipe applies a sequence of functions to value. All functions must accept and
// return the same type.
//
// Example:
//
//	result := Pipe(2,
//		func(n int) int { return n * 2 },
//		func(n int) int { return n + 1 },
//	)
func Pipe[T any](value T, fns ...func(T) T) T {
	result := value
	for _, fn := range fns {
		result = fn(result)
	}
	return result
}

// Compose composes functions in right-to-left order.
//
// Example:
//
//	fn := Compose(
//		func(n int) int { return n * 2 },
//		func(n int) int { return n + 3 },
//	)
//	value := fn(5)
func Compose[T any](fns ...func(T) T) func(T) T {
	return func(value T) T {
		result := value
		for i := len(fns) - 1; i >= 0; i-- {
			result = fns[i](result)
		}
		return result
	}
}

// Curry converts a binary function into its curried form.
//
// Example:
//
//	add := func(a, b int) int { return a + b }
//	curried := Curry(add)
//	addFive := curried(5)
//	result := addFive(3)
func Curry[A any, B any, C any](fn func(A, B) C) func(A) func(B) C {
	return func(a A) func(B) C {
		return func(b B) C {
			return fn(a, b)
		}
	}
}

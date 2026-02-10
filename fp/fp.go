// Package fp provides lightweight functional composition helpers for Go.
package fp

// Identity returns the supplied value unchanged.
func Identity[T any](v T) T {
	return v
}

// Constant returns a function that always returns v.
func Constant[T any](v T) func() T {
	return func() T {
		return v
	}
}

// Pipe applies a sequence of functions to value. All functions must accept and
// return the same type.
func Pipe[T any](value T, fns ...func(T) T) T {
	result := value
	for _, fn := range fns {
		result = fn(result)
	}
	return result
}

// Compose composes functions in right-to-left order.
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
func Curry[A any, B any, C any](fn func(A, B) C) func(A) func(B) C {
	return func(a A) func(B) C {
		return func(b B) C {
			return fn(a, b)
		}
	}
}

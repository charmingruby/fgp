// Package validated accumulates multiple errors while still returning values.
//
// Use it for input validation, DTO decoding, and config parsing where all
// issues should be reported at once instead of short-circuiting on the first
// failure.
package validated

import (
	"errors"

	"github.com/charmingruby/fgp/result"
)

// Validated wraps either a successful value or a collection of validation errors.
type Validated[E any, T any] struct {
	value  T
	errors []E
}

// Valid constructs a successful Validated value.
func Valid[E any, T any](value T) Validated[E, T] {
	return Validated[E, T]{value: value}
}

// Invalid constructs a failed Validated aggregating the provided errors.
func Invalid[E any, T any](errs ...E) Validated[E, T] {
	if len(errs) == 0 {
		return Validated[E, T]{errors: []E{}}
	}
	copyErrs := make([]E, len(errs))
	copy(copyErrs, errs)
	return Validated[E, T]{errors: copyErrs}
}

// IsValid reports whether the value is valid.
func (v Validated[E, T]) IsValid() bool {
	return len(v.errors) == 0
}

// Errors returns the collected errors. The returned slice is immutable copy.
func (v Validated[E, T]) Errors() []E {
	if len(v.errors) == 0 {
		return []E{}
	}
	copyErrs := make([]E, len(v.errors))
	copy(copyErrs, v.errors)
	return copyErrs
}

// UnsafeValue returns the stored value even when invalid.
func (v Validated[E, T]) UnsafeValue() T {
	return v.value
}

// Map transforms the stored value when valid.
func Map[E any, A any, B any](v Validated[E, A], fn func(A) B) Validated[E, B] {
	if !v.IsValid() {
		return Validated[E, B]{errors: v.errors}
	}
	return Valid[E, B](fn(v.value))
}

// Zip combines two Validated values, accumulating errors from both sides.
func Zip[E any, A any, B any](a Validated[E, A], b Validated[E, B]) Validated[E, result.Tuple2[A, B]] {
	if a.IsValid() && b.IsValid() {
		return Valid[E, result.Tuple2[A, B]](result.Tuple2[A, B]{First: a.value, Second: b.value})
	}
	return Validated[E, result.Tuple2[A, B]]{errors: appendErrors(a.errors, b.errors)}
}

// Sequence collapses a slice of Validated values, returning the first invalid
// state with accumulated errors or a slice of values when all succeeded.
func Sequence[E any, T any](items []Validated[E, T]) Validated[E, []T] {
	if len(items) == 0 {
		return Valid[E, []T]([]T{})
	}
	values := make([]T, 0, len(items))
	var errs []E
	for _, item := range items {
		if item.IsValid() {
			values = append(values, item.value)
			continue
		}
		errs = appendErrors(errs, item.errors)
	}
	if len(errs) > 0 {
		return Validated[E, []T]{errors: errs}
	}
	return Valid[E, []T](values)
}

// Traverse maps the input slice to Validated values and sequences them.
func Traverse[E any, A any, B any](items []A, fn func(A) Validated[E, B]) Validated[E, []B] {
	if len(items) == 0 {
		return Valid[E, []B]([]B{})
	}
	values := make([]B, 0, len(items))
	var errs []E
	for _, item := range items {
		res := fn(item)
		if res.IsValid() {
			values = append(values, res.value)
			continue
		}
		errs = appendErrors(errs, res.errors)
	}
	if len(errs) > 0 {
		return Validated[E, []B]{errors: errs}
	}
	return Valid[E, []B](values)
}

// FromResult lifts a Result into a Validated using error accumulation semantics.
func FromResult[T any](res result.Result[T]) Validated[error, T] {
	if res.IsOk() {
		return Valid[error](res.UnwrapOr(zero[T]()))
	}
	return Invalid[error, T](res.Err())
}

// ToResult converts a Validated of errors into a Result, joining errors when
// the value is invalid.
func ToResult[T any](v Validated[error, T]) result.Result[T] {
	if v.IsValid() {
		return result.Ok(v.value)
	}
	return result.Err[T](errors.Join(v.errors...))
}

func appendErrors[E any](dst []E, src []E) []E {
	if len(src) == 0 {
		return dst
	}
	if len(dst) == 0 {
		dst = make([]E, 0, len(src))
	}
	return append(dst, src...)
}

func zero[T any]() T {
	var z T
	return z
}

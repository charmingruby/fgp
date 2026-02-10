// Package option implements a generic Option type for presence/absence semantics.
package option

import (
	"errors"
	"fmt"

	"github.com/gustavodias/fgp/result"
)

// Option represents presence or absence of a value of type T. The zero value is
// None, so Options can be embedded safely. Values are stored inline (no pointer
// boxing) which makes Some(nil) safe for nil-capable types; use IsSome to
// distinguish between absence and an explicit nil.
type Option[T any] struct {
	value T
	ok    bool
}

// Some constructs an Option that wraps value. Some(nil) is valid when T accepts
// nil values; use IsSome to test for presence explicitly.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, ok: true}
}

// None constructs an empty Option for the provided type.
func None[T any]() Option[T] {
	return Option[T]{ok: false}
}

// FromOk constructs an Option from a value and ok flag, mirroring Go's common
// multi-return patterns (e.g. map lookups).
func FromOk[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// FromPtr creates an Option from a pointer, treating nil as None.
func FromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}
	return Some(*ptr)
}

// IsSome reports true when the Option contains a value (even if that value is
// nil). It is safe to call concurrently when the Option is not being mutated.
func (o Option[T]) IsSome() bool {
	return o.ok
}

// IsNone reports true when the Option is empty.
func (o Option[T]) IsNone() bool {
	return !o.ok
}

// Get returns the contained value along with a boolean indicating whether it
// was present.
func (o Option[T]) Get() (T, bool) {
	return o.value, o.ok
}

// UnsafeGet returns the contained value or panics when the Option is None. It
// should only be used in hot paths where presence is guaranteed.
func (o Option[T]) UnsafeGet() T {
	if !o.ok {
		panic("option: UnsafeGet on None")
	}
	return o.value
}

// GetOrElse returns the contained value when present, otherwise it returns the
// provided fallback value.
func (o Option[T]) GetOrElse(fallback T) T {
	if o.ok {
		return o.value
	}
	return fallback
}

// GetOrElseFunc behaves like GetOrElse but lazily evaluates the fallback only
// when necessary.
func (o Option[T]) GetOrElseFunc(fn func() T) T {
	if o.ok {
		return o.value
	}
	return fn()
}

// OrElse returns the Option itself when it is Some, otherwise returns other.
func (o Option[T]) OrElse(other Option[T]) Option[T] {
	if o.ok {
		return o
	}
	return other
}

// OrElseFunc behaves like OrElse but lazily constructs the replacement when
// necessary.
func (o Option[T]) OrElseFunc(fn func() Option[T]) Option[T] {
	if o.ok {
		return o
	}
	return fn()
}

// ToPtr converts the Option into a pointer, returning nil when None. The
// returned pointer references a copy of the stored value to preserve immutability.
func (o Option[T]) ToPtr() *T {
	if !o.ok {
		return nil
	}
	value := o.value
	return &value
}

// Filter keeps the value when predicate returns true, otherwise it becomes None.
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if o.ok && predicate(o.value) {
		return o
	}
	return None[T]()
}

// Fold collapses the Option into a single value by selecting onNone when the
// Option is empty or applying onSome to the contained value.
func Fold[T any, U any](o Option[T], onNone func() U, onSome func(T) U) U {
	if o.ok {
		return onSome(o.value)
	}
	return onNone()
}

// Map transforms the contained value with fn when present, returning a new
// Option of type U.
func Map[T any, U any](o Option[T], fn func(T) U) Option[U] {
	if o.ok {
		return Some(fn(o.value))
	}
	return None[U]()
}

// FlatMap chains the Option with another Option-valued function.
func FlatMap[T any, U any](o Option[T], fn func(T) Option[U]) Option[U] {
	if o.ok {
		return fn(o.value)
	}
	return None[U]()
}

// ToResult converts an Option into a Result, producing errFactory() when the
// Option is None. If errFactory returns nil the function wraps a descriptive
// error to avoid silent failures.
func (o Option[T]) ToResult(errFactory func() error) result.Result[T] {
	if o.ok {
		return result.Ok(o.value)
	}
	var err error
	if errFactory != nil {
		err = errFactory()
	}
	if err == nil {
		err = errors.New("option: missing value")
	}
	return result.Err[T](err)
}

// String implements fmt.Stringer for debugging. It is not intended for
// serialization and keeps implementation reflection-free.
func (o Option[T]) String() string {
	if o.ok {
		return fmt.Sprintf("Some(%v)", o.value)
	}
	return "None"
}

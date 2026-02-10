// Package option implements a generic Option type for presence/absence semantics.
//
// Example:
//
//	opt := option.Some("token")
//	value := opt.GetOrElse("missing")
//
// Option's Map/FlatMap/Traverse helpers obey the Functor and Monad laws validated
// in laws_option_test.go so they compose predictably in production pipelines.
package option

import (
	"errors"
	"fmt"

	"github.com/charmingruby/fgp/result"
)

// Option represents presence or absence of a value of type T. The zero value is
// None, so Options can be embedded safely. Values are stored inline (no pointer
// boxing) which makes Some(nil) safe for nil-capable types; use IsSome to
// distinguish between absence and an explicit nil.
//
// Example:
//
//	opt := Some("config")
//	value, ok := opt.Get()
//	if !ok {
//		value = "default"
//	}
//	_ = value // value == "config"
type Option[T any] struct {
	value T
	ok    bool
}

// Some constructs an Option that wraps value. Some(nil) is valid when T accepts
// nil values; use IsSome to test for presence explicitly.
//
// Example:
//
//	userID := Some(42)
//	if userID.IsSome() {
//		fmt.Println("have id", userID.UnsafeGet())
//	}
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, ok: true}
}

// None constructs an empty Option for the provided type.
//
// Example:
//
//	cache := None[string]()
//	value := cache.GetOrElse("fallback")
//	// value == "fallback"
func None[T any]() Option[T] {
	return Option[T]{ok: false}
}

// FromOk constructs an Option from a value and ok flag, mirroring Go's common
// multi-return patterns (e.g. map lookups).
//
// Example:
//
//	value, ok := headers["Authorization"]
//	token := FromOk(value, ok)
//	if token.IsNone() {
//		return errors.New("missing auth header")
//	}
func FromOk[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// FromPtr creates an Option from a pointer, treating nil as None.
//
// Example:
//
//	var cfg *Config
//	opt := FromPtr(cfg)
//	if opt.IsNone() {
//		return errors.New("config not loaded")
//	}
func FromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}
	return Some(*ptr)
}

// IsSome reports true when the Option contains a value (even if that value is
// nil). It is safe to call concurrently when the Option is not being mutated.
//
// Example:
//
//	if opt.IsSome() {
//		process(opt.UnsafeGet())
//	}
func (o Option[T]) IsSome() bool {
	return o.ok
}

// IsNone reports true when the Option is empty.
//
// Example:
//
//	if profile.IsNone() {
//		return errors.New("profile missing")
//	}
func (o Option[T]) IsNone() bool {
	return !o.ok
}

// Get returns the contained value along with a boolean indicating whether it
// was present.
//
// Example:
//
//	value, ok := opt.Get()
//	if !ok {
//		value = fallback
//	}
func (o Option[T]) Get() (T, bool) {
	return o.value, o.ok
}

// UnsafeGet returns the contained value or panics when the Option is None. It
// should only be used in hot paths where presence is guaranteed.
//
// Example:
//
//	func mustUserID(opt Option[int]) int {
//		return opt.UnsafeGet()
//	}
func (o Option[T]) UnsafeGet() T {
	if !o.ok {
		panic("option: UnsafeGet on None")
	}
	return o.value
}

// GetOrElse returns the contained value when present, otherwise it returns the
// provided fallback value.
//
// Example:
//
//	name := opt.GetOrElse("anonymous")
func (o Option[T]) GetOrElse(fallback T) T {
	if o.ok {
		return o.value
	}
	return fallback
}

// GetOrElseFunc behaves like GetOrElse but lazily evaluates the fallback only
// when necessary.
//
// Example:
//
//	data := opt.GetOrElseFunc(func() string {
//		return loadExpensiveDefault()
//	})
func (o Option[T]) GetOrElseFunc(fn func() T) T {
	if o.ok {
		return o.value
	}
	return fn()
}

// OrElse returns the Option itself when it is Some, otherwise returns other.
//
// Example:
//
//	primary := lookupUser()
//	fallback := option.Some(defaultUser)
//	user := primary.OrElse(fallback)
func (o Option[T]) OrElse(other Option[T]) Option[T] {
	if o.ok {
		return o
	}
	return other
}

// OrElseFunc behaves like OrElse but lazily constructs the replacement when
// necessary.
//
// Example:
//
//	config := loadFromCache().OrElseFunc(func() Option[Config] {
//		return loadFromDisk()
//	})
func (o Option[T]) OrElseFunc(fn func() Option[T]) Option[T] {
	if o.ok {
		return o
	}
	return fn()
}

// ToPtr converts the Option into a pointer, returning nil when None. The
// returned pointer references a copy of the stored value to preserve immutability.
//
// Example:
//
//	ptr := opt.ToPtr()
//	if ptr == nil {
//		return errors.New("missing value")
//	}
func (o Option[T]) ToPtr() *T {
	if !o.ok {
		return nil
	}
	value := o.value
	return &value
}

// Filter keeps the value when predicate returns true, otherwise it becomes None.
//
// Example:
//
//	adult := opt.Filter(func(u User) bool { return u.Age >= 18 })
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if o.ok && predicate(o.value) {
		return o
	}
	return None[T]()
}

// Fold collapses the Option into a single value by selecting onNone when the
// Option is empty or applying onSome to the contained value.
//
// Example:
//
//	greeting := Fold(opt,
//		func() string { return "guest" },
//		func(name string) string { return "hello " + name },
//	)
func Fold[T any, U any](o Option[T], onNone func() U, onSome func(T) U) U {
	if o.ok {
		return onSome(o.value)
	}
	return onNone()
}

// Map transforms the contained value with fn when present, returning a new
// Option of type U.
//
// Example:
//
//	lenOpt := Map(opt, func(s string) int { return len(s) })
func Map[T any, U any](o Option[T], fn func(T) U) Option[U] {
	if o.ok {
		return Some(fn(o.value))
	}
	return None[U]()
}

// FlatMap chains the Option with another Option-valued function.
//
// Example:
//
//	user := FlatMap(sessionOpt, loadUserBySession)
func FlatMap[T any, U any](o Option[T], fn func(T) Option[U]) Option[U] {
	if o.ok {
		return fn(o.value)
	}
	return None[U]()
}

// Tap executes fn when the Option is Some and always returns the original Option.
//
// Example:
//
//	opt := Tap(option.Some(user), func(u User) { metrics.Count("user_loaded") })
func Tap[T any](o Option[T], fn func(T)) Option[T] {
	if o.ok {
		fn(o.value)
	}
	return o
}

// Pair combines two related values.
//
// Example:
//
//	p := Pair[int, string]{First: 1, Second: "a"}
type Pair[A any, B any] struct {
	First  A
	Second B
}

// Zip combines two Options into one that is Some when both inputs are Some.
//
// Example:
//
//	combined := Zip(firstNameOpt, lastNameOpt)
func Zip[A any, B any](a Option[A], b Option[B]) Option[Pair[A, B]] {
	if a.ok && b.ok {
		return Some(Pair[A, B]{First: a.value, Second: b.value})
	}
	return None[Pair[A, B]]()
}

// Traverse maps items to Options using fn and collapses them into an Option of
// collected values. It short-circuits on the first None.
//
// Example:
//
//	users := Traverse(ids, func(id int) option.Option[User] { ... })
func Traverse[A any, B any](items []A, fn func(A) Option[B]) Option[[]B] {
	if len(items) == 0 {
		return Some([]B{})
	}
	values := make([]B, 0, len(items))
	for _, item := range items {
		res := fn(item)
		if !res.ok {
			return None[[]B]()
		}
		values = append(values, res.value)
	}
	return Some(values)
}

// Sequence converts a slice of Options into an Option containing all values when
// every element is Some. It fails fast on the first None.
//
// Example:
//
//	combined := Sequence([]option.Option[int]{option.Some(1), option.Some(2)})
func Sequence[T any](items []Option[T]) Option[[]T] {
	if len(items) == 0 {
		return Some([]T{})
	}
	values := make([]T, 0, len(items))
	for _, item := range items {
		if !item.ok {
			return None[[]T]()
		}
		values = append(values, item.value)
	}
	return Some(values)
}

// ToResult converts an Option into a Result, producing errFactory() when the
// Option is None. If errFactory returns nil the function wraps a descriptive
// error to avoid silent failures.
//
// Example:
//
//	res := opt.ToResult(func() error { return errors.New("missing user") })
//	value, err := res.Unwrap()
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
//
// Example:
//
//	fmt.Println(option.Some(3))  // Some(3)
//	fmt.Println(option.None[int]()) // None
func (o Option[T]) String() string {
	if o.ok {
		return fmt.Sprintf("Some(%v)", o.value)
	}
	return "None"
}

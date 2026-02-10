// Package seq offers eager and lazy functional helpers for Go slices.
//
// Example:
//
//	values := seq.Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
package seq

// Iterator is a lazy, pull-based iterator.
//
// Example:
//
//	it := Iterator[int]{next: func() (int, bool) { return 0, false }}
type Iterator[T any] struct {
	next func() (T, bool)
}

// Next yields the next value. When ok is false, iteration is complete.
//
// Example:
//
//	value, ok := it.Next()
//	if !ok {
//		return
//	}
func (it Iterator[T]) Next() (T, bool) {
	if it.next == nil {
		var zero T
		return zero, false
	}
	return it.next()
}

// FromSlice creates an iterator over the provided slice without copying.
//
// Example:
//
//	it := FromSlice([]int{1, 2})
//	val, _ := it.Next()
func FromSlice[T any](values []T) Iterator[T] {
	idx := 0
	return Iterator[T]{
		next: func() (T, bool) {
			if idx >= len(values) {
				var zero T
				return zero, false
			}
			v := values[idx]
			idx++
			return v, true
		},
	}
}

// MapIter lazily transforms iterator values.
//
// Example:
//
//	squared := MapIter(it, func(n int) int { return n * n })
func MapIter[A any, B any](it Iterator[A], fn func(A) B) Iterator[B] {
	return Iterator[B]{
		next: func() (B, bool) {
			v, ok := it.Next()
			if !ok {
				var zero B
				return zero, false
			}
			return fn(v), true
		},
	}
}

// FilterIter keeps values satisfying predicate.
//
// Example:
//
//	even := FilterIter(it, func(n int) bool { return n%2 == 0 })
func FilterIter[T any](it Iterator[T], predicate func(T) bool) Iterator[T] {
	return Iterator[T]{
		next: func() (T, bool) {
			for {
				v, ok := it.Next()
				if !ok {
					var zero T
					return zero, false
				}
				if predicate(v) {
					return v, true
				}
			}
		},
	}
}

// Take returns an iterator that yields at most n elements.
//
// Example:
//
//	firstTwo := Take(it, 2)
func Take[T any](it Iterator[T], n int) Iterator[T] {
	if n <= 0 {
		return Iterator[T]{}
	}
	count := 0
	return Iterator[T]{
		next: func() (T, bool) {
			if count >= n {
				var zero T
				return zero, false
			}
			v, ok := it.Next()
			if !ok {
				var zero T
				return zero, false
			}
			count++
			return v, true
		},
	}
}

// Drop skips the first n elements.
//
// Example:
//
//	skipPrefix := Drop(it, 5)
func Drop[T any](it Iterator[T], n int) Iterator[T] {
	if n <= 0 {
		return it
	}
	skipped := false
	return Iterator[T]{
		next: func() (T, bool) {
			if !skipped {
				for range n {
					if _, ok := it.Next(); !ok {
						var zero T
						return zero, false
					}
				}
				skipped = true
			}
			return it.Next()
		},
	}
}

// Range constructs an iterator that yields integers from start (inclusive) to
// end (exclusive). When start >= end the iterator is empty.
//
// Example:
//
//	it := Range(0, 3) // yields 0,1,2
func Range(start, end int) Iterator[int] {
	if start >= end {
		return Iterator[int]{}
	}
	current := start
	return Iterator[int]{
		next: func() (int, bool) {
			if current >= end {
				return 0, false
			}
			value := current
			current++
			return value, true
		},
	}
}

// Repeat creates an infinite iterator repeating value. Consumers should limit
// it with Take/TakeWhile to avoid unbounded loops.
//
// Example:
//
//	it := Repeat("retry")
func Repeat[T any](value T) Iterator[T] {
	return Iterator[T]{
		next: func() (T, bool) {
			return value, true
		},
	}
}

// Iterate repeatedly applies fn to state starting from seed.
//
// Example:
//
//	it := Iterate(1, func(n int) int { return n * 2 })
func Iterate[T any](seed T, fn func(T) T) Iterator[T] {
	state := seed
	return Iterator[T]{
		next: func() (T, bool) {
			value := state
			state = fn(state)
			return value, true
		},
	}
}

// TakeWhile yields elements while predicate returns true and stops immediately
// once predicate fails.
//
// Example:
//
//	pluck := TakeWhile(it, func(v int) bool { return v < 10 })
func TakeWhile[T any](it Iterator[T], predicate func(T) bool) Iterator[T] {
	if predicate == nil {
		return Iterator[T]{}
	}
	stopped := false
	return Iterator[T]{
		next: func() (T, bool) {
			if stopped {
				var zero T
				return zero, false
			}
			value, ok := it.Next()
			if !ok {
				var zero T
				return zero, false
			}
			if !predicate(value) {
				stopped = true
				var zero T
				return zero, false
			}
			return value, true
		},
	}
}

// DropWhile skips elements until predicate returns false, then yields all
// remaining values including the first that failed predicate.
//
// Example:
//
//	trimmed := DropWhile(it, func(v int) bool { return v == 0 })
func DropWhile[T any](it Iterator[T], predicate func(T) bool) Iterator[T] {
	skipped := false
	return Iterator[T]{
		next: func() (T, bool) {
			if !skipped {
				for {
					value, ok := it.Next()
					if !ok {
						var zero T
						return zero, false
					}
					if predicate == nil || !predicate(value) {
						skipped = true
						return value, true
					}
				}
			}
			return it.Next()
		},
	}
}

// ToSlice exhausts the iterator and collects its values.
//
// Example:
//
//	values := ToSlice(it)
func ToSlice[T any](it Iterator[T]) []T {
	var result []T
	for {
		v, ok := it.Next()
		if !ok {
			break
		}
		result = append(result, v)
	}
	if result == nil {
		return []T{}
	}
	return result
}

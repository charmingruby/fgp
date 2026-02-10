// Package seq offers eager and lazy functional helpers for Go slices.
package seq

// Iterator is a lazy, pull-based iterator.
type Iterator[T any] struct {
	next func() (T, bool)
}

// Next yields the next value. When ok is false, iteration is complete.
func (it Iterator[T]) Next() (T, bool) {
	if it.next == nil {
		var zero T
		return zero, false
	}
	return it.next()
}

// FromSlice creates an iterator over the provided slice without copying.
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

// ToSlice exhausts the iterator and collects its values.
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

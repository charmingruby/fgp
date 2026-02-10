package seq

// Map transforms each element using fn and returns a new slice with the same
// length as input.
func Map[A any, B any](in []A, fn func(A) B) []B {
	if len(in) == 0 {
		return []B{}
	}
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

// Filter keeps values satisfying predicate. The returned slice shares no
// backing array with the input to preserve immutability.
func Filter[T any](in []T, predicate func(T) bool) []T {
	if len(in) == 0 {
		return []T{}
	}
	result := make([]T, 0, len(in))
	for _, v := range in {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// FlatMap applies fn to each element and concatenates the resulting slices.
func FlatMap[A any, B any](in []A, fn func(A) []B) []B {
	if len(in) == 0 {
		return []B{}
	}
	// Estimate capacity by summing lengths lazily to avoid unnecessary passes.
	var out []B
	for _, v := range in {
		chunk := fn(v)
		if len(chunk) == 0 {
			continue
		}
		out = append(out, chunk...)
	}
	if out == nil {
		return []B{}
	}
	return out
}

// FoldLeft reduces the slice from left to right using the provided accumulator.
func FoldLeft[A any, B any](in []A, init B, fn func(B, A) B) B {
	acc := init
	for _, v := range in {
		acc = fn(acc, v)
	}
	return acc
}

// Reduce applies fn across elements, returning false when slice empty.
func Reduce[T any](in []T, fn func(T, T) T) (T, bool) {
	if len(in) == 0 {
		var zero T
		return zero, false
	}
	acc := in[0]
	for i := 1; i < len(in); i++ {
		acc = fn(acc, in[i])
	}
	return acc, true
}

// Find returns the first element satisfying predicate.
func Find[T any](in []T, predicate func(T) bool) (T, bool) {
	for _, v := range in {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Any reports whether any element satisfies predicate.
func Any[T any](in []T, predicate func(T) bool) bool {
	for _, v := range in {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All reports whether all elements satisfy predicate.
func All[T any](in []T, predicate func(T) bool) bool {
	for _, v := range in {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// GroupBy groups elements by the key returned from keySelector.
func GroupBy[T any, K comparable](in []T, keySelector func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, v := range in {
		key := keySelector(v)
		groups[key] = append(groups[key], v)
	}
	return groups
}

// DistinctBy removes duplicates determined by keySelector, preserving order.
func DistinctBy[T any, K comparable](in []T, keySelector func(T) K) []T {
	if len(in) == 0 {
		return []T{}
	}
	seen := make(map[K]struct{}, len(in))
	result := make([]T, 0, len(in))
	for _, v := range in {
		key := keySelector(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, v)
	}
	return result
}

// Partition splits the slice into two slices based on predicate outcome.
func Partition[T any](in []T, predicate func(T) bool) ([]T, []T) {
	if len(in) == 0 {
		return []T{}, []T{}
	}
	matches := make([]T, 0, len(in))
	rest := make([]T, 0, len(in))
	for _, v := range in {
		if predicate(v) {
			matches = append(matches, v)
		} else {
			rest = append(rest, v)
		}
	}
	return matches, rest
}

// Zip combines two slices into a slice of pairs up to the shortest length.
func Zip[A any, B any](a []A, b []B) []Pair[A, B] {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	result := make([]Pair[A, B], limit)
	for i := range limit {
		result[i] = Pair[A, B]{First: a[i], Second: b[i]}
	}
	return result
}

// Pair represents two related values.
type Pair[A any, B any] struct {
	First  A
	Second B
}

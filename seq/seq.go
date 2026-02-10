package seq

// Map transforms each element using fn and returns a new slice with the same
// length as input.
//
// Example:
//
//	doubled := Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
//	// doubled == []int{2, 4, 6}
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
//
// Example:
//
//	eve := Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
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
//
// Example:
//
//	letters := FlatMap([]string{"ab", "cd"}, func(s string) []string {
//		return strings.Split(s, "")
//	})
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
//
// Example:
//
//	sum := FoldLeft([]int{1, 2, 3}, 0, func(acc, n int) int {
//		return acc + n
//	})
func FoldLeft[A any, B any](in []A, init B, fn func(B, A) B) B {
	acc := init
	for _, v := range in {
		acc = fn(acc, v)
	}
	return acc
}

// Reduce applies fn across elements, returning false when slice empty.
//
// Example:
//
//	max, ok := Reduce([]int{5, 4, 9}, func(a, b int) int {
//		if a > b {
//			return a
//		}
//		return b
//	})
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
//
// Example:
//
//	value, ok := Find(users, func(u User) bool { return u.ID == id })
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
//
// Example:
//
//	hasAdmin := Any(users, func(u User) bool { return u.Role == "admin" })
func Any[T any](in []T, predicate func(T) bool) bool {
	for _, v := range in {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All reports whether all elements satisfy predicate.
//
// Example:
//
//	allPositive := All(nums, func(n int) bool { return n > 0 })
func All[T any](in []T, predicate func(T) bool) bool {
	for _, v := range in {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// GroupBy groups elements by the key returned from keySelector.
//
// Example:
//
//	byStatus := GroupBy(tasks, func(t Task) string { return t.Status })
func GroupBy[T any, K comparable](in []T, keySelector func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, v := range in {
		key := keySelector(v)
		groups[key] = append(groups[key], v)
	}
	return groups
}

// DistinctBy removes duplicates determined by keySelector, preserving order.
//
// Example:
//
//	unique := DistinctBy(users, func(u User) int { return u.ID })
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
//
// Example:
//
//	valid, invalid := Partition(users, func(u User) bool {
//		return u.Active
//	})
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
//
// Example:
//
//	pairs := Zip([]string{"a", "b"}, []int{1, 2})
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

// Chunk splits the slice into consecutive sub-slices of size chunkSize. The
// last chunk may be smaller. Each chunk is copied to preserve immutability.
//
// Example:
//
//	chunks := Chunk([]int{1,2,3,4}, 2) // [[1,2],[3,4]]
func Chunk[T any](in []T, chunkSize int) [][]T {
	if chunkSize <= 0 || len(in) == 0 {
		return [][]T{}
	}
	chunks := make([][]T, 0, (len(in)+chunkSize-1)/chunkSize)
	for i := 0; i < len(in); i += chunkSize {
		end := i + chunkSize
		if end > len(in) {
			end = len(in)
		}
		chunk := make([]T, end-i)
		copy(chunk, in[i:end])
		chunks = append(chunks, chunk)
	}
	return chunks
}

// Window returns a sliding window of size windowSize across the slice. Each
// window is copied to avoid sharing memory with input.
//
// Example:
//
//	windows := Window([]int{1,2,3}, 2) // [[1,2],[2,3]]
func Window[T any](in []T, windowSize int) [][]T {
	if windowSize <= 0 || len(in) == 0 || windowSize > len(in) {
		return [][]T{}
	}
	windows := make([][]T, 0, len(in)-windowSize+1)
	for i := 0; i <= len(in)-windowSize; i++ {
		window := make([]T, windowSize)
		copy(window, in[i:i+windowSize])
		windows = append(windows, window)
	}
	return windows
}

// ScanLeft returns the running accumulation values, including the initial seed
// as the first element of the returned slice.
//
// Example:
//
//	sums := ScanLeft([]int{1,2,3}, 0, func(acc, v int) int { return acc + v })
//	// sums == []int{0, 1, 3, 6}
func ScanLeft[A any, B any](in []A, init B, fn func(B, A) B) []B {
	result := make([]B, len(in)+1)
	result[0] = init
	acc := init
	for i, v := range in {
		acc = fn(acc, v)
		result[i+1] = acc
	}
	return result
}

// Collect fuses filter + map by executing fn for each element and appending the
// produced value when ok is true.
//
// Example:
//
//	evenSquares := Collect([]int{1,2,3}, func(v int) (int, bool) {
//		if v%2 != 0 {
//			return 0, false
//		}
//		return v * v, true
//	})
func Collect[A any, B any](in []A, fn func(A) (B, bool)) []B {
	if len(in) == 0 {
		return []B{}
	}
	out := make([]B, 0, len(in))
	for _, v := range in {
		mapped, ok := fn(v)
		if !ok {
			continue
		}
		out = append(out, mapped)
	}
	return out
}

// Pair represents two related values.
//
// Example:
//
//	p := Pair[string, int]{First: "a", Second: 1}
type Pair[A any, B any] struct {
	First  A
	Second B
}

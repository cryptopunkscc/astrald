package sig

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
	~float32 | ~float64 |
	~string
}

// SliceDiff returns two slices: one containing elements found only in a,
// and one containing elements found only in b. Preserves order and multiplicity.
func SliceDiff[T comparable](a, b []T) (onlyA []T, onlyB []T) {
	// create lookup sets
	setA := make(map[T]struct{})
	setB := make(map[T]struct{})

	for _, val := range a {
		setA[val] = struct{}{}
	}
	for _, val := range b {
		setB[val] = struct{}{}
	}

	// elements only in a
	onlyA = make([]T, 0)
	for _, val := range a {
		if _, exists := setB[val]; !exists {
			onlyA = append(onlyA, val)
		}
	}

	// elements only in b
	onlyB = make([]T, 0)
	for _, val := range b {
		if _, exists := setA[val]; !exists {
			onlyB = append(onlyB, val)
		}
	}

	return onlyA, onlyB
}

// SliceDiffOrdered does the same thing as SliceDiff but is optimized for ordered slices.
func SliceDiffOrdered[T Ordered](a, b []T) ([]T, []T) {
	var onlyInA, onlyInB []T
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			i++
			j++
		case a[i] < b[j]:
			// 'a[i]' is not in 'b'
			onlyInA = append(onlyInA, a[i])
			i++
		case a[i] > b[j]:
			// 'b[j]' is not in 'a'
			onlyInB = append(onlyInB, b[j])
			j++
		}
	}

	// append any remaining elements from a
	for i < len(a) {
		onlyInA = append(onlyInA, a[i])
		i++
	}

	// append any remaining elements from b
	for j < len(b) {
		onlyInB = append(onlyInB, b[j])
		j++
	}

	return onlyInA, onlyInB
}

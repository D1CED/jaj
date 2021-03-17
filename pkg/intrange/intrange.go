package intrange

// IntRange stores a max and min amount for range
type IntRange struct {
	min int
	max int
}

func New(min, max int) IntRange {
	return IntRange{min, max}
}

// IntRanges is a slice of IntRange
type IntRanges []IntRange

// Get returns true if the argument n is included in the closed range
// between min and max
func (r IntRange) Get(n int) bool {
	return n >= r.min && n <= r.max
}

// Get returns true if the argument n is included in the closed range
// between min and max of any of the provided IntRanges
func (rs IntRanges) Get(n int) bool {
	for _, r := range rs {
		if r.Get(n) {
			return true
		}
	}

	return false
}

// Min returns min value between a and b
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns max value between a and b
func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

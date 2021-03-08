package stringset

// Add adds a new value to the Map.
// If n is already in the map, then v is appended to the StringSet under that key.
// Otherwise a new StringSet is creayed containing v
func Add(mss map[string]StringSet, n, v string) {
	_, ok := mss[n]
	if !ok {
		mss[n] = Make()
	}
	mss[n].Set(v)
}

// StringSet is a basic set implementation for strings.
// This is used a lot so it deserves its own type.
// Other types of sets are used throughout the code but do not have
// their own typedef.
// String sets and <type>sets should be used throughout the code when applicable,
// they are a lot more flexible than slices and provide easy lookup.
type StringSet struct{ val map[string]struct{} }

// Set sets key in StringSet.
func (set StringSet) Set(v string) {
	set.val[v] = struct{}{}
}

// Extend sets multiple keys in StringSet.
func (set StringSet) Extend(s ...string) {
	for _, v := range s {
		set.val[v] = struct{}{}
	}
}

// Get returns true if the key exists in the set.
func (set StringSet) Get(v string) bool {
	_, exists := set.val[v]
	return exists
}

// Remove deletes a key from the set.
func (set StringSet) Remove(v string) {
	delete(set.val, v)
}

// ToSlice turns all keys into a string slice.
func (set StringSet) ToSlice() []string {
	slice := make([]string, 0, len(set.val))

	for v := range set.val {
		slice = append(slice, v)
	}

	return slice
}

// Copy copies a StringSet into a new structure of the same type.
func (set StringSet) Copy() StringSet {
	newSet := StringSet{make(map[string]struct{}, len(set.val))}

	for str := range set.val {
		newSet.Set(str)
	}

	return newSet
}

func (set StringSet) Len() int {
	return len(set.val)
}

func (set StringSet) Iter() map[string]struct{} {
	return set.val
}

// Make creates a new StringSet from a set of arguments
func Make(in ...string) StringSet {
	set := StringSet{make(map[string]struct{}, len(in))}

	for _, v := range in {
		set.Set(v)
	}

	return set
}

// Equal compares if two StringSets have the same values
func Equal(a, b StringSet) bool {
	if a.val == nil && b.val == nil {
		return true
	}

	if len(a.val) != len(b.val) {
		return false
	}

	for n := range a.val {
		if !b.Get(n) {
			return false
		}
	}

	return true
}

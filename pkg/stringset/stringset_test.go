package stringset_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sset "github.com/Jguer/yay/v10/pkg/stringset"
)

func TestStringSet(t *testing.T) {

	set := sset.Make("abc", "def")

	assert.True(t, set.Get("abc"))
	assert.False(t, set.Get("xyz"))

	set2 := sset.FromSlice([]string{"abc", "def"})

	assert.True(t, set2.Get("abc"))
	assert.False(t, set2.Get("xyz"))

	assert.True(t, sset.Equal(set, set2))

	set3 := set2.Copy()
	set3.Remove("abc")
	set3.Extend("ghi", "jkl")
	set3.Set("ghi")

	assert.True(t, set2.Get("abc"))
	assert.False(t, set2.Get("ghi"))
	assert.False(t, set2.Get("xyz"))
	assert.True(t, set3.Get("ghi"))
	assert.True(t, set3.Get("jkl"))
	assert.False(t, set3.Get("xyz"))

	assert.ElementsMatch(t, []string{"def", "ghi", "jkl"}, set3.ToSlice())
}

func TestMapStringSet_Add(t *testing.T) {

	m := sset.MapStringSet{
		"key1": sset.Make("abc"),
	}

	m.Add("key1", "def")
	m.Add("key2", "ghi")

	assert.True(t, m["key1"].Get("abc"))
	assert.True(t, m["key1"].Get("def"))
	assert.True(t, m["key2"].Get("ghi"))
}

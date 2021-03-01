package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jguer/yay/v10/pkg/settings/parser"
)

func TestParse(t *testing.T) {

	const (
		A parser.Enum = iota
		B
		C
		ABC
		DEF
	)

	fn := func(s string) (parser.Enum, bool) {
		switch s {
		case "a":
			return A, false
		case "b":
			return B, true
		case "c":
			return C, false
		case "abc":
			return ABC, false
		case "def":
			return DEF, true
		default:
			return parser.InvalidOption, false
		}
	}

	const cmdLine = "my-program -a target5 --abc -abc -bvalue1 target4 -b value2 --def value1 target2 --def=value2 target1 target3"

	a, err := parser.Parse(fn, strings.Split(cmdLine, " ")[1:], nil)

	assert.NoError(t, err)
	assert.NotNil(t, a)

	t.Run("targets", func(t *testing.T) {
		assert.Equal(t, []string{"target5", "target4", "target2", "target1", "target3"}, a.Targets())
	})

	t.Run("correct", func(t *testing.T) {
		var ctr = 0

		a.Iterate(func(e parser.Enum, ss []string) bool {
			switch e {
			case A:
				ctr++
			case B:
				assert.Equal(t, []string{"c", "value1", "value2"}, ss)
				ctr++
			case C:
				t.Errorf("unreachable %v", e)
			case ABC:
				ctr++
			case DEF:
				assert.Equal(t, []string{"value1", "value2"}, ss)
				ctr++
			default:
				t.Errorf("unknown argument %v", e)
			}
			return true
		})

		assert.Equal(t, 4, ctr)
	})

	t.Run("accessors", func(t *testing.T) {
		assert.True(t, a.Exists(A))
		assert.Equal(t, []string{"c", "value1", "value2"}, a.Get(B))
		assert.False(t, a.Exists(C))
		assert.True(t, a.Exists(ABC), "abc not set")
		assert.Equal(t, []string{"value1", "value2"}, a.Get(DEF))

		val, _ := a.Last(DEF)
		assert.Equal(t, "value2", val)
	})

	t.Run("count", func(t *testing.T) {
		assert.Equal(t, 2, a.Count(A))
		assert.Equal(t, 0, a.Count(C))
		assert.Equal(t, 1, a.Count(ABC))
	})

	t.Run("breaks", func(t *testing.T) {
		iter := 0
		a.Iterate(func(_ parser.Enum, _ []string) bool {
			iter++
			return false
		})
		assert.Equal(t, iter, 1)
	})

	t.Run("strict_order", func(t *testing.T) {
		iter := 0
		expectIter := []parser.Enum{A, ABC, B, DEF}

		a.Iterate(func(e parser.Enum, _ []string) bool {
			assert.Equal(t, expectIter[iter], e)
			iter++
			return true
		})
		assert.Equal(t, iter, 4)
	})

	t.Run("erronus", func(t *testing.T) {
		var err error

		// No arguments
		_, err = parser.Parse(fn, strings.Split("-b", " "), nil)
		assert.Error(t, err)

		_, err = parser.Parse(fn, strings.Split("--def", " "), nil)
		assert.Error(t, err)

		// Invalid options
		_, err = parser.Parse(fn, strings.Split("-acd", " "), nil)
		assert.Error(t, err)

		_, err = parser.Parse(fn, strings.Split("--acd", " "), nil)
		assert.Error(t, err)
	})
}

func TestSpec(t *testing.T) {
	var (
		myBool   bool
		myString string
		myList   []string
		myAll    []string
	)

	var spec parser.Spec
	spec.Specify("a all", false, func(_ []string) { myBool = true })
	spec.Specify("t test", true, func(ss []string) { myString, _ = parser.GetLast(ss) })
	spec.Targets(func(ss []string) { myList = ss })

	sub := spec.Sub("s subcommand", func() {})
	sub.Specify("a all", true, func(ss []string) { myAll = ss })

	err := spec.Parse(strings.Split("-a -s -a lol -a kek", " "), nil)
	require.NoError(t, err)

	assert.Equal(t, true, myBool)
	assert.Equal(t, "", myString)
	assert.Equal(t, []string(nil), myList)
	assert.Equal(t, []string{"lol", "kek"}, myAll)
}

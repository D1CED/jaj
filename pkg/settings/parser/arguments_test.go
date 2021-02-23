package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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
			return parser.InvalidFlag, false
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

func TestEnumerate(t *testing.T) {

	t.Run("short", func(t *testing.T) {
		my, alias := parser.Enumerate("ab:c", false)

		p, err := parser.Parse(alias, strings.Split("-a -bc -b x target1", " "), nil)
		assert.NoError(t, err)

		assert.True(t, p.Exists(my("a")))
		assert.False(t, p.Exists(my("c")))
		assert.Equal(t, []string{"c", "x"}, p.Get(my("b")))
		assert.Equal(t, []string{"target1"}, p.Targets())
	})

	t.Run("short+long", func(t *testing.T) {
		my, alias := parser.Enumerate("a(foo)b(bar):c(baz)", false)

		p, err := parser.Parse(alias, strings.Split("--foo --bar=c -b x target1", " "), nil)
		assert.NoError(t, err)

		assert.True(t, p.Exists(my("foo")))
		assert.False(t, p.Exists(my("c")))
		assert.Equal(t, []string{"c", "x"}, p.Get(my("b")))
		assert.Equal(t, []string{"target1"}, p.Targets())

		func() {
			defer func() { r := recover(); assert.NotNil(t, r) }()

			my("x")
		}()
	})

	t.Run("long", func(t *testing.T) {
		my, alias := parser.Enumerate("[foo][bar]:[baz]", false)

		p, err := parser.Parse(alias, strings.Split("--foo --bar=c --bar x target1", " "), nil)
		if !assert.NoError(t, err) {
			return
		}

		assert.True(t, p.Exists(my("foo")))
		assert.False(t, p.Exists(my("baz")))
		assert.Equal(t, []string{"c", "x"}, p.Get(my("bar")))
		assert.Equal(t, []string{"target1"}, p.Targets())
	})

	t.Run("multline", func(t *testing.T) {
		const desc = `
			a:bc
			d(def):
			g(gir)
			[foo]
			[bar]:
			[baz-zing]
		`
		my, alias := parser.Enumerate(desc, false)

		p, err := parser.Parse(alias, strings.Split("-ahello --baz-zing --bar=test --def target2 -gcb", " "), nil)
		assert.NoError(t, err)

		assert.True(t, p.Exists(my("baz-zing")))
		assert.True(t, p.Exists(my("def")))
		assert.True(t, p.Exists(my("c")))
		assert.True(t, p.Exists(my("b")))
		assert.True(t, p.Exists(my("gir")))
		assert.False(t, p.Exists(my("foo")))

		assert.Equal(t, []string{"hello"}, p.Get(my("a")))
		assert.Equal(t, []string{"target2"}, p.Get(my("d")))
		assert.Equal(t, []string{"test"}, p.Get(my("bar")))
	})

	t.Run("duplicate", func(t *testing.T) {
		defer func() {
			r := recover()
			t.Log(r)
			assert.NotNil(t, r)
		}()

		parser.Enumerate("[f]g(f):[baz]", false)
	})

	t.Run("doc", func(t *testing.T) {
		t.Skip("currently go test does not handle exit(0) as error")
		const desc = `
			a:bc
			d(def):
			g(gir)
			[foo]
			[bar]:
			[baz-zing]
		`
		_, alias := parser.Enumerate(desc, true)
		parser.Parse(alias, []string{"-h"}, nil)
		t.Fail()
	})
}

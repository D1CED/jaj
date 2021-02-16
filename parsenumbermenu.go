package main

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/Jguer/yay/v10/pkg/intrange"
	"github.com/Jguer/yay/v10/pkg/stringset"
)

// ParseNumberMenu parses input for number menus split by spaces or commas
// supports individual selection: 1 2 3 4
// supports range selections: 1-4 10-20
// supports negation: ^1 ^1-4
//
// include and exclude holds numbers that should be added and should not be added
// respectively. other holds anything that can't be parsed as an int. This is
// intended to allow words inside of number menus. e.g. 'all' 'none' 'abort'
// of course the implementation is up to the caller, this function mearley parses
// the input and organizes it
func ParseNumberMenu(input string) (include, exclude intrange.IntRanges,
	otherInclude, otherExclude stringset.StringSet) {
	include = make(intrange.IntRanges, 0)
	exclude = make(intrange.IntRanges, 0)
	otherInclude = make(stringset.StringSet)
	otherExclude = make(stringset.StringSet)

	words := strings.FieldsFunc(input, func(c rune) bool {
		return unicode.IsSpace(c) || c == ','
	})

	for _, word := range words {
		var num1 int
		var num2 int
		var err error
		invert := false
		other := otherInclude

		if word[0] == '^' {
			invert = true
			other = otherExclude
			word = word[1:]
		}

		ranges := strings.SplitN(word, "-", 2)

		num1, err = strconv.Atoi(ranges[0])
		if err != nil {
			other.Set(strings.ToLower(word))
			continue
		}

		if len(ranges) == 2 {
			num2, err = strconv.Atoi(ranges[1])
			if err != nil {
				other.Set(strings.ToLower(word))
				continue
			}
		} else {
			num2 = num1
		}

		mi := intrange.Min(num1, num2)
		ma := intrange.Max(num1, num2)

		if !invert {
			include = append(include, intrange.New(mi, ma))
		} else {
			exclude = append(exclude, intrange.New(mi, ma))
		}
	}

	return include, exclude, otherInclude, otherExclude
}

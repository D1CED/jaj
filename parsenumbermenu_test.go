package main

import (
	"testing"

	"github.com/Jguer/yay/v10/pkg/intrange"
	"github.com/Jguer/yay/v10/pkg/stringset"
)

func intRangesEqual(a, b intrange.IntRanges) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for n := range a {
		r1 := a[n]
		r2 := b[n]

		if r1 != r2 {
			return false
		}
	}

	return true
}

func TestParseNumberMenu(t *testing.T) {
	type result struct {
		Include      intrange.IntRanges
		Exclude      intrange.IntRanges
		OtherInclude stringset.StringSet
		OtherExclude stringset.StringSet
	}

	inputs := []string{
		"1 2 3 4 5",
		"1-10 5-15",
		"10-5 90-85",
		"1 ^2 ^10-5 99 ^40-38 ^123 60-62",
		"abort all none",
		"a-b ^a-b ^abort",
		"-9223372036854775809-9223372036854775809",
		"1\t2   3      4\t\t  \t 5",
		"1 2,3, 4,  5,6 ,7  ,8",
		"",
		"   \t   ",
		"A B C D E",
	}

	type IntRanges = intrange.IntRanges
	makeIntRange := intrange.New

	expected := []result{
		{IntRanges{
			makeIntRange(1, 1),
			makeIntRange(2, 2),
			makeIntRange(3, 3),
			makeIntRange(4, 4),
			makeIntRange(5, 5),
		}, IntRanges{}, make(stringset.StringSet), make(stringset.StringSet)},
		{IntRanges{
			makeIntRange(1, 10),
			makeIntRange(5, 15),
		}, IntRanges{}, make(stringset.StringSet), make(stringset.StringSet)},
		{IntRanges{
			makeIntRange(5, 10),
			makeIntRange(85, 90),
		}, IntRanges{}, make(stringset.StringSet), make(stringset.StringSet)},
		{
			IntRanges{
				makeIntRange(1, 1),
				makeIntRange(99, 99),
				makeIntRange(60, 62),
			},
			IntRanges{
				makeIntRange(2, 2),
				makeIntRange(5, 10),
				makeIntRange(38, 40),
				makeIntRange(123, 123),
			},
			make(stringset.StringSet), make(stringset.StringSet),
		},
		{IntRanges{}, IntRanges{}, stringset.Make("abort", "all", "none"), make(stringset.StringSet)},
		{IntRanges{}, IntRanges{}, stringset.Make("a-b"), stringset.Make("abort", "a-b")},
		{IntRanges{}, IntRanges{}, stringset.Make("-9223372036854775809-9223372036854775809"), make(stringset.StringSet)},
		{IntRanges{
			makeIntRange(1, 1),
			makeIntRange(2, 2),
			makeIntRange(3, 3),
			makeIntRange(4, 4),
			makeIntRange(5, 5),
		}, IntRanges{}, make(stringset.StringSet), make(stringset.StringSet)},
		{IntRanges{
			makeIntRange(1, 1),
			makeIntRange(2, 2),
			makeIntRange(3, 3),
			makeIntRange(4, 4),
			makeIntRange(5, 5),
			makeIntRange(6, 6),
			makeIntRange(7, 7),
			makeIntRange(8, 8),
		}, IntRanges{}, make(stringset.StringSet), make(stringset.StringSet)},
		{IntRanges{}, IntRanges{}, make(stringset.StringSet), make(stringset.StringSet)},
		{IntRanges{}, IntRanges{}, make(stringset.StringSet), make(stringset.StringSet)},
		{IntRanges{}, IntRanges{}, stringset.Make("a", "b", "c", "d", "e"), make(stringset.StringSet)},
	}

	for n, in := range inputs {
		res := expected[n]
		include, exclude, otherInclude, otherExclude := ParseNumberMenu(in)

		if !intRangesEqual(include, res.Include) ||
			!intRangesEqual(exclude, res.Exclude) ||
			!stringset.Equal(otherInclude, res.OtherInclude) ||
			!stringset.Equal(otherExclude, res.OtherExclude) {
			t.Fatalf("Test %d Failed: Expected: include=%+v exclude=%+v otherInclude=%+v otherExclude=%+v got include=%+v excluive=%+v otherInclude=%+v otherExclude=%+v",
				n+1, res.Include, res.Exclude, res.OtherInclude, res.OtherExclude, include, exclude, otherInclude, otherExclude)
		}
	}
}

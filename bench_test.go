package main_test

import "testing"

func BenchmarkAppend(b *testing.B) {

	const c = "some constant part"

	b.Run("two_appends", func(b *testing.B) {
		buf := make([]byte, 0, 1000)
		for i := 0; i < b.N; i++ {
			buf = append(buf, c...)
			buf = append(buf, byte(i))
		}
	})

	b.Run("single_append", func(b *testing.B) {
		buf := make([]byte, 0, 1000)
		for i := 0; i < b.N; i++ {
			buf = append(buf, c+string(byte(i))...)
		}
	})
}

func BenchmarkChooseMany(b *testing.B) {

	var val string

	b.Run("map", func(b *testing.B) {
		m := map[int]string{
			1:  "one",
			2:  "two",
			3:  "three",
			4:  "four",
			5:  "fife",
			6:  "six",
			7:  "seven",
			8:  "eight",
			9:  "nine",
			10: "ten",
			11: "eleven",
			12: "twelve",
			13: "thirteen",
			14: "fourteen",
			15: "fifeteen",
		}

		for i := 0; i < b.N; i++ {
			val = m[i]
		}
	})

	b.Run("switch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			switch i {
			case 1:
				val = "one"
			case 2:
				val = "two"
			case 3:
				val = "three"
			case 4:
				val = "four"
			case 5:
				val = "fife"
			case 6:
				val = "six"
			case 7:
				val = "seven"
			case 8:
				val = "eight"
			case 9:
				val = "nine"
			case 10:
				val = "ten"
			case 11:
				val = "eleven"
			case 12:
				val = "twelve"
			case 13:
				val = "thirteen"
			case 14:
				val = "fourteen"
			case 15:
				val = "fifeteen"
			default:
				val = ""
			}
		}
	})

	b.Log(val)
}

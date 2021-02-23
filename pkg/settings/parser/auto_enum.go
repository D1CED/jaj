package parser

import (
	"errors"
)

// Format:
//
// a(all):
// b
// c(clear)
// d
// e:
// [emit]
//
// longid :: alnum ( alnum | '-' )*
// single ::  ( (alnum '(' longid ')')? | '[' longid ']' ) ws (':' ws)?
// full   :: ws single+

func Enumerate(enumDescription string, withHelp bool) (func(string) Enum, func(string) (Enum, bool)) {
	ops, err := parse(enumDescription)
	if err != nil {
		panic(err)
	}

	type enArg struct {
		enum     Enum
		takesArg bool
	}

	m := make(map[string]enArg, len(ops)*2)
	for i, op := range ops {
		if op.ShortName != 0 {
			if _, ok := m[string(op.ShortName)]; ok {
				panic("same option twice")
			}
			m[string(op.ShortName)] = enArg{Enum(i), op.TakesArg}
		}
		if op.LongName != "" {
			if _, ok := m[op.LongName]; ok {
				panic("same option twice")
			}
			m[op.LongName] = enArg{Enum(i), op.TakesArg}
		}
	}

	fn := func(s string) (Enum, bool) {
		ea, ok := m[s]
		if !ok {
			return InvalidFlag, false
		}
		return ea.enum, ea.takesArg
	}

	return func(s string) Enum {
		e, _ := fn(s)
		if e == InvalidFlag {
			panic("unknown arg")
		}
		return e
	}, fn
}

func isWS(r rune) bool {
	switch r {
	default:
		return false
	case ' ', '\t', '\n', '\r':
		return true
	}
}

func isAlnum(r rune) bool {
	if ('0' <= r && r <= '9') || 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' {
		return true
	}
	return false
}

type option struct {
	ShortName rune
	LongName  string
	TakesArg  bool
}

func parse(s string) ([]option, error) {

	var ErrParse = errors.New("Parse error")

	const (
		slong = iota
		long
		end
		afterEnd
		err
	)

	gathered := []option{}
	startIdx := 0
	cur := option{}
	state := afterEnd

	for i, r := range s {
		switch state {
		case afterEnd:
			if isWS(r) {
				break
			}
			if r == '[' {
				state = slong
				break
			}
			if isAlnum(r) {
				cur.ShortName = r
				state = end
				break
			}
			state = err
		case end:
			if r == '(' {
				state = slong
				break
			}
			if r == '[' {
				gathered = append(gathered, cur)
				cur = option{}
				state = slong
				break
			}
			if isWS(r) {
				state = end
				break
			}
			if r == ':' {
				cur.TakesArg = true
				gathered = append(gathered, cur)
				cur = option{}
				state = afterEnd
				break
			}
			if isAlnum(r) {
				gathered = append(gathered, cur)
				cur = option{ShortName: r}
				state = end
				break
			}
			state = err
		case slong:
			if !isAlnum(r) {
				state = err
				break
			}
			startIdx = i
			state = long
		case long:
			if isAlnum(r) || r == '-' {
				break
			}
			if r == ')' || r == ']' {
				cur.LongName = s[startIdx:i]
				state = end
				break
			}
			state = err
		}
	}

	switch state {
	case end:
		gathered = append(gathered, cur)
		fallthrough
	case afterEnd:
		return gathered, nil
	case err, long, slong:
		return nil, ErrParse
	default:
		panic("invalid state")
	}
}

package parser

import (
	"errors"
	"io"
	"os"
	"text/tabwriter"
)

// Enumerate creates an enumeration mapping to command line options form a string.
// It can also generate a help text for you.
//
// Format
//
// The format is inspired by getopt.
//
//     alnum  = (* any alphanumeric character *) ;
//     ws     = (* any whitespace character *) ;
//     wss    = { ws } ;
//     longid = alnum , { alnum | "-" } ;
//     single = wss , ( alnum , [ "(" , longid , ")" ] ) | ( "[" , longid , "]" ) , [ ":" ] ;
//     full   = single , { single } , wss ;
func Enumerate(enumDescription string) (func(string) Enum, OptionMapping) {
	return enumerate(enumDescription, false, "")
}

func EnumerateWithHelp(enumDescription, helpText string) (func(string) Enum, OptionMapping) {
	return enumerate(enumDescription, true, helpText)
}

func enumerate(enumDescription string, withHelp bool, helpText string) (func(string) Enum, OptionMapping) {
	ops, err := parse(enumDescription)
	if err != nil {
		panic(err)
	}

	if withHelp {
		ops = append(ops, option{ShortName: 'h', LongName: "help"})
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
			return InvalidOption, false
		}
		return ea.enum, ea.takesArg
	}

	hfn := fn
	if withHelp {
		hfn = func(s string) (Enum, bool) {
			r, b := fn(s)
			if r == m["help"].enum {
				printDoc(ops, helpText)
			}
			return r, b
		}
	}

	return func(s string) Enum {
		e, _ := fn(s)
		if e == InvalidOption {
			panic("unknown arg")
		}
		return e
	}, hfn
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
	var ErrEmpty = errors.New("Empty")

	const (
		start    = iota // a
		short           // abe
		afterEnd        // ae
		long            // c, f
		lend            // d, g

		err // s
	)

	gathered := []option{}
	startIdx := 0
	cur := option{}
	state := start
	endToken := ')'

	for i, r := range s {
		switch state {
		case start:
			if isWS(r) {
				break
			}
			if isAlnum(r) {
				cur.ShortName = r
				state = short
				break
			}
			if '[' == r {
				state = long
				endToken = ']'
				break
			}
			state = err
		case short:
			if isWS(r) {
				gathered = append(gathered, cur)
				cur = option{}
				state = start
				break
			}
			if r == ':' {
				cur.TakesArg = true
				gathered = append(gathered, cur)
				cur = option{}
				state = start
				break
			}
			if isAlnum(r) {
				gathered = append(gathered, cur)
				cur = option{ShortName: r}
				break
			}
			if '[' == r {
				endToken = ']'
				state = long
				break
			}
			if '(' == r {
				endToken = ')'
				state = long
				break
			}
			state = err
		case afterEnd:
			if isAlnum(r) {
				if cur != (option{}) {
					gathered = append(gathered, cur)
					cur = option{ShortName: r}
				}
				cur.ShortName = r
				state = short
				break
			}
			if r == '[' {
				endToken = ']'
				state = long
			}
			if isWS(r) {
				state = start
			}
			if r == ':' {
				cur.TakesArg = true
				state = start
			}
			if r == ':' || r == '[' || isWS(r) {
				gathered = append(gathered, cur)
				cur = option{}
				break
			}
			state = err
		case long:
			if isAlnum(r) {
				startIdx = i
				state = lend
				break
			}
			state = err
		case lend:
			if isAlnum(r) || r == '-' {
				break
			}
			if r == endToken {
				cur.LongName = s[startIdx:i]
				state = afterEnd
				break
			}
			state = err
		}
	}

	switch state {
	case start, afterEnd, short:
		if cur != (option{}) {
			gathered = append(gathered, cur)
		}
		if len(gathered) == 0 {
			return nil, ErrEmpty
		}
		return gathered, nil
	case err, lend, long:
		return nil, ErrParse
	default:
		panic("invalid state")
	}
}

func printDoc(opts []option, helpText string) {

	E := func(_ int, err error) {
		if err != nil {
			panic(err)
		}
	}

	if helpText != "" {
		E(os.Stdout.WriteString(helpText))
		if helpText[len(helpText)-1] != '\n' {
			E(os.Stdout.WriteString("\n\n"))
		} else if helpText[len(helpText)-2] != '\n' {
			E(os.Stdout.WriteString("\n"))
		}
	}

	E(os.Stdout.WriteString(os.Args[0]))
	E(os.Stdout.WriteString(" [options] <arguments>\n\n"))
	E(os.Stdout.WriteString("OPTIONS\n"))
	err := doc(opts, os.Stdout)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

func doc(opts []option, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 6, 8, 1, ' ', 0)

	for _, o := range opts {
		tw.Write([]byte("  "))
		if o.ShortName != 0 && o.LongName != "" {
			tw.Write([]byte("-" + string(o.ShortName) + "\t--" + o.LongName))
		} else if o.ShortName != 0 {
			tw.Write([]byte("-" + string(o.ShortName) + "\t"))
		} else {
			tw.Write([]byte("\t--" + o.LongName))
		}
		if o.TakesArg {
			tw.Write([]byte("\t<arg>\n"))
		} else {
			tw.Write([]byte("\t\n"))
		}
	}
	return tw.Flush()
}

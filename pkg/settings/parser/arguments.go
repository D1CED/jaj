/*
Package parser accumulates command line arguments

It supports short and long options.
You define an enum and mapping first.
You gather all arguments with the Iterate method.

Example

	const (
    	a Enum = iota // -a or --all bool
        b             // -b          val
        c             // -c          bool
        abc           // --abc       bool
        def           // --def       val
	)

	switch flag {
	case "a", "all": return a, false
	case "b":        return b, true
	case "c":        return c, false
	case "abc":      return abc, false
	case "def":      return def, true
	}

    my-program -a target5 --abc -abc -bvalue1 target4 -b value2 --def value1 target2 --def=value2 target1 target3

	Result:

	a    true (2)
	b    [c value1 value2]
	c    false
	abc  true (1)
	def  [value1 value2]

	targets [target5 target4 target2 target1 target3]


Think about explicit error types.
*/
package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Enum int

const InvalidFlag Enum = -1

type opArgs struct {
	Option    Enum
	Arguments []string
}

type Arguments struct {
	options []opArgs
	targets []string
	alias   func(string) (Enum, bool)
}

func getIdx(op Enum, ops []opArgs) (int, bool) {
	for i, v := range ops {
		if op == v.Option {
			return i, true
		}
	}
	return 0, false
}

func put(a *Arguments, e Enum, param string, nonBoolOp bool) {
	idx, ok := getIdx(e, a.options)
	if ok {
		if nonBoolOp {
			a.options[idx].Arguments = append(a.options[idx].Arguments, param)
		} else {
			a.options[idx].Arguments[0] = string([]byte{a.options[idx].Arguments[0][0] + 1})
		}
	} else {
		if nonBoolOp {
			a.options = append(a.options, opArgs{e, []string{param}})
		} else {
			a.options = append(a.options, opArgs{e, []string{"\x01"}})
		}
	}
}

func Parse(fn func(string) (Enum, bool), args []string, r io.Reader) (*Arguments, error) {

	var (
		usedNext = false
		dash     = false
		ddash    = false
		err      error
		a        = &Arguments{
			alias:   fn,
			options: make([]opArgs, 0, len(args)),
			targets: make([]string, 0, len(args)),
		}
	)

	for k, arg := range args {

		if usedNext {
			usedNext = false
			continue
		}

		var nextArg string
		var nextArgExists bool
		if k+1 < len(args) {
			if false /* len(args[k+1]) > 1 && args[k+1][0] == '-' */ {
				nextArg = ""
			} else {
				nextArg = args[k+1]
				nextArgExists = true
			}
		} else {
			nextArg = ""
			nextArgExists = false
		}

		switch {
		case arg == "--":
			ddash = true
		case arg == "-":
			dash = true
		case strings.HasPrefix(arg, "--"):
			usedNext, err = parseLongOption(a, arg[2:], nextArg, nextArgExists)
		case strings.HasPrefix(arg, "-"):
			usedNext, err = parseShortOption(a, arg[1:], nextArg, nextArgExists)
		default:
			fallthrough
		case ddash:
			a.targets = append(a.targets, arg)
		}

		if err != nil {
			return nil, fmt.Errorf("at index %d: %w", k, err)
		}
	}

	if dash {
		err = parseIn(a, r)
	}

	return a, err
}

func parseShortOption(a *Arguments, arg, param string, exists bool) (bool, error) {

	for k, char := range arg {
		alias, wantArg := a.alias(string(char))

		if alias == InvalidFlag {
			return false, fmt.Errorf("unknown argument %c", char)
		}

		if wantArg {

			if k == len(arg)-1 {
				put(a, alias, param, true)
				if !exists {
					return false, fmt.Errorf("missing value for %c", char)
				}
				return true, nil
			} else {
				put(a, alias, arg[k+1:], true)
				return false, nil
			}

		} else {
			put(a, alias, "", false)
		}
	}
	return false, nil
}

func parseLongOption(a *Arguments, arg, param string, exists bool) (bool, error) {

	split := strings.SplitN(arg, "=", 2)
	arg = split[0]
	next := true
	if len(split) == 2 {
		param = split[1]
		exists = true
		next = false
	}

	alias, wantArg := a.alias(arg)

	if alias == InvalidFlag {
		return false, fmt.Errorf("unknown argument %q", arg)
	}
	if wantArg && !exists {
		return false, fmt.Errorf("missing value for %q", arg)
	}
	if wantArg {
		put(a, alias, param, true)

	} else {
		next = false
		put(a, alias, "", false)
	}
	return next, nil
}

func parseIn(a *Arguments, r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		a.targets = append(a.targets, scanner.Text())
	}

	return scanner.Err()
}

func (a *Arguments) Exists(op Enum) bool {
	_, ok := getIdx(op, a.options)
	return ok
}

func (a *Arguments) Get(op Enum) []string {
	idx, ok := getIdx(op, a.options)
	if !ok {
		return nil
	}
	return a.options[idx].Arguments
}

func (a *Arguments) Last(op Enum) (string, bool) {
	res := a.Get(op)
	if len(res) == 0 {
		return "", false
	}
	return res[len(res)-1], true
}

func (a *Arguments) Count(op Enum) int {
	idx, ok := getIdx(op, a.options)
	if !ok {
		return 0
	}
	ss := a.options[idx].Arguments
	return GetCount(ss)
}

func (a *Arguments) Targets() []string {
	return a.targets
}

func (a *Arguments) Iterate(fn func(Enum, []string) bool) {
	for _, kv := range a.options {
		if !fn(kv.Option, kv.Arguments) {
			return
		}
	}
}

func GetCount(ss []string) int {
	// '!' is 33 and the first ascii non-whitespace character
	if len(ss) != 1 || len(ss[0]) != 1 || ss[0][0] > '!' {
		panic("Count likely called on non-bool option")
	}
	return int(ss[0][0])
}

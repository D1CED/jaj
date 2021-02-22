/*
Package parser accumulates command line arguments

It supports short and long options.
You define an enum and mapping first.
You gather all arguments with the Iterate method.

Example

    -a bool
    -b val
    -c bool

    --abc bool
    --def val

    my-program -a target5 --abc -abc -bvalue1 target4 -b value2 --def value1 target2 --def=value2 target1 target3

	Result:

	a    true
	b    [c value1 value2]
	c    false
	abc  true
	def  [value1 value2]

	targets [target5 target4 target2 target1 target3]


Think about counting argument type.
Think about explicit error types.

TODO: implement strict ordering
*/
package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Enum int

const targetSlot = -1

const InvalidFlag Enum = -1

type Arguments struct {
	options map[Enum][]string
	alias   func(string) (Enum, bool)
}

func Parse(fn func(string) (Enum, bool), args []string, r io.Reader) (Arguments, error) {

	var (
		m        = make(map[Enum][]string)
		usedNext = false
		dash     = false
		ddash    = false
		err      error
		a        = Arguments{m, fn}
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
			m[targetSlot] = append(m[targetSlot], arg)
		}

		if err != nil {
			return Arguments{}, err
		}
	}

	if dash {
		err = parseIn(a, r)
	}

	return a, err
}

func parseShortOption(a Arguments, arg, param string, exists bool) (bool, error) {

	for k, char := range arg {
		alias, wantArg := a.alias(string(char))

		if alias == InvalidFlag {
			return false, fmt.Errorf("unknown argument %c", char)
		}

		if wantArg {

			if k == len(arg)-1 {
				a.options[alias] = append(a.options[alias], param)
				if !exists {
					return false, fmt.Errorf("missing value for %c", char)
				}
				return true, nil
			} else {
				a.options[alias] = append(a.options[alias], arg[k+1:])
				return false, nil
			}

		} else {
			a.options[alias] = make([]string, 0)
		}
	}
	return false, nil
}

func parseLongOption(a Arguments, arg, param string, exists bool) (bool, error) {

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
		a.options[alias] = append(a.options[alias], param)

	} else {
		next = false
		a.options[alias] = make([]string, 0)
	}
	return next, nil
}

func parseIn(a Arguments, r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		a.options[targetSlot] = append(a.options[targetSlot], scanner.Text())
	}

	return scanner.Err()
}

func (a Arguments) Exists(option string) bool {
	op, _ := a.alias(option)
	_, ok := a.options[op]
	return ok
}

func (a Arguments) Get(option string) []string {
	op, _ := a.alias(option)
	return a.options[op]
}

func (a Arguments) Targets() []string {
	return a.options[targetSlot]
}

func (a Arguments) Iterate(fn func(Enum, []string) bool) {
	for k, vv := range a.options {
		if k == targetSlot {
			continue
		}
		if !fn(k, vv) {
			break
		}
	}
}

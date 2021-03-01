package parser

import (
	"io"
	"strings"
)

// IMPROVE: add support for multiple levels of nesting subcommands

const (
	subcommandSentinel = -2

	readyStatus      = 0
	doneStatus       = 1
	subcommandStatus = 2
)

type specOpts struct {
	alias    []string
	id       Enum
	takesArg bool
	cb       func([]string)
}

type Spec struct {
	specs    []specOpts
	subSpecs []struct {
		alias []string
		Spec  *Spec
		cb    func()
	}
	targets func([]string)
	ctr     *int
	done    uint8 // see *Status
}

func (s *Spec) Specify(str string, takesArg bool, fn func([]string)) {
	aliases := split(str)
	if s.ctr == nil {
		s.ctr = new(int)
	}
	(*s.ctr)++
	id := *s.ctr
	s.specs = append(s.specs, specOpts{aliases, Enum(id), takesArg, fn})
}

func (s *Spec) Targets(fn func([]string)) {
	s.targets = fn
}

func (s *Spec) Sub(str string, fn func()) *Spec {
	if s.done == subcommandStatus {
		panic("no sub on sub")
	}
	if s.ctr == nil {
		s.ctr = new(int)
	}
	ss := split(str)

	sub := &Spec{ctr: s.ctr, done: subcommandStatus}
	s.subSpecs = append(s.subSpecs, struct {
		alias []string
		Spec  *Spec
		cb    func()
	}{ss, sub, fn})
	return sub
}

func (s *Spec) Parse(args []string, r *io.Reader) error {
	if s.done != readyStatus {
		panic("done")
	}
	s.done = doneStatus

	if s.ctr == nil {
		return nil
	}

	// ------------

	stack := []*Spec{s}

	mapping := func(option string) (Enum, bool) {
		for i := len(stack) - 1; i >= 0; i-- {
			s := stack[i]
			for _, sub := range s.subSpecs {
				for _, als := range sub.alias {
					if als == option {
						if i != len(stack)-1 {
							return InvalidOption, false
						}
						stack = append(stack, sub.Spec)
						sub.cb()
						return subcommandSentinel, false
					}
				}
			}
			for _, opts := range s.specs {
				for _, opt := range opts.alias {
					if opt == option {
						return opts.id, opts.takesArg
					}
				}
			}
		}
		return InvalidOption, false
	}

	a, err := Parse(mapping, args, r)
	if err != nil {
		return err
	}

	// ------------

	m := make(map[Enum]func([]string))

	for _, opts := range s.specs {
		m[opts.id] = opts.cb
	}
	for _, sub := range s.subSpecs {
		for _, opts := range sub.Spec.specs {
			m[opts.id] = opts.cb
		}
	}

	iter := func(enum Enum, ss []string) bool {
		if enum == subcommandSentinel {
			return true
		}
		m[enum](ss)
		return true
	}

	a.Iterate(iter)

	return nil
}

func split(s string) []string {
	return strings.Split(s, " ")
}

package settings

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/Jguer/yay/v10/pkg/text"
)

type Option struct {
	Global bool
	Args   []string
}

func (o *Option) Add(args ...string) {
	if o.Args == nil {
		o.Args = args
		return
	}
	o.Args = append(o.Args, args...)
}

func (o *Option) First() string {
	if len(o.Args) == 0 {
		return ""
	}
	return o.Args[0]
}

func (o *Option) Set(arg string) {
	o.Args = []string{arg}
}

func (o *Option) String() string {
	return fmt.Sprintf("Global:%v Args:%v", o.Global, o.Args)
}

// Arguments Parses command line arguments in a way we can interact with programmatically but
// also in a way that can easily be passed to pacman later on.
type Arguments struct {
	Op      string
	Options map[string]*Option
	Targets []string
}

func MakeArguments() *Arguments {
	return &Arguments{
		"",
		make(map[string]*Option),
		make([]string, 0),
	}
}

func (a *Arguments) String() string {
	return fmt.Sprintf("Op:%v Options:%+v Targets: %v", a.Op, a.Options, a.Targets)
}

func (a *Arguments) CreateOrAppendOption(option string, values ...string) {
	if a.Options[option] == nil {
		a.Options[option] = &Option{
			Args: values,
		}
	} else {
		a.Options[option].Add(values...)
	}
}

func (a *Arguments) CopyGlobal() *Arguments {
	cp := MakeArguments()
	for k, v := range a.Options {
		if v.Global {
			cp.Options[k] = v
		}
	}

	return cp
}

func (a *Arguments) Copy() (cp *Arguments) {
	cp = MakeArguments()

	cp.Op = a.Op

	for k, v := range a.Options {
		cp.Options[k] = v
	}

	cp.Targets = make([]string, len(a.Targets))
	copy(cp.Targets, a.Targets)

	return
}

func (a *Arguments) DelArg(options ...string) {
	for _, option := range options {
		delete(a.Options, option)
	}
}

func (a *Arguments) addOP(op string) error {
	if a.Op != "" {
		return errors.New(text.T("only one operation may be used at a time"))
	}

	a.Op = op
	return nil
}

func (a *Arguments) addParam(option, arg string) error {
	if !isArg(option) {
		return errors.New(text.Tf("invalid option '%s'", option))
	}

	if isOp(option) {
		return a.addOP(option)
	}

	a.CreateOrAppendOption(option, strings.Split(arg, ",")...)

	if isGlobal(option) {
		a.Options[option].Global = true
	}
	return nil
}

func (a *Arguments) AddArg(options ...string) error {
	for _, option := range options {
		err := a.addParam(option, "")
		if err != nil {
			return err
		}
	}
	return nil
}

// Multiple args acts as an OR operator
func (a *Arguments) ExistsArg(options ...string) bool {
	for _, option := range options {
		if _, exists := a.Options[option]; exists {
			return true
		}
	}
	return false
}

func (a *Arguments) GetArg(options ...string) (arg string, double, exists bool) {
	for _, option := range options {
		value, exists := a.Options[option]
		if exists {
			return value.First(), len(value.Args) >= 2, len(value.Args) >= 1
		}
	}

	return arg, false, false
}

func (a *Arguments) GetArgs(option string) (args []string) {
	value, exists := a.Options[option]
	if exists {
		return value.Args
	}

	return nil
}

func (a *Arguments) AddTarget(targets ...string) {
	a.Targets = append(a.Targets, targets...)
}

func (a *Arguments) ClearTargets() {
	a.Targets = make([]string, 0)
}

// Multiple args acts as an OR operator
func (a *Arguments) ExistsDouble(options ...string) bool {
	for _, option := range options {
		if value, exists := a.Options[option]; exists {
			return len(value.Args) >= 2
		}
	}
	return false
}

func formatArg(arg string) string {
	if len(arg) > 1 {
		arg = "--" + arg
	} else {
		arg = "-" + arg
	}

	return arg
}

func (a *Arguments) FormatArgs() (args []string) {
	if a.Op != "" {
		args = append(args, formatArg(a.Op))
	}

	for option, arg := range a.Options {
		if arg.Global || option == "--" {
			continue
		}

		formattedOption := formatArg(option)
		for _, value := range arg.Args {
			args = append(args, formattedOption)
			if hasParam(option) {
				args = append(args, value)
			}
		}
	}
	return args
}

func (a *Arguments) FormatGlobals() (args []string) {
	for option, arg := range a.Options {
		if !arg.Global {
			continue
		}
		formattedOption := formatArg(option)

		for _, value := range arg.Args {
			args = append(args, formattedOption)
			if hasParam(option) {
				args = append(args, value)
			}
		}
	}
	return args
}

// Parses short hand options such as:
// -Syu -b/some/path -
func (a *Arguments) parseShortOption(arg, param string) (usedNext bool, err error) {
	if arg == "-" {
		err = a.AddArg("-")
		return
	}

	arg = arg[1:]

	for k, _char := range arg {
		char := string(_char)

		if hasParam(char) {
			if k < len(arg)-1 {
				err = a.addParam(char, arg[k+1:])
			} else {
				usedNext = true
				err = a.addParam(char, param)
			}

			break
		} else {
			err = a.AddArg(char)

			if err != nil {
				return
			}
		}
	}

	return
}

// Parses full length options such as:
// --sync --refresh --sysupgrade --dbpath /some/path --
func (a *Arguments) parseLongOption(arg, param string) (usedNext bool, err error) {
	if arg == "--" {
		err = a.AddArg(arg)
		return
	}

	arg = arg[2:]

	split := strings.SplitN(arg, "=", 2)
	switch {
	case len(split) == 2:
		err = a.addParam(split[0], split[1])
	case hasParam(arg):
		err = a.addParam(arg, param)
		usedNext = true
	default:
		err = a.AddArg(arg)
	}

	return
}

func (a *Arguments) parseStdin() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		a.AddTarget(scanner.Text())
	}

	return os.Stdin.Close()
}

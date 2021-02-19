package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
)

/*
type option struct {
	Global bool
	Args   []string
}

func (o *option) Add(args ...string) {
	if o.Args == nil {
		o.Args = args
		return
	}
	o.Args = append(o.Args, args...)
}

func (o *option) First() string {
	if len(o.Args) == 0 {
		return ""
	}
	return o.Args[0]
}

func (o *option) Set(arg string) {
	o.Args = []string{arg}
}

func (o *option) String() string {
	return fmt.Sprintf("Global:%v Args:%v", o.Global, o.Args)
}
*/

func First(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func add(a *Arguments, key string, values ...string) {
	a.Options[key] = append(a.Options[key], values...)
}

// Arguments Parses command line arguments in a way we can interact with programmatically but
// also in a way that can easily be passed to pacman later on.
type Arguments struct {
	Op      string
	Options map[string][]string
	Targets []string

	isArg    stringset.StringSet
	isOp     stringset.StringSet
	isGlobal stringset.StringSet
	hasParam stringset.StringSet
}

func New(isArg, isOp, isGlobal, hasParam []string) *Arguments {
	return &Arguments{
		Op:      "",
		Options: make(map[string][]string),
		Targets: nil,

		isArg:    stringset.FromSlice(isArg),
		isOp:     stringset.FromSlice(isOp),
		isGlobal: stringset.FromSlice(isGlobal),
		hasParam: stringset.FromSlice(hasParam),
	}
}

func (a *Arguments) Parse(args []string) error {
	usedNext := false

	for k, arg := range args {
		var nextArg string

		if usedNext {
			usedNext = false
			continue
		}

		if k+1 < len(args) {
			nextArg = args[k+1]
		}

		var err error
		switch {
		case a.ExistsArg("--"):
			a.AddTarget(arg)
		case strings.HasPrefix(arg, "--"):
			usedNext, err = a.parseLongOption(arg, nextArg)
		case strings.HasPrefix(arg, "-"):
			usedNext, err = a.parseShortOption(arg, nextArg)
		default:
			a.AddTarget(arg)
		}

		if err != nil {
			return err
		}
	}

	if a.Op == "" {
		a.Op = "Y"
	}

	if a.ExistsArg("-") {
		if err := a.parseStdin(); err != nil {
			return err
		}
		a.DelArg("-")

		file, err := os.Open("/dev/tty")
		if err != nil {
			return err
		}

		os.Stdin = file
	}

	return nil
}

func (a *Arguments) String() string {
	return fmt.Sprintf("Op:%v Options:%+v Targets: %v", a.Op, a.Options, a.Targets)
}

func (a *Arguments) CreateOrAppendOption(optionStr string, values ...string) {
	if a.Options[optionStr] == nil {
		a.Options[optionStr] = values
	} else {
		a.Options[optionStr] = append(a.Options[optionStr], values...)
	}
}

func (a *Arguments) CopyGlobal() *Arguments {
	cp := &Arguments{
		isArg:    a.isArg,
		isOp:     a.isOp,
		isGlobal: a.isGlobal,
		hasParam: a.hasParam,

		Op:      "",
		Options: make(map[string][]string, len(a.Options)),
		Targets: nil,
	}

	for k, v := range a.Options {
		if a.isGlobal.Get(k) {
			cp.Options[k] = v
		}
	}

	return cp
}

func (a *Arguments) Copy() *Arguments {
	cp := &Arguments{
		Op:       a.Op,
		isArg:    a.isArg,
		isOp:     a.isOp,
		isGlobal: a.isGlobal,
		hasParam: a.hasParam,

		Options: make(map[string][]string, len(a.Options)),
		Targets: make([]string, len(a.Targets)),
	}

	for k, v := range a.Options {
		cp.Options[k] = v
	}
	copy(cp.Targets, a.Targets)

	return cp
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
	if !a.isArg.Get(option) {
		return errors.New(text.Tf("invalid option '%s'", option))
	}

	if a.isOp.Get(option) {
		return a.addOP(option)
	}

	a.CreateOrAppendOption(option, strings.Split(arg, ",")...)

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
			return First(value), len(value) >= 2, len(value) >= 1
		}
	}

	return arg, false, false
}

func (a *Arguments) GetArgs(option string) (args []string) {
	value, exists := a.Options[option]
	if exists {
		return value[0:len(value):len(value)]
	}

	return nil
}

func (a *Arguments) AddTarget(targets ...string) { a.Targets = append(a.Targets, targets...) }

func (a *Arguments) ClearTargets() { a.Targets = make([]string, 0) }

// Multiple args acts as an OR operator
func (a *Arguments) ExistsDouble(options ...string) bool {
	for _, option := range options {
		if value, exists := a.Options[option]; exists {
			return len(value) >= 2
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
		if a.isGlobal.Get(option) || option == "--" {
			continue
		}

		formattedOption := formatArg(option)
		for _, value := range arg {
			args = append(args, formattedOption)
			if a.hasParam.Get(option) {
				args = append(args, value)
			}
		}
	}
	return args
}

func (a *Arguments) FormatGlobals() (args []string) {
	for option, arg := range a.Options {
		if !a.isGlobal.Get(option) {
			continue
		}
		formattedOption := formatArg(option)

		for _, value := range arg {
			args = append(args, formattedOption)
			if a.hasParam.Get(option) {
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

		if a.hasParam.Get(char) {
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
	case a.hasParam.Get(arg):
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

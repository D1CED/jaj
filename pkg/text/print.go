package text

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/leonelquinteros/gotext"
	"golang.org/x/sys/unix"
)

const (
	arrow      = "==>"
	smallArrow = " ->"
	opSymbol   = "::"
)

var cachedColumnCount = -1

var (
	Out    io.Writer
	ErrOut io.Writer
	In     io.Reader
)

// CaptureOutput takes two io.Writer and runs a function so that the output gets written to those.
// If you supply nil for the writers the output will be discarded.
//
// Be very careful with concurrent use!!!
// No writing function may run concurrently.
func CaptureOutput(out, errOut io.Writer, f func()) {
	if out == nil {
		out = ioutil.Discard
	}
	if errOut == nil {
		errOut = ioutil.Discard
	}

	safeOut := Out
	safeErrOut := ErrOut

	Out = out
	ErrOut = errOut

	f()

	Out = safeOut
	ErrOut = safeErrOut
}

func Print(a ...interface{}) {
	fmt.Fprint(Out, a...)
}

func Println(a ...interface{}) {
	fmt.Fprintln(Out, a...)
}

func Printf(format string, a ...interface{}) {
	fmt.Fprintf(Out, format, a...)
}

func EPrint(a ...interface{}) {
	fmt.Fprint(ErrOut, a...)
}

func EPrintln(a ...interface{}) {
	fmt.Fprintln(ErrOut, a...)
}

func EPrintf(format string, a ...interface{}) {
	fmt.Fprintf(ErrOut, format, a...)
}

func OperationInfoln(a ...interface{}) {
	fmt.Fprint(Out, append([]interface{}{Bold(Cyan(opSymbol + " ")), boldCode}, a...)...)
	fmt.Fprintln(Out, ResetCode)
}

func OperationInfo(a ...interface{}) {
	fmt.Fprint(Out, append([]interface{}{Bold(Cyan(opSymbol + " ")), boldCode}, a...)...)
	fmt.Fprint(Out, ResetCode)
}

func SprintOperationInfo(a ...interface{}) string {
	return fmt.Sprint(append([]interface{}{Bold(Cyan(opSymbol + " ")), boldCode}, a...)...) + ResetCode
}

func Info(a ...interface{}) {
	fmt.Fprint(Out, append([]interface{}{Bold(Green(arrow + " "))}, a...)...)
}

func Infoln(a ...interface{}) {
	fmt.Fprintln(Out, append([]interface{}{Bold(Green(arrow))}, a...)...)
}

func SprintWarn(a ...interface{}) string {
	return fmt.Sprint(append([]interface{}{Bold(yellow(smallArrow + " "))}, a...)...)
}

func Warn(a ...interface{}) {
	fmt.Fprint(Out, append([]interface{}{Bold(yellow(smallArrow + " "))}, a...)...)
}

func Warnln(a ...interface{}) {
	fmt.Fprintln(Out, append([]interface{}{Bold(yellow(smallArrow))}, a...)...)
}

func SprintError(a ...interface{}) string {
	return fmt.Sprint(append([]interface{}{Bold(Red(smallArrow + " "))}, a...)...)
}

func Error(a ...interface{}) {
	fmt.Fprint(ErrOut, append([]interface{}{Bold(Red(smallArrow + " "))}, a...)...)
}

func Errorln(a ...interface{}) {
	fmt.Fprintln(ErrOut, append([]interface{}{Bold(Red(smallArrow))}, a...)...)
}

func getColumnCount() int {
	if cachedColumnCount > 0 {
		return cachedColumnCount
	}
	if count, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil {
		cachedColumnCount = count
		return cachedColumnCount
	}
	if ws, err := unix.IoctlGetWinsize(syscall.Stdout, unix.TIOCGWINSZ); err == nil {
		cachedColumnCount = int(ws.Col)
		return cachedColumnCount
	}
	return 80
}

func PrintInfoValue(key string, values ...string) {
	// 16 (text) + 1 (:) + 1 ( )
	const (
		keyLength  = 18
		delimCount = 2
	)

	str := fmt.Sprintf(Bold("%-16s: "), key)
	if len(values) == 0 || (len(values) == 1 && values[0] == "") {
		fmt.Fprintf(Out, "%s%s\n", str, gotext.Get("None"))
		return
	}

	maxCols := getColumnCount()
	cols := keyLength + len(values[0])
	str += values[0]
	for _, value := range values[1:] {
		if maxCols > keyLength && cols+len(value)+delimCount >= maxCols {
			cols = keyLength
			str += "\n" + strings.Repeat(" ", keyLength)
		} else if cols != keyLength {
			str += strings.Repeat(" ", delimCount)
			cols += delimCount
		}
		str += value
		cols += len(value)
	}
	fmt.Println(str)
}

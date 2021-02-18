package text

import (
	"fmt"
	"os"

	"github.com/leonelquinteros/gotext"
)

// TrTemplate or translation template distinguishes itself from string
// to avoid accidental misuse. String literals work fine.
type TrTemplate string

func Init(localePath string) {
	if envLocalePath := os.Getenv("LOCALE_PATH"); envLocalePath != "" {
		localePath = envLocalePath
	}

	gotext.Configure(localePath, os.Getenv("LANG"), "yay")

	out = os.Stdout
	errOut = os.Stderr
}

func T(s TrTemplate) string { return gotext.Get(string(s)) }

func Tf(s TrTemplate, args ...interface{}) string { return gotext.Get(string(s), args...) }

func ErrT(s TrTemplate) error { return fmt.Errorf(gotext.Get(string(s))) }

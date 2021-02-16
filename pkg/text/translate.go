package text

import (
	"os"

	"github.com/leonelquinteros/gotext"
)

func Init(localePath string) {
	if envLocalePath := os.Getenv("LOCALE_PATH"); envLocalePath != "" {
		localePath = envLocalePath
	}

	gotext.Configure(localePath, os.Getenv("LANG"), "yay")
}

func T(s string) string { return gotext.Get(s) }

func Tf(s string, args ...interface{}) string { return gotext.Get(s, args...) }

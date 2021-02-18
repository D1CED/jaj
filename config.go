package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/text"
)

// Verbosity settings for search
const (
	numberMenu = iota
	detailed
	minimal
)

var yayVersion = "10.1.0"

var localePath = "/usr/share/locale"

// Editor returns the preferred system editor.
func editor(edt, editFlags string) (editor string, args []string) {
	switch {
	case edt != "":
		editor, err := exec.LookPath(edt)
		if err != nil {
			text.EPrintln(err)
		} else {
			return editor, strings.Fields(editFlags)
		}
		fallthrough
	case os.Getenv("EDITOR") != "":
		if editorArgs := strings.Fields(os.Getenv("EDITOR")); len(editorArgs) != 0 {
			editor, err := exec.LookPath(editorArgs[0])
			if err != nil {
				text.EPrintln(err)
			} else {
				return editor, editorArgs[1:]
			}
		}
		fallthrough
	case os.Getenv("VISUAL") != "":
		if editorArgs := strings.Fields(os.Getenv("VISUAL")); len(editorArgs) != 0 {
			editor, err := exec.LookPath(editorArgs[0])
			if err != nil {
				text.EPrintln(err)
			} else {
				return editor, editorArgs[1:]
			}
		}
		fallthrough
	default:
		text.EPrintln()
		text.Errorln(text.Tf("%s is not set", text.Bold(text.Cyan("$EDITOR"))))
		text.Warnln(text.Tf("Add %s or %s to your environment variables", text.Bold(text.Cyan("$EDITOR")), text.Bold(text.Cyan("$VISUAL"))))

		for {
			text.Infoln(text.T("Edit PKGBUILD with?"))
			editorInput, err := getInput("")
			if err != nil {
				text.EPrintln(err)
				continue
			}

			editorArgs := strings.Fields(editorInput)
			if len(editorArgs) == 0 {
				continue
			}

			editor, err := exec.LookPath(editorArgs[0])
			if err != nil {
				text.EPrintln(err)
				continue
			}
			return editor, editorArgs[1:]
		}
	}
}

func getInput(defaultValue string) (string, error) {
	text.Info()
	if defaultValue != "" || settings.NoConfirm {
		text.Println(defaultValue)
		return defaultValue, nil
	}

	reader := bufio.NewReader(text.In())

	buf, overflow, err := reader.ReadLine()
	if err != nil {
		return "", err
	}

	if overflow {
		return "", text.ErrT("input too long")
	}

	return string(buf), nil
}

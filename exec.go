package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/exe"
	"github.com/Jguer/yay/v10/pkg/settings/parser"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/text"
)

func sudoLoopBackground(cmdRunner exe.Runner, conf *settings.Configuration) {
	updateSudo(cmdRunner, conf)
	go sudoLoop(cmdRunner, conf)
}

func sudoLoop(cmdRunner exe.Runner, conf *settings.Configuration) {
	for {
		updateSudo(cmdRunner, conf)
		time.Sleep(298 * time.Second)
	}
}

func updateSudo(cmdRunner exe.Runner, conf *settings.Configuration) {
	for {
		mSudoFlags := strings.Fields(conf.SudoFlags)
		mSudoFlags = append([]string{"-v"}, mSudoFlags...)
		err := cmdRunner.Show(exec.Command(conf.SudoBin, mSudoFlags...))
		if err != nil {
			text.EPrintln(err)
		} else {
			break
		}
	}
}

// waitLock will lock yay checking the status of db.lck until it does not exist
func waitLock(dbPath string) {
	lockDBPath := filepath.Join(dbPath, "db.lck")
	if _, err := os.Stat(lockDBPath); err != nil {
		return
	}

	text.Warnln(text.Tf("%s is present.", lockDBPath))
	text.Warn(text.T("There may be another Pacman instance running. Waiting..."))

	for {
		time.Sleep(3 * time.Second)
		if _, err := os.Stat(lockDBPath); err != nil {
			text.Println()
			return
		}
	}
}

func passToPacman(rt *runtime.Runtime, args *parser.Arguments) *exec.Cmd {
	argArr := make([]string, 0, 32)

	if settings.NeedRoot(args, rt.Mode) {
		argArr = append(argArr, rt.Config.SudoBin)
		argArr = append(argArr, strings.Fields(rt.Config.SudoFlags)...)
	}

	argArr = append(argArr, rt.Config.PacmanBin)
	argArr = append(argArr, args.FormatGlobals()...)
	argArr = append(argArr, args.FormatArgs()...)
	if settings.NoConfirm {
		argArr = append(argArr, "--noconfirm")
	}

	argArr = append(argArr, "--config", rt.Config.PacmanConf, "--")
	argArr = append(argArr, args.Targets...)

	if settings.NeedRoot(args, rt.Mode) {
		waitLock(rt.PacmanConf.DBPath)
	}
	return exec.Command(argArr[0], argArr[1:]...)
}

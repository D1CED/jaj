package yay

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Jguer/yay/v10/pkg/exe"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/text"
)

func SudoLoopBackground(cmdRunner exe.Runner, conf *settings.YayConfig) {
	updateSudo(cmdRunner, conf)
	go sudoLoop(cmdRunner, conf)
}

func sudoLoop(cmdRunner exe.Runner, conf *settings.YayConfig) {
	for {
		updateSudo(cmdRunner, conf)
		time.Sleep(298 * time.Second)
	}
}

func updateSudo(cmdRunner exe.Runner, conf *settings.YayConfig) {
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

func PassToPacman(rt *Runtime, args *settings.PacmanConf) *exec.Cmd {
	argArr := make([]string, 0, 32)

	if settings.NeedRoot(args, rt.Config.Mode) {
		argArr = append(argArr, rt.Config.SudoBin)
		argArr = append(argArr, strings.Fields(rt.Config.SudoFlags)...)
	}

	argArr = append(argArr, rt.Config.PacmanBin)
	argArr = append(argArr, args.FormatGlobalArgs()...)
	argArr = append(argArr, args.FormatAsArgs()...)
	if rt.Config.Pacman.NoConfirm {
		argArr = append(argArr, "--noconfirm")
	}

	argArr = append(argArr, "--config", rt.Config.PacmanConf, "--")
	argArr = append(argArr, *args.Targets...)

	if settings.NeedRoot(args, rt.Config.Mode) {
		waitLock(rt.Config.Pacman.DBPath)
	}
	return exec.Command(argArr[0], argArr[1:]...)
}

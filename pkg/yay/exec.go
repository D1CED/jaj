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

func sudoLoopBackground(cmdRunner exe.Runner, conf *settings.YayConfig) {
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

func needRoot(p *settings.PacmanConf, tMode settings.TargetMode, help, version bool) bool {
	if help || version {
		return false
	}

	switch mode := p.ModeConf.(type) {
	default:
		return false
	case *settings.SConf:
		if mode.Refresh != 0 {
			return true
		}
		if mode.Print {
			return false
		}
		if mode.Search != "" {
			return false
		}
		if mode.List {
			return false
		}
		if mode.Groups != 0 {
			return false
		}
		if mode.Info != 0 {
			return false
		}
		if mode.Clean != 0 && tMode == settings.ModeAUR {
			return false
		}
		return true

	case *settings.RConf:
		if mode.Print {
			return false
		}
		return true
	case *settings.QConf:
		if mode.Check != 0 {
			return true
		}
		return false

	case *settings.DConf:
		if mode.Check != 0 {
			return false
		}
		return true
	case *settings.FConf:
		if mode.Refresh != 0 {
			return true
		}
		return false
	case *settings.UConf:
		return true
	}
}

func PassToPacman(rt *Runtime, args *settings.PacmanConf) *exec.Cmd {
	argArr := make([]string, 0, 32)

	help := rt.Config.MainOperation == settings.OpHelp
	version := rt.Config.MainOperation == settings.OpVersion
	needRoot := needRoot(args, rt.Config.Mode, help, version)

	if needRoot {
		argArr = append(argArr, rt.Config.SudoBin)
		argArr = append(argArr, strings.Fields(rt.Config.SudoFlags)...)
	}

	argArr = append(argArr, rt.Config.PacmanBin)
	argArr = append(argArr, args.FormatAsArgs(
		rt.Config.MainOperation == settings.OpHelp,
		rt.Config.MainOperation == settings.OpVersion)...,
	)

	if len(*args.Targets) != 0 {
		argArr = append(argArr, "--")
		argArr = append(argArr, *args.Targets...)
	}

	if needRoot {
		waitLock(rt.Config.Pacman.DBPath)
	}
	return exec.Command(argArr[0], argArr[1:]...)
}

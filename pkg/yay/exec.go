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
		if mode.Search {
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

func PassToPacman(conf *settings.YayConfig, args *settings.PacmanConf) *exec.Cmd {
	argArr := make([]string, 0, 32)

	help := conf.MainOperation == settings.OpHelp
	version := conf.MainOperation == settings.OpVersion
	needRoot := needRoot(args, conf.Mode, help, version)

	if needRoot {
		argArr = append(argArr, conf.SudoBin)
		argArr = append(argArr, strings.Fields(conf.SudoFlags)...)
	}

	argArr = append(argArr, conf.PacmanBin)
	argArr = append(argArr, args.FormatAsArgs(help, version)...)

	if len(*args.Targets) != 0 {
		argArr = append(argArr, "--")
		argArr = append(argArr, *args.Targets...)
	}

	if needRoot {
		waitLock(conf.Pacman.DBPath)
	}
	return exec.Command(argArr[0], argArr[1:]...)
}

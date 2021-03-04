package main

import (
	"github.com/Jguer/go-alpm/v2"

	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/yay"
)

const yayVersion = "devel"

func usage() { text.Print(settings.Usage) }

func handleCmd(rt *yay.Runtime) error {

	switch rt.Config.MainOperation {
	case settings.OpDatabase, settings.OpFiles, settings.OpDepTest, settings.OpUpgrade:
		return rt.CmdRunner.Show(yay.PassToPacman(rt, rt.Config.Pacman))
	case settings.OpHelp:
		return handleHelp(rt)
	case settings.OpVersion:
		return handleVersion()
	case settings.OpQuery:
		return yay.HandleQuery(rt, rt.Config.Pacman.ModeConf.(*settings.QConf))
	case settings.OpRemove:
		return yay.HandleRemove(rt.Config.Pacman.ModeConf.(*settings.RConf), rt)
	case settings.OpSync:
		return yay.HandleSync(rt.Config.Pacman.ModeConf.(*settings.SConf), rt)
	case settings.OpGetPkgbuild:
		return yay.HandleGetpkgbuild(rt.Config.ModeConf.(*settings.GConf), rt)
	case settings.OpShow:
		return yay.HandlePrint(rt.Config.ModeConf.(*settings.PConf), yayVersion, rt)
	case settings.OpYay:
		return yay.HandleYay(rt.Config.ModeConf.(*settings.YConf), rt)
	default:
		return text.ErrT("unhandled operation")
	}
}

func handleHelp(rt *yay.Runtime) error {
	if rt.Config.IsPacmanOp() {
		return rt.CmdRunner.Show(yay.PassToPacman(rt, rt.Config.Pacman))
	}
	usage()
	return nil
}

func handleVersion() error {
	text.Printf("yay v%s - libalpm v%s\n", yayVersion, alpm.Version())
	return nil
}

package yay

import (
	"github.com/Jguer/yay/v10/pkg/completion"
	"github.com/Jguer/yay/v10/pkg/news"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/text"
)

func HandleQuery(rt *Runtime, cmdArgs *settings.QConf) error {
	if rt.Config.SudoLoop && cmdArgs.Check != 0 {
		sudoLoopBackground(rt.CmdRunner, rt.Config)
	}

	if cmdArgs.Upgrades != 0 {
		return printUpdateList(rt.Config.Pacman, rt, cmdArgs.Upgrades > 1)
	}
	return rt.CmdRunner.Show(PassToPacman(rt.Config, rt.Config.Pacman))
}

func HandlePrint(cmdArgs *settings.PConf, yayVersion string, rt *Runtime) (err error) {
	switch {
	case cmdArgs.DefaultConfig:
		text.Println(settings.Defaults().AsJSONString())
	case cmdArgs.CurrentConfig:
		text.Println(rt.Config.AsJSONString())
	case cmdArgs.NumberUpgrades:
		err = printNumberOfUpdates(rt, false)
	case cmdArgs.News:
		double := cmdArgs.News
		quiet := cmdArgs.Quiet
		err = news.PrintNewsFeed(rt.HttpClient, rt.DB.LastBuildTime(), rt.Config.SortMode, double, quiet)
	case cmdArgs.Complete != 0:
		err = completion.Show(
			rt.DB, rt.HttpClient, rt.Config.AURURL, rt.Config.CompletionPath,
			rt.Config.CompletionInterval, cmdArgs.Complete > 1)
	case cmdArgs.LocalStats:
		err = localStatistics(rt.DB, rt.AUR, yayVersion, rt.Config.RequestSplitN)
	}
	return err
}

func HandleYay(cmdArgs *settings.YConf, rt *Runtime) error {
	if cmdArgs.GenDevDB {
		return createDevelDB(rt)
	}
	if cmdArgs.Clean != 0 {
		return cleanDependencies(rt, rt.Config.Pacman, cmdArgs.Clean > 1)
	}
	if len(rt.Config.Targets) > 0 {
		return handleYogurt(rt.Config.Pacman, rt)
	}
	return nil
}

func HandleGetpkgbuild(cmdArgs *settings.GConf, rt *Runtime) error {
	return getPkgbuilds(rt.Config.Targets, rt, cmdArgs.Force)
}

func handleYogurt(cmdArgs *settings.PacmanConf, rt *Runtime) error {
	rt.Config.SearchMode = settings.NumberMenu
	return displayNumberMenu(*cmdArgs.Targets, rt)
}

func HandleSync(cmdArgs *settings.SConf, rt *Runtime) error {
	if rt.Config.SudoLoop &&
		(cmdArgs.Refresh != 0 || !(cmdArgs.Print ||
			cmdArgs.Search ||
			cmdArgs.List ||
			cmdArgs.Groups != 0 ||
			cmdArgs.Info != 0 ||
			(cmdArgs.Clean != 0 && rt.Config.Mode == settings.ModeAUR))) {

		sudoLoopBackground(rt.CmdRunner, rt.Config)
	}

	targets := rt.Config.Targets

	if cmdArgs.Search {
		if cmdArgs.Quiet {
			rt.Config.SearchMode = settings.Minimal
		} else {
			rt.Config.SearchMode = settings.Detailed
		}
		return syncSearch(targets, rt)
	}
	if cmdArgs.Print {
		return rt.CmdRunner.Show(PassToPacman(rt.Config, rt.Config.Pacman))
	}
	if cmdArgs.Clean != 0 {
		return syncClean(rt)
	}
	if cmdArgs.List {
		return syncList(rt, cmdArgs.Quiet)
	}
	if cmdArgs.Groups != 0 {
		return rt.CmdRunner.Show(PassToPacman(rt.Config, rt.Config.Pacman))
	}
	if cmdArgs.Info != 0 {
		return syncInfo(rt.Config.Pacman, targets, rt)
	}
	if cmdArgs.SysUpgrade != 0 {
		return install(rt, rt.Config.Pacman, cmdArgs, false)
	}
	if len(rt.Config.Targets) > 0 {
		return install(rt, rt.Config.Pacman, cmdArgs, false)
	}
	if cmdArgs.Refresh != 0 {
		return rt.CmdRunner.Show(PassToPacman(rt.Config, rt.Config.Pacman))
	}
	return nil
}

func HandleRemove(cmdArgs *settings.RConf, rt *Runtime) error {
	if rt.Config.SudoLoop && !cmdArgs.Print {
		sudoLoopBackground(rt.CmdRunner, rt.Config)
	}

	err := rt.CmdRunner.Show(PassToPacman(rt.Config, rt.Config.Pacman))
	if err == nil {
		rt.VCSStore.RemovePackage(rt.Config.Targets)
	}
	return err
}

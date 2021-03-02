package yay

import (
	"bufio"
	"net/http"

	"github.com/Jguer/yay/v10/pkg/completion"
	"github.com/Jguer/yay/v10/pkg/news"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/view"
)

func HandleQuery(rt *Runtime, cmdArgs *settings.QConf) error {
	if cmdArgs.Upgrades != 0 {
		return printUpdateList(rt.Config.Pacman, rt, cmdArgs.Upgrades > 1)
	}
	return rt.CmdRunner.Show(PassToPacman(rt, rt.Config.Pacman))
}

func HandlePrint(cmdArgs *settings.PConf, yayVersion string, rt *Runtime) (err error) {
	switch {
	case cmdArgs.DefaultConfig:
		text.Println(settings.Defaults().AsJSONString())
	case cmdArgs.CurrentConfig:
		text.Printf("%v", rt.Config.AsJSONString())
	case cmdArgs.NumberUpgrades:
		err = printNumberOfUpdates(rt, false)
	case cmdArgs.News:
		double := cmdArgs.News
		quiet := cmdArgs.Quiet
		err = news.PrintNewsFeed(rt.DB.LastBuildTime(), rt.Config.SortMode, double, quiet)
	case cmdArgs.Complete != 0:
		err = completion.Show(rt.DB, rt.Config.AURURL, rt.Config.CompletionPath, rt.Config.CompletionInterval, cmdArgs.Complete > 1)
	case cmdArgs.LocalStats:
		err = localStatistics(rt.DB, yayVersion, rt.Config.RequestSplitN)
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
		return HandleYogurt(rt.Config.Pacman, rt)
	}
	return nil
}

func HandleGetpkgbuild(cmdArgs *settings.GConf, rt *Runtime) error {
	return getPkgbuilds(rt.Config.Targets, rt, cmdArgs.Force)
}

func HandleYogurt(cmdArgs *settings.PacmanConf, rt *Runtime) error {
	rt.Config.SearchMode = settings.NumberMenu
	return displayNumberMenu(*cmdArgs.Targets, rt)
}

func HandleSync(cmdArgs *settings.SConf, rt *Runtime) error {
	targets := rt.Config.Pacman.Targets

	if cmdArgs.Search != "" {
		if cmdArgs.Quiet {
			rt.Config.SearchMode = settings.Minimal
		} else {
			rt.Config.SearchMode = settings.Detailed
		}
		return syncSearch(*targets, rt)
	}
	if cmdArgs.Print {
		return rt.CmdRunner.Show(PassToPacman(rt, rt.Config.Pacman))
	}
	if cmdArgs.Clean != 0 {
		return syncClean(rt)
	}
	if cmdArgs.List {
		return syncList(rt, cmdArgs.Quiet)
	}
	if cmdArgs.Groups != 0 {
		return rt.CmdRunner.Show(PassToPacman(rt, rt.Config.Pacman))
	}
	if cmdArgs.Info != 0 {
		return syncInfo(rt.Config.Pacman, *targets, rt)
	}
	if cmdArgs.SysUpgrade != 0 {
		return install(rt, rt.Config.Pacman, cmdArgs, false)
	}
	if len(rt.Config.Targets) > 0 {
		return install(rt, rt.Config.Pacman, cmdArgs, false)
	}
	if cmdArgs.Refresh != 0 {
		return rt.CmdRunner.Show(PassToPacman(rt, rt.Config.Pacman))
	}
	return nil
}

func HandleRemove(cmdArgs *settings.RConf, rt *Runtime) error {
	err := rt.CmdRunner.Show(PassToPacman(rt, rt.Config.Pacman))
	if err == nil {
		rt.VCSStore.RemovePackage(rt.Config.Targets)
	}
	return err
}

// NumberMenu presents a CLI for selecting packages to install.
func displayNumberMenu(pkgS []string, rt *Runtime) error {
	var (
		aurErr, repoErr error
		aq              aurQuery
		pq              repoQuery
		lenaq, lenpq    int
	)

	pkgS = query.RemoveInvalidTargets(pkgS, rt.Config.Mode)

	if rt.Config.Mode == settings.ModeAUR || rt.Config.Mode == settings.ModeAny {
		aq, aurErr = narrowSearch(pkgS, true, rt.Config.SearchBy, rt.Config.SortBy)
		lenaq = len(aq)
	}
	if rt.Config.Mode == settings.ModeRepo || rt.Config.Mode == settings.ModeAny {
		pq = queryRepo(pkgS, rt.DB, rt.Config.SortMode)
		lenpq = len(pq)
		if repoErr != nil {
			return repoErr
		}
	}

	if lenpq == 0 && lenaq == 0 {
		return text.ErrT("no packages match search")
	}

	switch rt.Config.SortMode {
	case settings.TopDown:
		if rt.Config.Mode == settings.ModeRepo || rt.Config.Mode == settings.ModeAny {
			printSearchRepo(pq, rt.DB, rt.Config.SearchMode, rt.Config.SortMode)
		}
		if rt.Config.Mode == settings.ModeAUR || rt.Config.Mode == settings.ModeAny {
			printSearchAUR(aq, rt.DB, lenpq+1, rt.Config.SearchMode, rt.Config.SortMode)
		}
	case settings.BottomUp:
		if rt.Config.Mode == settings.ModeAUR || rt.Config.Mode == settings.ModeAny {
			printSearchAUR(aq, rt.DB, lenpq+1, rt.Config.SearchMode, rt.Config.SortMode)
		}
		if rt.Config.Mode == settings.ModeRepo || rt.Config.Mode == settings.ModeAny {
			printSearchRepo(pq, rt.DB, rt.Config.SearchMode, rt.Config.SortMode)
		}
	default:
		return text.ErrT("invalid sort mode. Fix with yay -Y --bottomup --save")
	}

	if aurErr != nil {
		text.Errorln(text.Tf("Error during AUR search: %s\n", aurErr))
		text.Warnln(text.T("Showing repo packages only"))
	}

	text.Infoln(text.T("Packages to install (eg: 1 2 3, 1-3 or ^4)"))
	text.Info()

	reader := bufio.NewReader(text.In())

	numberBuf, overflow, err := reader.ReadLine()
	if err != nil {
		return err
	}
	if overflow {
		return text.ErrT("input too long")
	}

	include, exclude, _, otherExclude := view.ParseNumberMenu(string(numberBuf))
	arguments := rt.Config.Pacman.DeepCopy()

	isInclude := len(exclude) == 0 && len(otherExclude) == 0

	for i, pkg := range pq {
		var target int
		switch rt.Config.SortMode {
		case settings.TopDown:
			target = i + 1
		case settings.BottomUp:
			target = len(pq) - i
		default:
			return text.ErrT("invalid sort mode. Fix with yay -Y --bottomup --save")
		}

		if (isInclude && include.Get(target)) || (!isInclude && !exclude.Get(target)) {
			*arguments.Targets = append(*arguments.Targets, pkg.DB().Name()+"/"+pkg.Name())
		}
	}

	for i := range aq {
		var target int

		switch rt.Config.SortMode {
		case settings.TopDown:
			target = i + 1 + len(pq)
		case settings.BottomUp:
			target = len(aq) - i + len(pq)
		default:
			return text.ErrT("invalid sort mode. Fix with yay -Y --bottomup --save")
		}

		if (isInclude && include.Get(target)) || (!isInclude && !exclude.Get(target)) {
			*arguments.Targets = append(*arguments.Targets, "aur/"+aq[i].Name)
		}
	}

	if len(*arguments.Targets) == 0 {
		text.Println(text.T(" there is nothing to do"))
		return nil
	}

	if rt.Config.SudoLoop {
		SudoLoopBackground(rt.CmdRunner, rt.Config)
	}

	return install(rt, arguments, arguments.ModeConf.(*settings.SConf), true)
}

func syncList(rt *Runtime, quiet bool) error {
	aur := false

	for i := len(rt.Config.Targets) - 1; i >= 0; i-- {
		if rt.Config.Targets[i] == "aur" && (rt.Config.Mode == settings.ModeAny || rt.Config.Mode == settings.ModeAUR) {

			rt.Config.Targets = append(rt.Config.Targets[:i], rt.Config.Targets[i+1:]...)
			aur = true
		}
	}

	if (rt.Config.Mode == settings.ModeAny || rt.Config.Mode == settings.ModeAUR) && (len(rt.Config.Targets) == 0 || aur) {
		resp, err := http.Get(rt.Config.AURURL + "/packages.gz")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		scanner.Scan()
		for scanner.Scan() {
			name := scanner.Text()
			if quiet {
				text.Println(name)
			} else {
				text.Printf("%s %s %s", text.Magenta("aur"), text.Bold(name), text.Bold(text.Green(text.T("unknown-version"))))

				if rt.DB.LocalPackage(name) != nil {
					text.Print(text.Bold(text.Blue(text.T(" [Installed]"))))
				}

				text.Println()
			}
		}
	}

	if (rt.Config.Mode == settings.ModeAny || rt.Config.Mode == settings.ModeRepo) && (len(rt.Config.Targets) != 0 || !aur) {
		return rt.CmdRunner.Show(PassToPacman(rt, rt.Config.Pacman))
	}

	return nil
}

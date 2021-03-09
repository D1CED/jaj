package yay

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/view"
)

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
		aq, aurErr = narrowSearch(rt.AUR, pkgS, true, rt.Config.SearchBy, rt.Config.SortBy)
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

	isInclude := len(exclude) == 0 && otherExclude.Len() == 0

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
		sudoLoopBackground(rt.CmdRunner, rt.Config)
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
		return rt.CmdRunner.Show(PassToPacman(rt.Config, rt.Config.Pacman))
	}

	return nil
}

// PrintInfo prints package info like pacman -Si.
func printInfo(a *query.Pkg, aurURL string, extendedInfo bool) {
	text.PrintInfoValue(text.T("Repository"), "aur")
	text.PrintInfoValue(text.T("Name"), a.Name)
	text.PrintInfoValue(text.T("Keywords"), a.Keywords...)
	text.PrintInfoValue(text.T("Version"), a.Version)
	text.PrintInfoValue(text.T("Description"), a.Description)
	text.PrintInfoValue(text.T("URL"), a.URL)
	text.PrintInfoValue(text.T("AUR URL"), aurURL+"/packages/"+a.Name)
	text.PrintInfoValue(text.T("Groups"), a.Groups...)
	text.PrintInfoValue(text.T("Licenses"), a.License...)
	text.PrintInfoValue(text.T("Provides"), a.Provides...)
	text.PrintInfoValue(text.T("Depends On"), a.Depends...)
	text.PrintInfoValue(text.T("Make Deps"), a.MakeDepends...)
	text.PrintInfoValue(text.T("Check Deps"), a.CheckDepends...)
	text.PrintInfoValue(text.T("Optional Deps"), a.OptDepends...)
	text.PrintInfoValue(text.T("Conflicts With"), a.Conflicts...)
	text.PrintInfoValue(text.T("Maintainer"), a.Maintainer)
	text.PrintInfoValue(text.T("Votes"), fmt.Sprintf("%d", a.NumVotes))
	text.PrintInfoValue(text.T("Popularity"), fmt.Sprintf("%f", a.Popularity))
	text.PrintInfoValue(text.T("First Submitted"), text.FormatTimeQuery(a.FirstSubmitted))
	text.PrintInfoValue(text.T("Last Modified"), text.FormatTimeQuery(a.LastModified))

	if a.OutOfDate != 0 {
		text.PrintInfoValue(text.T("Out-of-date"), text.FormatTimeQuery(a.OutOfDate))
	} else {
		text.PrintInfoValue(text.T("Out-of-date"), "No")
	}

	if extendedInfo {
		text.PrintInfoValue("ID", fmt.Sprintf("%d", a.ID))
		text.PrintInfoValue(text.T("Package Base ID"), fmt.Sprintf("%d", a.PackageBaseID))
		text.PrintInfoValue(text.T("Package Base"), a.PackageBase)
		text.PrintInfoValue(text.T("Snapshot URL"), aurURL+a.URLPath)
	}

	text.Println()
}

// BiggestPackages prints the name of the ten biggest packages in the system.
func biggestPackages(dbExecutor db.Executor) {
	pkgS := dbExecutor.BiggestPackages()

	if len(pkgS) < 10 {
		return
	}

	for i := 0; i < 10; i++ {
		text.Printf("%s: %s\n", text.Bold(pkgS[i].Name()), text.Cyan(text.Human(pkgS[i].ISize())))
	}
	// Could implement size here as well, but we just want the general idea
}

// localStatistics prints installed packages statistics.
func localStatistics(dbExecutor db.Executor, aur *query.AUR, yayVersion string, requestSplitN int) error {
	info := query.Statistics(dbExecutor)

	_, remoteNames, err := query.GetPackageNamesBySource(dbExecutor)
	if err != nil {
		return err
	}

	text.Infoln(text.Tf("Yay version v%s", yayVersion))
	text.Println(text.Bold(text.Cyan("===========================================")))
	text.Infoln(text.Tf("Total installed packages: %s", text.Cyan(strconv.Itoa(info.Totaln))))
	text.Infoln(text.Tf("Total foreign installed packages: %s", text.Cyan(strconv.Itoa(len(remoteNames)))))
	text.Infoln(text.Tf("Explicitly installed packages: %s", text.Cyan(strconv.Itoa(info.Expln))))
	text.Infoln(text.Tf("Total Size occupied by packages: %s", text.Cyan(text.Human(info.TotalSize))))
	text.Println(text.Bold(text.Cyan("===========================================")))
	text.Infoln(text.T("Ten biggest packages:"))
	biggestPackages(dbExecutor)
	text.Println(text.Bold(text.Cyan("===========================================")))

	query.AURInfoPrint(aur, remoteNames, requestSplitN)

	return nil
}

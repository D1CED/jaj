package main

import (
	"fmt"
	"strconv"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/parser"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/upgrade"
)

// PrintSearch handles printing search results in a given format
func (q aurQuery) printSearch(dbExecutor db.Executor, start int, searchMode enum, sortOrder enum) {
	for i := range q {
		var toprint string
		if searchMode == numberMenu {
			switch sortOrder {
			case settings.TopDown:
				toprint += text.Magenta(strconv.Itoa(start+i) + " ")
			case settings.BottomUp:
				toprint += text.Magenta(strconv.Itoa(len(q)+start-i-1) + " ")
			default:
				text.Warnln(text.T("invalid sort mode. Fix with yay -Y --bottomup --save"))
			}
		} else if searchMode == minimal {
			text.Println(q[i].Name)
			continue
		}

		toprint += text.Bold(text.ColorHash("aur")) + "/" + text.Bold(q[i].Name) +
			" " + text.Cyan(q[i].Version) +
			text.Bold(" (+"+strconv.Itoa(q[i].NumVotes)) +
			" " + text.Bold(strconv.FormatFloat(q[i].Popularity, 'f', 2, 64)+") ")

		if q[i].Maintainer == "" {
			toprint += text.Bold(text.Red(text.T("(Orphaned)"))) + " "
		}

		if q[i].OutOfDate != 0 {
			toprint += text.Bold(text.Red(text.Tf("(Out-of-date: %s)", text.FormatTime(q[i].OutOfDate)))) + " "
		}

		if pkg := dbExecutor.LocalPackage(q[i].Name); pkg != nil {
			if pkg.Version() != q[i].Version {
				toprint += text.Bold(text.Green(text.Tf("(Installed: %s)", pkg.Version())))
			} else {
				toprint += text.Bold(text.Green(text.T("(Installed)")))
			}
		}
		toprint += "\n    " + q[i].Description
		text.Println(toprint)
	}
}

// PrintSearch receives a RepoSearch type and outputs pretty text.
func (s repoQuery) printSearch(dbExecutor db.Executor, searchMode enum, sortMode enum) {
	for i, res := range s {
		var toprint string
		if searchMode == numberMenu {
			switch sortMode {
			case settings.TopDown:
				toprint += text.Magenta(strconv.Itoa(i+1) + " ")
			case settings.BottomUp:
				toprint += text.Magenta(strconv.Itoa(len(s)-i) + " ")
			default:
				text.Warnln(text.T("invalid sort mode. Fix with yay -Y --bottomup --save"))
			}
		} else if searchMode == minimal {
			text.Println(res.Name())
			continue
		}

		toprint += text.Bold(text.ColorHash(res.DB().Name())) + "/" + text.Bold(res.Name()) +
			" " + text.Cyan(res.Version()) +
			text.Bold(" ("+text.Human(res.Size())+
				" "+text.Human(res.ISize())+") ")

		packageGroups := dbExecutor.PackageGroups(res)
		if len(packageGroups) != 0 {
			toprint += fmt.Sprint(packageGroups, " ")
		}

		if pkg := dbExecutor.LocalPackage(res.Name()); pkg != nil {
			if pkg.Version() != res.Version() {
				toprint += text.Bold(text.Green(text.Tf("(Installed: %s)", pkg.Version())))
			} else {
				toprint += text.Bold(text.Green(text.T("(Installed)")))
			}
		}

		toprint += "\n    " + res.Description()
		text.Println(toprint)
	}
}

// Pretty print a set of packages from the same package base.

// PrintInfo prints package info like pacman -Si.
func PrintInfo(a *query.Pkg, aurURL string, extendedInfo bool) {
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
func localStatistics(dbExecutor db.Executor, requestSplitN int) error {
	info := statistics(dbExecutor)

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

	query.AURInfoPrint(remoteNames, requestSplitN)

	return nil
}

func printNumberOfUpdates(rt *runtime.Runtime, enableDowngrade bool) error {
	warnings := query.NewWarnings()

	var (
		aurUp  upgrade.UpSlice
		repoUp upgrade.UpSlice
		err    error
	)

	text.CaptureOutput(nil, nil, func() {
		aurUp, repoUp, err = upList(warnings, rt, enableDowngrade)
	})

	if err != nil {
		return err
	}
	text.Println(len(aurUp) + len(repoUp))

	return nil
}

func printUpdateList(cmdArgs *parser.Arguments, rt *runtime.Runtime, enableDowngrade bool) error {
	targets := stringset.FromSlice(cmdArgs.Targets)
	warnings := query.NewWarnings()

	var (
		err         error
		localNames  []string
		remoteNames []string
		aurUp       upgrade.UpSlice
		repoUp      upgrade.UpSlice
	)
	text.CaptureOutput(nil, nil, func() {
		localNames, remoteNames, err = query.GetPackageNamesBySource(rt.DB)
		if err != nil {
			return
		}

		aurUp, repoUp, err = upList(warnings, rt, enableDowngrade)
	})

	if err != nil {
		return err
	}

	noTargets := len(targets) == 0

	if !cmdArgs.ExistsArg("m", "foreign") {
		for _, pkg := range repoUp {
			if noTargets || targets.Get(pkg.Name) {
				if cmdArgs.ExistsArg("q", "quiet") {
					text.Printf("%s\n", pkg.Name)
				} else {
					text.Printf("%s %s -> %s\n", text.Bold(pkg.Name), text.Green(pkg.LocalVersion), text.Green(pkg.RemoteVersion))
				}
				delete(targets, pkg.Name)
			}
		}
	}

	if !cmdArgs.ExistsArg("n", "native") {
		for _, pkg := range aurUp {
			if noTargets || targets.Get(pkg.Name) {
				if cmdArgs.ExistsArg("q", "quiet") {
					text.Printf("%s\n", pkg.Name)
				} else {
					text.Printf("%s %s -> %s\n", text.Bold(pkg.Name), text.Green(pkg.LocalVersion), text.Green(pkg.RemoteVersion))
				}
				delete(targets, pkg.Name)
			}
		}
	}

	missing := false

outer:
	for pkg := range targets {
		for _, name := range localNames {
			if name == pkg {
				continue outer
			}
		}

		for _, name := range remoteNames {
			if name == pkg {
				continue outer
			}
		}

		text.Errorln(text.Tf("package '%s' was not found", pkg))
		missing = true
	}

	if missing {
		return ErrMissing
	}

	return nil
}

package main

import (
	"sort"
	"strings"
	"sync"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/multierror"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/upgrade"
)

// upList returns lists of packages to upgrade from each source.
func upList(warnings *query.AURWarnings, rt *runtime.Runtime, enableDowngrade bool) (aurUp, repoUp upgrade.UpSlice, err error) {
	remote, remoteNames := query.GetRemotePackages(rt.DB)

	var wg sync.WaitGroup
	var develUp upgrade.UpSlice
	var errs multierror.MultiError

	aurdata := make(map[string]*query.Pkg)

	for _, pkg := range remote {
		if pkg.ShouldIgnore() {
			warnings.Ignore.Set(pkg.Name())
		}
	}

	if rt.Config.Mode == settings.ModeAny || rt.Config.Mode == settings.ModeRepo {
		text.OperationInfoln(text.T("Searching databases for updates..."))
		wg.Add(1)
		go func() {
			repoUp, err = rt.DB.RepoUpgrades(enableDowngrade)
			errs.Add(err)
			wg.Done()
		}()
	}

	if rt.Config.Mode == settings.ModeAny || rt.Config.Mode == settings.ModeAUR {
		text.OperationInfoln(text.T("Searching AUR for updates..."))

		var _aurdata []*query.Pkg
		_aurdata, err = query.AURInfo(remoteNames, warnings, rt.Config.RequestSplitN)
		errs.Add(err)
		if err == nil {
			for _, pkg := range _aurdata {
				aurdata[pkg.Name] = pkg
			}

			wg.Add(1)
			go func() {
				aurUp = upgrade.UpAUR(remote, aurdata, rt.Config.TimeUpdate)
				wg.Done()
			}()

			if rt.Config.Devel {
				text.OperationInfoln(text.T("Checking development packages..."))
				wg.Add(1)
				go func() {
					develUp = upgrade.UpDevel(remote, aurdata, rt.VCSStore)
					wg.Done()
				}()
			}
		}
	}

	wg.Wait()

	printLocalNewerThanAUR(remote, aurdata)

	if develUp != nil {
		names := make(stringset.StringSet)
		for _, up := range develUp {
			names.Set(up.Name)
		}
		for _, up := range aurUp {
			if !names.Get(up.Name) {
				develUp = append(develUp, up)
			}
		}

		aurUp = develUp
	}

	return aurUp, repoUp, errs.Return()
}

func printLocalNewerThanAUR(remote []db.IPackage, aurdata map[string]*query.Pkg) {
	for _, pkg := range remote {
		aurPkg, ok := aurdata[pkg.Name()]
		if !ok {
			continue
		}

		left, right := upgrade.GetVersionDiff(pkg.Version(), aurPkg.Version)

		if !isDevelPackage(pkg) && db.VerCmp(pkg.Version(), aurPkg.Version) > 0 {
			text.Warnln(text.Tf("%s: local (%s) is newer than AUR (%s)",
				text.Cyan(pkg.Name()),
				left, right,
			))
		}
	}
}

func isDevelName(name string) bool {
	for _, suffix := range []string{"git", "svn", "hg", "bzr", "nightly"} {
		if strings.HasSuffix(name, "-"+suffix) {
			return true
		}
	}

	return strings.Contains(name, "-always-")
}

func isDevelPackage(pkg db.IPackage) bool {
	return isDevelName(pkg.Name()) || isDevelName(pkg.Base())
}

// upgradePkgs handles updating the cache and installing updates.
func upgradePkgs(conf *settings.YayConfig, aurUp, repoUp upgrade.UpSlice) (ignore, aurNames stringset.StringSet, err error) {
	ignore = make(stringset.StringSet)
	aurNames = make(stringset.StringSet)

	allUpLen := len(repoUp) + len(aurUp)
	if allUpLen == 0 {
		return ignore, aurNames, nil
	}

	if !conf.UpgradeMenu {
		for _, pkg := range aurUp {
			aurNames.Set(pkg.Name)
		}

		return ignore, aurNames, nil
	}

	sort.Sort(repoUp)
	sort.Sort(aurUp)
	allUp := append(repoUp, aurUp...)
	text.Printf("%s"+text.Bold(" %d ")+"%s\n", text.Bold(text.Cyan("::")), allUpLen, text.Bold(text.T("Packages to upgrade.")))
	allUp.Print()

	text.Infoln(text.T("Packages to exclude: (eg: \"1 2 3\", \"1-3\", \"^4\" or repo name)"))

	numbers, err := getInput(conf.AnswerUpgrade, conf.Pacman.NoConfirm)
	if err != nil {
		return nil, nil, err
	}

	// upgrade menu asks you which packages to NOT upgrade so in this case
	// include and exclude are kind of swapped
	include, exclude, otherInclude, otherExclude := ParseNumberMenu(numbers)

	isInclude := len(exclude) == 0 && len(otherExclude) == 0

	for i, pkg := range repoUp {
		if isInclude && otherInclude.Get(pkg.Repository) {
			ignore.Set(pkg.Name)
		}

		if isInclude && !include.Get(len(repoUp)-i+len(aurUp)) {
			continue
		}

		if !isInclude && (exclude.Get(len(repoUp)-i+len(aurUp)) || otherExclude.Get(pkg.Repository)) {
			continue
		}

		ignore.Set(pkg.Name)
	}

	for i, pkg := range aurUp {
		if isInclude && otherInclude.Get(pkg.Repository) {
			continue
		}

		if isInclude && !include.Get(len(aurUp)-i) {
			aurNames.Set(pkg.Name)
		}

		if !isInclude && (exclude.Get(len(aurUp)-i) || otherExclude.Get(pkg.Repository)) {
			aurNames.Set(pkg.Name)
		}
	}

	return ignore, aurNames, err
}

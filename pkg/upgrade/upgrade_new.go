package upgrade

import (
	"sort"
	"strings"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/view"
)

func PrintLocalNewerThanAUR(remote []db.IPackage, aurdata map[string]*query.Pkg) {
	for _, pkg := range remote {
		aurPkg, ok := aurdata[pkg.Name()]
		if !ok {
			continue
		}

		left, right := GetVersionDiff(pkg.Version(), aurPkg.Version)

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
func UpgradePkgs(conf *settings.YayConfig, aurUp, repoUp []Upgrade) (ignore, aurNames stringset.StringSet, err error) {
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

	sort.Slice(repoUp, upgradeLess(repoUp))
	sort.Slice(aurUp, upgradeLess(aurUp))
	allUp := append(repoUp, aurUp...)
	text.Printf("%s"+text.Bold(" %d ")+"%s\n", text.Bold(text.Cyan("::")), allUpLen, text.Bold(text.T("Packages to upgrade.")))
	printUpSlice(allUp)

	text.Infoln(text.T("Packages to exclude: (eg: \"1 2 3\", \"1-3\", \"^4\" or repo name)"))

	numbers, err := view.GetInput(conf.AnswerUpgrade, conf.Pacman.NoConfirm)
	if err != nil {
		return nil, nil, err
	}

	// upgrade menu asks you which packages to NOT upgrade so in this case
	// include and exclude are kind of swapped
	include, exclude, otherInclude, otherExclude := view.ParseNumberMenu(numbers)

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

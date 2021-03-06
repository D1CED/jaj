package dep

import (
	"fmt"

	"github.com/Jguer/yay/v10/pkg/db"
	rpc "github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
)

type Order struct {
	Aur     []Base
	Repo    []db.IPackage
	Runtime stringset.StringSet
}

func GetOrder(dp *Pool) *Order {
	do := &Order{
		make([]Base, 0),
		make([]db.IPackage, 0),
		stringset.Make(),
	}

	for _, target := range dp.targets {
		dep := target.DepString()
		aurPkg := dp.Aur[dep]
		if aurPkg != nil && pkgSatisfies(aurPkg.Name, aurPkg.Version, dep) {
			do.orderPkgAur(aurPkg, dp, true)
		}

		aurPkg = dp.findSatisfierAur(dep)
		if aurPkg != nil {
			do.orderPkgAur(aurPkg, dp, true)
		}

		repoPkg := dp.findSatisfierRepo(dep)
		if repoPkg != nil {
			do.orderPkgRepo(repoPkg, dp, true)
		}
	}

	return do
}

func (do *Order) orderPkgAur(pkg *rpc.Pkg, dp *Pool, runtime bool) {
	if runtime {
		do.Runtime.Set(pkg.Name)
	}
	delete(dp.Aur, pkg.Name)

	for i, deps := range [3][]string{pkg.Depends, pkg.MakeDepends, pkg.CheckDepends} {
		for _, dep := range deps {
			aurPkg := dp.findSatisfierAur(dep)
			if aurPkg != nil {
				do.orderPkgAur(aurPkg, dp, runtime && i == 0)
			}

			repoPkg := dp.findSatisfierRepo(dep)
			if repoPkg != nil {
				do.orderPkgRepo(repoPkg, dp, runtime && i == 0)
			}
		}
	}

	for i, base := range do.Aur {
		if base.Pkgbase() == pkg.PackageBase {
			do.Aur[i] = append(base, pkg)
			return
		}
	}

	do.Aur = append(do.Aur, Base{pkg})
}

func (do *Order) orderPkgRepo(pkg db.IPackage, dp *Pool, runtime bool) {
	if runtime {
		do.Runtime.Set(pkg.Name())
	}
	delete(dp.repo, pkg.Name())

	for _, dep := range dp.alpmExecutor.PackageDepends(pkg) {
		repoPkg := dp.findSatisfierRepo(dep.String())
		if repoPkg != nil {
			do.orderPkgRepo(repoPkg, dp, runtime)
		}
	}

	do.Repo = append(do.Repo, pkg)
}

func (do *Order) HasMake() bool {
	lenAur := 0
	for _, base := range do.Aur {
		lenAur += len(base)
	}

	return do.Runtime.Len() != lenAur+len(do.Repo)
}

func (do *Order) GetMake() []string {
	makeOnly := []string{}

	for _, base := range do.Aur {
		for _, pkg := range base {
			if !do.Runtime.Get(pkg.Name) {
				makeOnly = append(makeOnly, pkg.Name)
			}
		}
	}

	for _, pkg := range do.Repo {
		if !do.Runtime.Get(pkg.Name()) {
			makeOnly = append(makeOnly, pkg.Name())
		}
	}

	return makeOnly
}

// Print prints repository packages to be downloaded
func (do *Order) Print() {
	var (
		repo     = ""
		repoMake = ""
		aur      = ""
		aurMake  = ""

		repoLen     = 0
		repoMakeLen = 0
		aurLen      = 0
		aurMakeLen  = 0
	)

	for _, pkg := range do.Repo {
		pkgStr := fmt.Sprintf("  %s-%s", pkg.Name(), pkg.Version())
		if do.Runtime.Get(pkg.Name()) {
			repo += pkgStr
			repoLen++
		} else {
			repoMake += pkgStr
			repoMakeLen++
		}
	}

	for _, base := range do.Aur {
		pkg := base.Pkgbase()
		pkgStr := "  " + pkg + "-" + base[0].Version
		pkgStrMake := pkgStr

		push := false
		pushMake := false

		switch {
		case len(base) > 1, pkg != base[0].Name:
			pkgStr += " ("
			pkgStrMake += " ("

			for _, split := range base {
				if do.Runtime.Get(split.Name) {
					pkgStr += split.Name + " "
					aurLen++
					push = true
				} else {
					pkgStrMake += split.Name + " "
					aurMakeLen++
					pushMake = true
				}
			}

			pkgStr = pkgStr[:len(pkgStr)-1] + ")"
			pkgStrMake = pkgStrMake[:len(pkgStrMake)-1] + ")"
		case do.Runtime.Get(base[0].Name):
			aurLen++
			push = true
		default:
			aurMakeLen++
			pushMake = true
		}

		if push {
			aur += pkgStr
		}
		if pushMake {
			aurMake += pkgStrMake
		}
	}

	printDownloads("Repo", repoLen, repo)
	printDownloads("Repo Make", repoMakeLen, repoMake)
	printDownloads("Aur", aurLen, aur)
	printDownloads("Aur Make", aurMakeLen, aurMake)
}

func printDownloads(repoName string, length int, packages string) {
	if length < 1 {
		return
	}

	repoInfo := fmt.Sprintf(text.Bold(text.Blue("[%s:%d]")), repoName, length)
	text.Println(repoInfo + text.Cyan(packages))
}

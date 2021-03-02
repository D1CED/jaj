package yay

import (
	"fmt"
	"strconv"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/text"
)

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
func localStatistics(dbExecutor db.Executor, yayVersion string, requestSplitN int) error {
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

	query.AURInfoPrint(remoteNames, requestSplitN)

	return nil
}

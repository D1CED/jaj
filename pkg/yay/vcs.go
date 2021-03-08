package yay

import (
	"sync"

	"github.com/Jguer/yay/v10/pkg/dep"
	"github.com/Jguer/yay/v10/pkg/multierror"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/upgrade"
)

// upList returns lists of packages to upgrade from each source.
func upList(warnings *query.AURWarnings, rt *Runtime, enableDowngrade bool) (aurUp, repoUp upgrade.UpSlice, err error) {
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
		_aurdata, err = query.AURInfo(rt.AUR, remoteNames, warnings, rt.Config.RequestSplitN)
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

	upgrade.PrintLocalNewerThanAUR(remote, aurdata)

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

// createDevelDB forces yay to create a DB of the existing development packages
func createDevelDB(rt *Runtime) error {
	var mux sync.Mutex
	var wg sync.WaitGroup

	_, remoteNames, err := query.GetPackageNamesBySource(rt.DB)
	if err != nil {
		return err
	}

	info, err := query.AURInfoPrint(rt.AUR, remoteNames, rt.Config.RequestSplitN)
	if err != nil {
		return err
	}

	bases := dep.GetBases(info)
	toSkip := pkgbuildsToSkip(bases, stringset.FromSlice(remoteNames), rt.Config.ReDownload, rt.Config.BuildDir)
	_, err = downloadPkgbuilds(buildRun{rt.CmdBuilder, rt.CmdRunner}, bases, toSkip, rt.Config.BuildDir, rt.Config.AURURL)
	if err != nil {
		return err
	}

	srcinfos, err := parseSrcinfoFiles(bases, false, rt.Config.BuildDir)
	if err != nil {
		return err
	}

	for i := range srcinfos {
		for iP := range srcinfos[i].Packages {
			wg.Add(1)
			go rt.VCSStore.Update(srcinfos[i].Packages[iP].Pkgname, srcinfos[i].Source, &mux, &wg)
		}
	}

	wg.Wait()
	text.OperationInfoln(text.T("GenDB finished. No packages were installed"))
	return err
}

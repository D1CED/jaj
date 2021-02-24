package main

import (
	"sync"

	"github.com/Jguer/yay/v10/pkg/dep"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
)

// createDevelDB forces yay to create a DB of the existing development packages
func createDevelDB(rt *runtime.Runtime) error {
	var mux sync.Mutex
	var wg sync.WaitGroup

	_, remoteNames, err := query.GetPackageNamesBySource(rt.DB)
	if err != nil {
		return err
	}

	info, err := query.AURInfoPrint(remoteNames, rt.Config.Conf.RequestSplitN)
	if err != nil {
		return err
	}

	bases := dep.GetBases(info)
	toSkip := pkgbuildsToSkip(bases, stringset.FromSlice(remoteNames), rt.Config.Conf.ReDownload, rt.Config.Conf.BuildDir)
	_, err = downloadPkgbuilds(BuildRun{rt.CmdBuilder, rt.CmdRunner}, bases, toSkip, rt.Config.Conf.BuildDir, rt.Config.Conf.AURURL)
	if err != nil {
		return err
	}

	srcinfos, err := parseSrcinfoFiles(bases, false, rt.Config.Conf.BuildDir)
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

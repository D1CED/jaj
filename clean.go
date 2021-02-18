package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/dep"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
)

// CleanDependencies removes all dangling dependencies in system
func cleanDependencies(rt *runtime.Runtime, cmdArgs *settings.Arguments, removeOptional bool) error {
	hanging := hangingPackages(removeOptional, rt.DB)
	if len(hanging) != 0 {
		return cleanRemove(rt, cmdArgs, hanging)
	}

	return nil
}

// CleanRemove sends a full removal command to pacman with the pkgName slice
func cleanRemove(rt *runtime.Runtime, cmdArgs *settings.Arguments, pkgNames []string) error {
	if len(pkgNames) == 0 {
		return nil
	}

	arguments := cmdArgs.CopyGlobal()
	_ = arguments.AddArg("R")
	arguments.AddTarget(pkgNames...)

	return rt.CmdRunner.Show(passToPacman(rt, arguments))
}

func syncClean(rt *runtime.Runtime, cmdArgs *settings.Arguments) error {
	keepInstalled := false
	keepCurrent := false

	_, removeAll, _ := cmdArgs.GetArg("c", "clean")

	for _, v := range rt.PacmanConf.CleanMethod {
		if v == "KeepInstalled" {
			keepInstalled = true
		} else if v == "KeepCurrent" {
			keepCurrent = true
		}
	}

	if rt.Mode == settings.ModeRepo || rt.Mode == settings.ModeAny {
		if err := rt.CmdRunner.Show(passToPacman(rt, cmdArgs)); err != nil {
			return err
		}
	}

	if !(rt.Mode == settings.ModeAUR || rt.Mode == settings.ModeAny) {
		return nil
	}

	var question string
	if removeAll {
		question = text.T("Do you want to remove ALL AUR packages from cache?")
	} else {
		question = text.T("Do you want to remove all other AUR packages from cache?")
	}

	text.Println(text.T("\nBuild directory:"), rt.Config.BuildDir)

	if text.ContinueTask(question, true, settings.NoConfirm) {
		if err := cleanAUR(rt.Config, keepInstalled, keepCurrent, removeAll, rt.DB); err != nil {
			return err
		}
	}

	if removeAll {
		return nil
	}

	if text.ContinueTask(text.T("Do you want to remove ALL untracked AUR files?"), true, settings.NoConfirm) {
		return cleanUntracked(rt)
	}

	return nil
}

func cleanAUR(conf *settings.Configuration, keepInstalled, keepCurrent, removeAll bool, dbExecutor db.Executor) error {
	text.Println(text.T("removing AUR packages from cache..."))

	installedBases := make(stringset.StringSet)
	inAURBases := make(stringset.StringSet)

	remotePackages, _ := query.GetRemotePackages(dbExecutor)

	files, err := ioutil.ReadDir(conf.BuildDir)
	if err != nil {
		return err
	}

	cachedPackages := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		cachedPackages = append(cachedPackages, file.Name())
	}

	// Most people probably don't use keep current and that is the only
	// case where this is needed.
	// Querying the AUR is slow and needs internet so don't do it if we
	// don't need to.
	if keepCurrent {
		info, errInfo := query.AURInfo(cachedPackages, &query.AURWarnings{}, conf.RequestSplitN)
		if errInfo != nil {
			return errInfo
		}

		for _, pkg := range info {
			inAURBases.Set(pkg.PackageBase)
		}
	}

	for _, pkg := range remotePackages {
		if pkg.Base() != "" {
			installedBases.Set(pkg.Base())
		} else {
			installedBases.Set(pkg.Name())
		}
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		if !removeAll {
			if keepInstalled && installedBases.Get(file.Name()) {
				continue
			}

			if keepCurrent && inAURBases.Get(file.Name()) {
				continue
			}
		}

		err = os.RemoveAll(filepath.Join(conf.BuildDir, file.Name()))
		if err != nil {
			return nil
		}
	}

	return nil
}

func cleanUntracked(rt *runtime.Runtime) error {
	text.Println(text.T("removing untracked AUR files from cache..."))

	files, err := ioutil.ReadDir(rt.Config.BuildDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		dir := filepath.Join(rt.Config.BuildDir, file.Name())
		if isGitRepository(dir) {
			if err := rt.CmdRunner.Show(rt.CmdBuilder.BuildGitCmd(dir, "clean", "-fx")); err != nil {
				text.Warnln(text.T("Unable to clean:"), dir)
				return err
			}
		}
	}
	return nil
}

func isGitRepository(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return !os.IsNotExist(err)
}

func cleanAfter(rt *runtime.Runtime, bases []dep.Base) {
	text.Println(text.T("removing untracked AUR files from cache..."))

	for i, base := range bases {
		dir := filepath.Join(rt.Config.BuildDir, base.Pkgbase())
		if !isGitRepository(dir) {
			continue
		}

		text.OperationInfoln(text.Tf("Cleaning (%d/%d): %s", i+1, len(bases), text.Cyan(dir)))

		_, stderr, err := rt.CmdRunner.Capture(rt.CmdBuilder.BuildGitCmd(dir, "reset", "--hard", "HEAD"), 0)
		if err != nil {
			text.Errorln(text.Tf("error resetting %s: %s", base.String(), stderr))
		}

		if err := rt.CmdRunner.Show(rt.CmdBuilder.BuildGitCmd(dir, "clean", "-fx", "--exclude='*.pkg.*'")); err != nil {
			text.EPrintln(err)
		}
	}
}

func cleanBuilds(buildDir string, bases []dep.Base) {
	for i, base := range bases {
		dir := filepath.Join(buildDir, base.Pkgbase())
		text.OperationInfoln(text.Tf("Deleting (%d/%d): %s", i+1, len(bases), text.Cyan(dir)))
		if err := os.RemoveAll(dir); err != nil {
			text.EPrintln(err)
		}
	}
}

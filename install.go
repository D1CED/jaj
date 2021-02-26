package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	alpm "github.com/Jguer/go-alpm/v2"
	gosrc "github.com/Morganamilo/go-srcinfo"

	"github.com/Jguer/yay/v10/pkg/completion"
	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/dep"
	"github.com/Jguer/yay/v10/pkg/multierror"
	"github.com/Jguer/yay/v10/pkg/pgp"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/upgrade"
)

const gitEmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func asdeps(cmdArgs1 *settings.PacmanConf, rt *runtime.Runtime, pkgs []string) error {

	if len(pkgs) == 0 {
		return nil
	}

	cmdArgs := cmdArgs1.DeepCopy()

	cmdArgs.ModeConf = &settings.DConf{AsDeps: ""}

	cmdArgs.Targets = append(cmdArgs.Targets, pkgs...)
	_, stderr, err := rt.CmdRunner.Capture(passToPacman(rt, cmdArgs), 0)
	if err != nil {
		return fmt.Errorf("%s %s", stderr, err)
	}

	return nil
}

func asexp(cmdArgs *settings.PacmanConf, rt *runtime.Runtime, pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}

	cmdArgs = cmdArgs.DeepCopy()
	cmdArgs.ModeConf = &settings.DConf{
		AsExplicit: "", // TODO
	}
	cmdArgs.Targets = append(cmdArgs.Targets, pkgs...)
	_, stderr, err := rt.CmdRunner.Capture(passToPacman(rt, cmdArgs), 0)
	if err != nil {
		return fmt.Errorf("%s %s", stderr, err)
	}

	return nil
}

// Install handles package installs
func install(rt *runtime.Runtime, pacmanConf *settings.PacmanConf, sconf *settings.SConf, ignoreProviders bool) (err error) {

	var (
		incompatible stringset.StringSet
		do           *dep.Order

		aurUp  upgrade.UpSlice
		repoUp upgrade.UpSlice

		srcinfos map[string]*gosrc.Srcinfo
	)

	warnings := query.NewWarnings()

	if rt.Config.Mode == settings.ModeAny || rt.Config.Mode == settings.ModeRepo {
		if rt.Config.CombinedUpgrade {
			if sconf.Refresh != 0 {
				err = earlyRefresh(pacmanConf, rt)
				if err != nil {
					return text.ErrT("error refreshing databases")
				}
			}
		} else if sconf.Refresh != 0 || sconf.SysUpgrade != 0 || len(rt.Config.Pacman.Targets) > 0 {
			err = earlyPacmanCall(rt, rt.Config.Pacman, sconf)
			if err != nil {
				return err
			}
		}
	}

	// we may have done -Sy, our handle now has an old
	// database.
	err = rt.DB.RefreshHandle()
	if err != nil {
		return err
	}

	localNames, remoteNames, err := query.GetPackageNamesBySource(rt.DB)
	if err != nil {
		return err
	}

	remoteNamesCache := stringset.FromSlice(remoteNames)
	localNamesCache := stringset.FromSlice(localNames)

	requestTargets := pacmanConf.DeepCopy().Targets

	// create the arguments to pass for the repo install
	arguments := pacmanConf.DeepCopy()
	argumentsSConf := arguments.ModeConf.(*settings.SConf)

	argumentsSConf.AsDeps = false
	argumentsSConf.AsExplicit = false
	arguments.Targets = nil

	if rt.Config.Mode == settings.ModeAUR {
		argumentsSConf.SysUpgrade = 0
	}

	// if we are doing -u also request all packages needing update
	if sconf.SysUpgrade != 0 {
		aurUp, repoUp, err = upList(warnings, rt, sconf.SysUpgrade > 1)
		if err != nil {
			return err
		}

		warnings.Print()

		ignore, aurUp, errUp := upgradePkgs(rt.Config, aurUp, repoUp)
		if errUp != nil {
			return errUp
		}

		for _, up := range repoUp {
			if !ignore.Get(up.Name) {
				requestTargets = append(requestTargets, up.Name)
				pacmanConf.Targets = append(pacmanConf.Targets, up.Name)
			}
		}

		for up := range aurUp {
			requestTargets = append(requestTargets, "aur/"+up)
			pacmanConf.Targets = append(pacmanConf.Targets, "aur/"+up)
		}

		if len(ignore) > 0 {
			argumentsSConf.Ignore = append(argumentsSConf.Ignore, ignore.ToSlice()...)
		}
	}

	targets := stringset.FromSlice(pacmanConf.Targets)

	dp, err := dep.GetPool(
		requestTargets, warnings, rt.DB, rt.Config.Mode, ignoreProviders,
		pacmanConf.NoConfirm, rt.Config.Provides, rt.Config.ReBuild,
		rt.Config.RequestSplitN,
	)
	if err != nil {
		return err
	}

	if sconf.NoDeps == 1 {
		err = dp.CheckMissing()
		if err != nil {
			return err
		}
	}

	if len(dp.Aur) == 0 {
		if !rt.Config.CombinedUpgrade {
			if sconf.SysUpgrade != 0 {
				text.Println(text.T(" there is nothing to do"))
			}
			return nil
		}

		pacmanConf.ModeConf = &settings.SConf{}

		if len(argumentsSConf.Ignore) > 0 {
			argumentsSConf.Ignore = append(argumentsSConf.Ignore, argumentsSConf.Ignore...)
		}
		return rt.CmdRunner.Show(passToPacman(rt, pacmanConf))
	}

	if len(dp.Aur) > 0 && os.Geteuid() == 0 {
		return text.ErrT("refusing to install AUR packages as root, aborting")
	}

	var conflicts stringset.MapStringSet
	if sconf.NoDeps == 1 {
		conflicts, err = dp.CheckConflicts(rt.Config.UseAsk, pacmanConf.NoConfirm)
		if err != nil {
			return err
		}
	}

	do = dep.GetOrder(dp)
	if err != nil {
		return err
	}

	for _, pkg := range do.Repo {
		arguments.Targets = append(arguments.Targets, pkg.DB().Name()+"/"+pkg.Name())
	}

	for _, pkg := range dp.Groups {
		arguments.Targets = append(arguments.Targets, pkg)
	}

	if len(do.Aur) == 0 && len(arguments.Targets) == 0 && (sconf.SysUpgrade == 0 || rt.Config.Mode == settings.ModeAUR) {
		text.Println(text.T(" there is nothing to do"))
		return nil
	}

	do.Print()
	text.Println()

	if rt.Config.CleanAfter {
		defer cleanAfter(rt, do.Aur)
	}

	if do.HasMake() {
		switch rt.Config.RemoveMake {
		case "yes":
			defer func() {
				err = removeMake(do, rt)
			}()

		case "no":
			break
		default:
			if text.ContinueTask(text.T("Remove make dependencies after install?"), false, rt.DB.NoConfirm()) {
				defer func() {
					err = removeMake(do, rt)
				}()
			}
		}
	}

	if rt.Config.CleanMenu {
		if anyExistInCache(do.Aur, rt.Config.BuildDir) {
			askClean := pkgbuildNumberMenu(do.Aur, remoteNamesCache, rt.Config.BuildDir)
			toClean, errClean := cleanNumberMenu(do.Aur, remoteNamesCache, askClean, rt.Config.AnswerClean, rt.Config.BuildDir, pacmanConf.NoConfirm)
			if errClean != nil {
				return errClean
			}

			cleanBuilds(rt.Config.BuildDir, toClean)
		}
	}

	toSkip := pkgbuildsToSkip(do.Aur, targets, rt.Config.ReDownload, rt.Config.BuildDir)
	cloned, err := downloadPkgbuilds(BuildRun{rt.CmdBuilder, rt.CmdRunner}, do.Aur, toSkip, rt.Config.BuildDir, rt.Config.AURURL)
	if err != nil {
		return err
	}

	var toDiff []dep.Base
	var toEdit []dep.Base

	if rt.Config.DiffMenu {
		pkgbuildNumberMenu(do.Aur, remoteNamesCache, rt.Config.BuildDir)
		toDiff, err = diffNumberMenu(do.Aur, remoteNamesCache, rt.Config.AnswerDiff, rt.Config.AnswerEdit, pacmanConf.NoConfirm)
		if err != nil {
			return err
		}

		if len(toDiff) > 0 {
			err = showPkgbuildDiffs(BuildRun{rt.CmdBuilder, rt.CmdRunner}, &rt.Config.PersistentYayConfig, toDiff, cloned)
			if err != nil {
				return err
			}
		}
	}

	if len(toDiff) > 0 {
		oldValue := rt.DB.NoConfirm()
		rt.DB.SetNoConfirm(false)
		text.Println()
		if !text.ContinueTask(text.T("Proceed with install?"), true, false) {
			return text.ErrT("aborting due to user")
		}
		err = updatePkgbuildSeenRef(BuildRun{rt.CmdBuilder, rt.CmdRunner}, toDiff, rt.Config.BuildDir)
		if err != nil {
			text.Errorln(err.Error())
		}

		rt.DB.SetNoConfirm(oldValue)
	}

	err = mergePkgbuilds(BuildRun{rt.CmdBuilder, rt.CmdRunner}, do.Aur, rt.Config.BuildDir)
	if err != nil {
		return err
	}

	srcinfos, err = parseSrcinfoFiles(do.Aur, true, rt.Config.BuildDir)
	if err != nil {
		return err
	}

	if rt.Config.EditMenu {
		pkgbuildNumberMenu(do.Aur, remoteNamesCache, rt.Config.BuildDir)
		toEdit, err = editNumberMenu(do.Aur, remoteNamesCache, rt.Config.AnswerDiff, rt.Config.AnswerEdit, pacmanConf.NoConfirm)
		if err != nil {
			return err
		}

		if len(toEdit) > 0 {
			err = editPkgbuilds(toEdit, srcinfos, rt.Config)
			if err != nil {
				return err
			}
		}
	}

	if len(toEdit) > 0 {
		text.Println()
		if !text.ContinueTask(text.T("Proceed with install?"), true, false) {
			return errors.New(text.T("aborting due to user"))
		}
	}

	incompatible, err = getIncompatible(do.Aur, srcinfos, rt.DB, pacmanConf.NoConfirm)
	if err != nil {
		return err
	}

	if rt.Config.PGPFetch {
		err = pgp.CheckPgpKeys(do.Aur, srcinfos, rt.Config.GpgBin, rt.Config.GpgFlags, pacmanConf.NoConfirm)
		if err != nil {
			return err
		}
	}

	if !rt.Config.CombinedUpgrade {
		argumentsSConf.SysUpgrade = 0
	}

	if len(arguments.Targets) > 0 || argumentsSConf.SysUpgrade != 0 {
		if errShow := rt.CmdRunner.Show(passToPacman(rt, arguments)); errShow != nil {
			return errors.New(text.T("error installing repo packages"))
		}

		deps := make([]string, 0)
		exp := make([]string, 0)

		for _, pkg := range do.Repo {
			if !dp.Explicit.Get(pkg.Name()) && !localNamesCache.Get(pkg.Name()) && !remoteNamesCache.Get(pkg.Name()) {
				deps = append(deps, pkg.Name())
				continue
			}

			if sconf.AsDeps && dp.Explicit.Get(pkg.Name()) {
				deps = append(deps, pkg.Name())
			} else if sconf.AsExplicit && dp.Explicit.Get(pkg.Name()) {
				exp = append(exp, pkg.Name())
			}
		}

		if errDeps := asdeps(pacmanConf, rt, deps); errDeps != nil {
			return errDeps
		}
		if errExp := asexp(pacmanConf, rt, exp); errExp != nil {
			return errExp
		}
	}

	go func() {
		_ = completion.Update(rt.DB, rt.Config.AURURL, rt.Config.CompletionPath, rt.Config.CompletionInterval, false)
	}()

	err = downloadPkgbuildsSources(BuildRun{rt.CmdBuilder, rt.CmdRunner}, do.Aur, incompatible, rt.Config.BuildDir)
	if err != nil {
		return err
	}

	var upgr *settings.Upgrade
	switch t := pacmanConf.ModeConf.(type) {
	case *settings.UConf:
		upgr = &t.Upgrade
	case *settings.SConf:
		upgr = &t.Upgrade
	}
	err = buildInstallPkgbuilds(rt, pacmanConf, upgr, dp, do, srcinfos, incompatible, conflicts)
	if err != nil {
		return err
	}

	return nil
}

func removeMake(do *dep.Order, rt *runtime.Runtime) error {
	removeArguments := &settings.PacmanConf{
		ModeConf: &settings.RConf{
			Unneeded: true,
		},
	}
	for _, pkg := range do.GetMake() {
		removeArguments.Targets = append(removeArguments.Targets, pkg)
	}

	oldValue := rt.DB.NoConfirm()
	rt.DB.SetNoConfirm(true)
	err := rt.CmdRunner.Show(passToPacman(rt, removeArguments))
	rt.DB.SetNoConfirm(oldValue)

	return err
}

func inRepos(dbExecutor db.Executor, pkg string) bool {
	target := dep.ToTarget(pkg)

	if target.DB == "aur" {
		return false
	} else if target.DB != "" {
		return true
	}

	previousHideMenus := dbExecutor.HideMenus()
	dbExecutor.SetHideMenus(false)
	exists := dbExecutor.SyncSatisfierExists(target.DepString())
	dbExecutor.SetHideMenus(previousHideMenus)

	return exists || len(dbExecutor.PackagesFromGroup(target.Name)) > 0
}

func earlyPacmanCall(rt *runtime.Runtime, pacmanConf *settings.PacmanConf, sconf *settings.SConf) error {

	arguments := pacmanConf.DeepCopy()
	arguments.ModeConf = &settings.SConf{}

	targets := pacmanConf.Targets
	pacmanConf.Targets = nil
	arguments.Targets = nil

	if rt.Config.Mode == settings.ModeRepo {
		arguments.Targets = targets
	} else {
		// separate aur and repo targets
		for _, target := range targets {
			if inRepos(rt.DB, target) {
				arguments.Targets = append(arguments.Targets, target)
			} else {
				pacmanConf.Targets = append(pacmanConf.Targets, target)
			}
		}
	}

	if sconf.Refresh != 0 || sconf.SysUpgrade != 0 || len(arguments.Targets) > 0 {
		if err := rt.CmdRunner.Show(passToPacman(rt, arguments)); err != nil {
			return errors.New(text.T("error installing repo packages"))
		}
	}

	return nil
}

func earlyRefresh(cmdArgs *settings.PacmanConf, rt *runtime.Runtime) error {

	cmdArgs.ModeConf.(*settings.SConf).Refresh = 0

	arguments := cmdArgs.DeepCopy()
	arguments.ModeConf = &settings.SConf{ // TODO
		SysUpgrade: 0,
		Search:     "",
		Info:       0,
		List:       false,
	}
	arguments.Targets = []string{}

	return rt.CmdRunner.Show(passToPacman(rt, arguments))
}

func getIncompatible(bases []dep.Base, srcinfos map[string]*gosrc.Srcinfo, dbExecutor db.Executor, noConfirm bool) (stringset.StringSet, error) {
	incompatible := make(stringset.StringSet)
	basesMap := make(map[string]dep.Base)
	alpmArch, err := dbExecutor.AlpmArch()
	if err != nil {
		return nil, err
	}

nextpkg:
	for _, base := range bases {
		for _, arch := range srcinfos[base.Pkgbase()].Arch {
			if arch == "any" || arch == alpmArch {
				continue nextpkg
			}
		}

		incompatible.Set(base.Pkgbase())
		basesMap[base.Pkgbase()] = base
	}

	if len(incompatible) > 0 {
		text.Warnln(text.T("The following packages are not compatible with your architecture:"))
		for pkg := range incompatible {
			text.Print("  " + text.Cyan(basesMap[pkg].String()))
		}

		text.Println()

		if !text.ContinueTask(text.T("Try to build them anyway?"), true, noConfirm) {
			return nil, errors.New(text.T("aborting due to user"))
		}
	}

	return incompatible, nil
}

func parsePackageList(dir string, br BuildRun) (pkgdests map[string]string, pkgVersion string, err error) {

	stdout, stderr, err := br.Run.Capture(br.Build.BuildMakepkgCmd(dir, "--packagelist"), 0)
	if err != nil {
		return nil, "", fmt.Errorf("%s %s", stderr, err)
	}

	lines := strings.Split(stdout, "\n")
	pkgdests = make(map[string]string)

	for _, line := range lines {
		if line == "" {
			continue
		}

		fileName := filepath.Base(line)
		split := strings.Split(fileName, "-")

		if len(split) < 4 {
			return nil, "", errors.New(text.Tf("cannot find package name: %v", split))
		}

		// pkgname-pkgver-pkgrel-arch.pkgext
		// This assumes 3 dashes after the pkgname, Will cause an error
		// if the PKGEXT contains a dash. Please no one do that.
		pkgName := strings.Join(split[:len(split)-3], "-")
		pkgVersion = strings.Join(split[len(split)-3:len(split)-1], "-")
		pkgdests[pkgName] = line
	}

	return pkgdests, pkgVersion, nil
}

func anyExistInCache(bases []dep.Base, buildDir string) bool {
	for _, base := range bases {
		pkg := base.Pkgbase()
		dir := filepath.Join(buildDir, pkg)

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			return true
		}
	}

	return false
}

func pkgbuildNumberMenu(bases []dep.Base, installed stringset.StringSet, buildDir string) bool {
	toPrint := ""
	askClean := false

	for n, base := range bases {
		pkg := base.Pkgbase()
		dir := filepath.Join(buildDir, pkg)

		toPrint += fmt.Sprintf(text.Magenta("%3d")+" %-40s", len(bases)-n,
			text.Bold(base.String()))

		anyInstalled := false
		for _, b := range base {
			anyInstalled = anyInstalled || installed.Get(b.Name)
		}

		if anyInstalled {
			toPrint += text.Bold(text.Green(text.T(" (Installed)")))
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			toPrint += text.Bold(text.Green(text.T(" (Build Files Exist)")))
			askClean = true
		}

		toPrint += "\n"
	}

	text.Print(toPrint)

	return askClean
}

func cleanNumberMenu(bases []dep.Base, installed stringset.StringSet, hasClean bool, answerClean, buildDir string, noConfirm bool) ([]dep.Base, error) {
	toClean := make([]dep.Base, 0)

	if !hasClean {
		return toClean, nil
	}

	text.Infoln(text.T("Packages to cleanBuild?"))
	text.Infoln(text.Tf("%s [A]ll [Ab]ort [I]nstalled [No]tInstalled or (1 2 3, 1-3, ^4)", text.Cyan(text.T("[N]one"))))
	cleanInput, err := getInput(answerClean, noConfirm)
	if err != nil {
		return nil, err
	}

	cInclude, cExclude, cOtherInclude, cOtherExclude := ParseNumberMenu(cleanInput)
	cIsInclude := len(cExclude) == 0 && len(cOtherExclude) == 0

	if cOtherInclude.Get("abort") || cOtherInclude.Get("ab") {
		return nil, text.ErrT("aborting due to user")
	}

	if !cOtherInclude.Get("n") && !cOtherInclude.Get("none") {
		for i, base := range bases {
			pkg := base.Pkgbase()
			anyInstalled := false
			for _, b := range base {
				anyInstalled = anyInstalled || installed.Get(b.Name)
			}

			dir := filepath.Join(buildDir, pkg)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				continue
			}

			if !cIsInclude && cExclude.Get(len(bases)-i) {
				continue
			}

			if anyInstalled && (cOtherInclude.Get("i") || cOtherInclude.Get("installed")) {
				toClean = append(toClean, base)
				continue
			}

			if !anyInstalled && (cOtherInclude.Get("no") || cOtherInclude.Get("notinstalled")) {
				toClean = append(toClean, base)
				continue
			}

			if cOtherInclude.Get("a") || cOtherInclude.Get("all") {
				toClean = append(toClean, base)
				continue
			}

			if cIsInclude && (cInclude.Get(len(bases)-i) || cOtherInclude.Get(pkg)) {
				toClean = append(toClean, base)
				continue
			}

			if !cIsInclude && (!cExclude.Get(len(bases)-i) && !cOtherExclude.Get(pkg)) {
				toClean = append(toClean, base)
				continue
			}
		}
	}

	return toClean, nil
}

func editNumberMenu(bases []dep.Base, installed stringset.StringSet, answerDiff, answerEdit string, noConfirm bool) ([]dep.Base, error) {
	return editDiffNumberMenu(bases, installed, false, answerDiff, answerEdit, noConfirm)
}

func diffNumberMenu(bases []dep.Base, installed stringset.StringSet, answerDiff, answerEdit string, noConfirm bool) ([]dep.Base, error) {
	return editDiffNumberMenu(bases, installed, true, answerDiff, answerEdit, noConfirm)
}

func editDiffNumberMenu(bases []dep.Base, installed stringset.StringSet, diff bool, answerDiff, answerEdit string, noConfirm bool) ([]dep.Base, error) {
	toEdit := make([]dep.Base, 0)
	var editInput string
	var err error

	if diff {
		text.Infoln(text.T("Diffs to show?"))
		text.Infoln(text.Tf("%s [A]ll [Ab]ort [I]nstalled [No]tInstalled or (1 2 3, 1-3, ^4)", text.Cyan(text.T("[N]one"))))
		editInput, err = getInput(answerDiff, noConfirm)
		if err != nil {
			return nil, err
		}
	} else {
		text.Infoln(text.T("PKGBUILDs to edit?"))
		text.Infoln(text.Tf("%s [A]ll [Ab]ort [I]nstalled [No]tInstalled or (1 2 3, 1-3, ^4)", text.Cyan(text.T("[N]one"))))
		editInput, err = getInput(answerEdit, noConfirm)
		if err != nil {
			return nil, err
		}
	}

	eInclude, eExclude, eOtherInclude, eOtherExclude := ParseNumberMenu(editInput)
	eIsInclude := len(eExclude) == 0 && len(eOtherExclude) == 0

	if eOtherInclude.Get("abort") || eOtherInclude.Get("ab") {
		return nil, text.ErrT("aborting due to user")
	}

	if !eOtherInclude.Get("n") && !eOtherInclude.Get("none") {
		for i, base := range bases {
			pkg := base.Pkgbase()
			anyInstalled := false
			for _, b := range base {
				anyInstalled = anyInstalled || installed.Get(b.Name)
			}

			if !eIsInclude && eExclude.Get(len(bases)-i) {
				continue
			}

			if anyInstalled && (eOtherInclude.Get("i") || eOtherInclude.Get("installed")) {
				toEdit = append(toEdit, base)
				continue
			}

			if !anyInstalled && (eOtherInclude.Get("no") || eOtherInclude.Get("notinstalled")) {
				toEdit = append(toEdit, base)
				continue
			}

			if eOtherInclude.Get("a") || eOtherInclude.Get("all") {
				toEdit = append(toEdit, base)
				continue
			}

			if eIsInclude && (eInclude.Get(len(bases)-i) || eOtherInclude.Get(pkg)) {
				toEdit = append(toEdit, base)
			}

			if !eIsInclude && (!eExclude.Get(len(bases)-i) && !eOtherExclude.Get(pkg)) {
				toEdit = append(toEdit, base)
			}
		}
	}

	return toEdit, nil
}

func updatePkgbuildSeenRef(br BuildRun, bases []dep.Base, buildDir string) error {
	var errMulti multierror.MultiError
	for _, base := range bases {
		pkg := base.Pkgbase()
		err := gitUpdateSeenRef(br, buildDir, pkg)
		if err != nil {
			errMulti.Add(err)
		}
	}
	return errMulti.Return()
}

func showPkgbuildDiffs(br BuildRun, conf *settings.PersistentYayConfig, bases []dep.Base, cloned stringset.StringSet) error {
	var errMulti multierror.MultiError
	for _, base := range bases {
		pkg := base.Pkgbase()
		dir := filepath.Join(conf.BuildDir, pkg)
		start, err := getLastSeenHash(br, conf.BuildDir, pkg)
		if err != nil {
			errMulti.Add(err)
			continue
		}

		if cloned.Get(pkg) {
			start = gitEmptyTree
		} else {
			hasDiff, err := gitHasDiff(br, conf.BuildDir, pkg)
			if err != nil {
				errMulti.Add(err)
				continue
			}

			if !hasDiff {
				text.Warnln(text.Tf("%s: No changes -- skipping", text.Cyan(base.String())))
				continue
			}
		}

		args := []string{
			"diff",
			start + "..HEAD@{upstream}", "--src-prefix",
			dir + "/", "--dst-prefix", dir + "/", "--", ".", ":(exclude).SRCINFO",
		}
		if text.UseColor {
			args = append(args, "--color=always")
		} else {
			args = append(args, "--color=never")
		}
		_ = br.Run.Show(br.Build.BuildGitCmd(dir, args...))
	}

	return errMulti.Return()
}

func editPkgbuilds(bases []dep.Base, srcinfos map[string]*gosrc.Srcinfo, conf *settings.YayConfig) error {
	pkgbuilds := make([]string, 0, len(bases))
	for _, base := range bases {
		pkg := base.Pkgbase()
		dir := filepath.Join(conf.BuildDir, pkg)
		pkgbuilds = append(pkgbuilds, filepath.Join(dir, "PKGBUILD"))

		for _, splitPkg := range srcinfos[pkg].SplitPackages() {
			if splitPkg.Install != "" {
				pkgbuilds = append(pkgbuilds, filepath.Join(dir, splitPkg.Install))
			}
		}
	}

	if len(pkgbuilds) > 0 {
		editor, editorArgs := editor(conf.Editor, conf.EditorFlags, conf.Pacman.NoConfirm)
		editorArgs = append(editorArgs, pkgbuilds...)
		editcmd := exec.Command(editor, editorArgs...)
		editcmd.Stdin, editcmd.Stdout, editcmd.Stderr = text.AllPorts()
		err := editcmd.Run()
		if err != nil {
			return errors.New(text.Tf("editor did not exit successfully, aborting: %s", err))
		}
	}

	return nil
}

func parseSrcinfoFiles(bases []dep.Base, errIsFatal bool, buildDir string) (map[string]*gosrc.Srcinfo, error) {
	srcinfos := make(map[string]*gosrc.Srcinfo)
	for k, base := range bases {
		pkg := base.Pkgbase()
		dir := filepath.Join(buildDir, pkg)

		text.OperationInfoln(text.Tf("(%d/%d) Parsing SRCINFO: %s", k+1, len(bases), text.Cyan(base.String())))

		pkgbuild, err := gosrc.ParseFile(filepath.Join(dir, ".SRCINFO"))
		if err != nil {
			if !errIsFatal {
				text.Warnln(text.Tf("failed to parse %s -- skipping: %s", base.String(), err))
				continue
			}
			return nil, errors.New(text.Tf("failed to parse %s: %s", base.String(), err))
		}

		srcinfos[pkg] = pkgbuild
	}

	return srcinfos, nil
}

func pkgbuildsToSkip(bases []dep.Base, targets stringset.StringSet, reDownload, buildDir string) stringset.StringSet {
	toSkip := make(stringset.StringSet)

	for _, base := range bases {
		isTarget := false
		for _, pkg := range base {
			isTarget = isTarget || targets.Get(pkg.Name)
		}

		if (reDownload == "yes" && isTarget) || reDownload == "all" {
			continue
		}

		dir := filepath.Join(buildDir, base.Pkgbase(), ".SRCINFO")
		pkgbuild, err := gosrc.ParseFile(dir)

		if err == nil {
			if alpm.VerCmp(pkgbuild.Version(), base.Version()) >= 0 {
				toSkip.Set(base.Pkgbase())
			}
		}
	}

	return toSkip
}

func mergePkgbuilds(br BuildRun, bases []dep.Base, buildDir string) error {
	for _, base := range bases {
		err := gitMerge(br, buildDir, base.Pkgbase())
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadPkgbuilds(br BuildRun, bases []dep.Base, toSkip stringset.StringSet, buildDir, aurURL string) (stringset.StringSet, error) {
	cloned := make(stringset.StringSet)
	downloaded := 0
	var wg sync.WaitGroup
	var mux sync.Mutex
	var errs multierror.MultiError

	download := func(base dep.Base) {
		defer wg.Done()
		pkg := base.Pkgbase()

		if toSkip.Get(pkg) {
			mux.Lock()
			downloaded++
			text.OperationInfoln(
				text.Tf("PKGBUILD up to date, Skipping (%d/%d): %s",
					downloaded, len(bases), text.Cyan(base.String())))
			mux.Unlock()
			return
		}

		clone, err := gitDownload(br, aurURL+"/"+pkg+".git", buildDir, pkg)
		if err != nil {
			errs.Add(err)
			return
		}
		if clone {
			mux.Lock()
			cloned.Set(pkg)
			mux.Unlock()
		}

		mux.Lock()
		downloaded++
		text.OperationInfoln(text.Tf("Downloaded PKGBUILD (%d/%d): %s", downloaded, len(bases), text.Cyan(base.String())))
		mux.Unlock()
	}

	count := 0
	for _, base := range bases {
		wg.Add(1)
		go download(base)
		count++
		if count%25 == 0 {
			wg.Wait()
		}
	}

	wg.Wait()

	return cloned, errs.Return()
}

func downloadPkgbuildsSources(br BuildRun, bases []dep.Base, incompatible stringset.StringSet, buildDir string) (err error) {
	for _, base := range bases {
		pkg := base.Pkgbase()
		dir := filepath.Join(buildDir, pkg)
		args := []string{"--verifysource", "-Ccf"}

		if incompatible.Get(pkg) {
			args = append(args, "--ignorearch")
		}

		err = br.Run.Show(br.Build.BuildMakepkgCmd(dir, args...))
		if err != nil {
			return errors.New(text.Tf("error downloading sources: %s", text.Cyan(base.String())))
		}
	}

	return
}

func buildInstallPkgbuilds(
	rt *runtime.Runtime,
	cmdArgs *settings.PacmanConf,
	pacmanUpgrade *settings.Upgrade,
	dp *dep.Pool,
	do *dep.Order,
	srcinfos map[string]*gosrc.Srcinfo,
	incompatible stringset.StringSet,
	conflicts stringset.MapStringSet,
) error {

	var trans settings.Transaction
	var upgr settings.Upgrade
	switch t := cmdArgs.ModeConf.(type) {
	case *settings.UConf:
		trans = t.Transaction
		upgr = t.Upgrade
	case *settings.SConf:
		trans = t.Transaction
		upgr = t.Upgrade
	default:
		return fmt.Errorf("invalid pacman main mode of operation")
	}

	arguments := cmdArgs.DeepCopy()
	arguments.Targets = nil
	arguments.ModeConf = &settings.UConf{
		Transaction: trans,
		Upgrade:     upgr,
	}

	arguments.NoConfirm = false

	deps := make([]string, 0)
	exp := make([]string, 0)

	oldConfirm := rt.DB.NoConfirm()
	rt.DB.SetNoConfirm(true)

	//remotenames: names of all non repo packages on the system
	localNames, remoteNames, err := query.GetPackageNamesBySource(rt.DB)
	if err != nil {
		return err
	}

	// cache as a stringset. maybe make it return a string set in the first
	// place
	remoteNamesCache := stringset.FromSlice(remoteNames)
	localNamesCache := stringset.FromSlice(localNames)

	doInstall := func() error {
		if len(arguments.Targets) == 0 {
			return nil
		}

		if errShow := rt.CmdRunner.Show(passToPacman(rt, arguments)); errShow != nil {
			return errShow
		}

		if errStore := rt.VCSStore.Save(); err != nil {
			text.EPrintln(errStore)
		}

		if errDeps := asdeps(cmdArgs, rt, deps); err != nil {
			return errDeps
		}
		if errExps := asexp(cmdArgs, rt, exp); err != nil {
			return errExps
		}

		rt.DB.SetNoConfirm(oldConfirm)

		arguments.Targets = []string{}
		deps = make([]string, 0)
		exp = make([]string, 0)
		rt.DB.SetNoConfirm(true)
		return nil
	}

	for _, base := range do.Aur {
		pkg := base.Pkgbase()
		dir := filepath.Join(rt.Config.BuildDir, pkg)
		built := true

		satisfied := true
	all:
		for _, pkg := range base {
			for _, deps := range [3][]string{pkg.Depends, pkg.MakeDepends, pkg.CheckDepends} {
				for _, dep := range deps {
					if !dp.AlpmExecutor.LocalSatisfierExists(dep) {
						satisfied = false
						text.Warnln(text.Tf("%s not satisfied, flushing install queue", dep))
						break all
					}
				}
			}
		}

		if !satisfied || !rt.Config.BatchInstall {
			err = doInstall()
			if err != nil {
				return err
			}
		}

		srcinfo := srcinfos[pkg]

		args := []string{"--nobuild", "-fC"}

		if incompatible.Get(pkg) {
			args = append(args, "--ignorearch")
		}

		// pkgver bump
		if err = rt.CmdRunner.Show(rt.CmdBuilder.BuildMakepkgCmd(dir, args...)); err != nil {
			return errors.New(text.Tf("error making: %s", base.String()))
		}

		pkgdests, pkgVersion, errList := parsePackageList(dir, BuildRun{rt.CmdBuilder, rt.CmdRunner})
		if errList != nil {
			return errList
		}

		isExplicit := false
		for _, b := range base {
			isExplicit = isExplicit || dp.Explicit.Get(b.Name)
		}
		if rt.Config.ReBuild == "no" || (rt.Config.ReBuild == "yes" && !isExplicit) {
			for _, split := range base {
				pkgdest, ok := pkgdests[split.Name]
				if !ok {
					return errors.New(text.Tf("could not find PKGDEST for: %s", split.Name))
				}

				if _, errStat := os.Stat(pkgdest); os.IsNotExist(errStat) {
					built = false
				} else if errStat != nil {
					return errStat
				}
			}
		} else {
			built = false
		}

		if pacmanUpgrade.Needed {
			installed := true
			for _, split := range base {
				installed = dp.AlpmExecutor.IsCorrectVersionInstalled(split.Name, pkgVersion)
			}

			if installed {
				err = rt.CmdRunner.Show(
					rt.CmdBuilder.BuildMakepkgCmd(
						dir, "-c", "--nobuild", "--noextract", "--ignorearch"))
				if err != nil {
					return errors.New(text.Tf("error making: %s", err))
				}

				text.EPrintln(text.Tf("%s is up to date -- skipping", text.Cyan(pkg+"-"+pkgVersion)))
				continue
			}
		}

		if built {
			err = rt.CmdRunner.Show(
				rt.CmdBuilder.BuildMakepkgCmd(
					dir, "-c", "--nobuild", "--noextract", "--ignorearch"))
			if err != nil {
				return errors.New(text.Tf("error making: %s", err))
			}

			text.Warnln(text.Tf("%s already made -- skipping build", text.Cyan(pkg+"-"+pkgVersion)))
		} else {
			args := []string{"-cf", "--noconfirm", "--noextract", "--noprepare", "--holdver"}

			if incompatible.Get(pkg) {
				args = append(args, "--ignorearch")
			}

			if errMake := rt.CmdRunner.Show(rt.CmdBuilder.BuildMakepkgCmd(dir, args...)); errMake != nil {
				return errors.New(text.Tf("error making: %s", base.String()))
			}
		}

		// conflicts have been checked so answer y for them
		if rt.Config.UseAsk && rt.Config.Pacman.Ask != 0 {

			rt.Config.Pacman.Ask = int(alpm.QuestionType(rt.Config.Pacman.Ask) | alpm.QuestionTypeConflictPkg)

		} else {
			for _, split := range base {
				if _, ok := conflicts[split.Name]; ok {
					rt.DB.SetNoConfirm(false)
					break
				}
			}
		}

		doAddTarget := func(name string, optional bool) error {
			pkgdest, ok := pkgdests[name]
			if !ok {
				if optional {
					return nil
				}

				return errors.New(text.Tf("could not find PKGDEST for: %s", name))
			}

			if _, errStat := os.Stat(pkgdest); os.IsNotExist(errStat) {
				if optional {
					return nil
				}

				return errors.New(
					text.Tf(
						"the PKGDEST for %s is listed by makepkg but does not exist: %s",
						name, pkgdest))
			}

			arguments.Targets = append(arguments.Targets, pkgdest)
			if pacmanUpgrade.AsDeps {
				deps = append(deps, name)
			} else if pacmanUpgrade.AsExplicit {
				exp = append(exp, name)
			} else if !dp.Explicit.Get(name) && !localNamesCache.Get(name) && !remoteNamesCache.Get(name) {
				deps = append(deps, name)
			}

			return nil
		}

		for _, split := range base {
			if errAdd := doAddTarget(split.Name, false); errAdd != nil {
				return errAdd
			}

			if errAddDebug := doAddTarget(split.Name+"-debug", true); errAddDebug != nil {
				return errAddDebug
			}
		}

		var mux sync.Mutex
		var wg sync.WaitGroup
		for _, pkg := range base {
			wg.Add(1)
			go rt.VCSStore.Update(pkg.Name, srcinfo.Source, &mux, &wg)
		}

		wg.Wait()
	}

	err = doInstall()
	rt.DB.SetNoConfirm(oldConfirm)
	return err
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/dep"
	"github.com/Jguer/yay/v10/pkg/multierror"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings/exe"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/text"
)

const gitDiffRefName = "AUR_SEEN"

type BuildRun struct {
	Build *exe.CmdBuilder
	Run   exe.Runner
}

// Update the YAY_DIFF_REVIEW ref to HEAD. We use this ref to determine which diff were
// reviewed by the user
func gitUpdateSeenRef(br BuildRun, path, name string) error {
	_, stderr, err := br.Run.Capture(
		br.Build.BuildGitCmd(
			filepath.Join(path, name), "update-ref", gitDiffRefName, "HEAD"), 0)
	if err != nil {
		return fmt.Errorf("%s %s", stderr, err)
	}
	return nil
}

// Return wether or not we have reviewed a diff yet. It checks for the existence of
// YAY_DIFF_REVIEW in the git ref-list
func gitHasLastSeenRef(br BuildRun, path, name string) bool {
	_, _, err := br.Run.Capture(
		br.Build.BuildGitCmd(
			filepath.Join(path, name), "rev-parse", "--quiet", "--verify", gitDiffRefName), 0)
	return err == nil
}

// Returns the last reviewed hash. If YAY_DIFF_REVIEW exists it will return this hash.
// If it does not it will return empty tree as no diff have been reviewed yet.
func getLastSeenHash(br BuildRun, path, name string) (string, error) {
	if gitHasLastSeenRef(br, path, name) {
		stdout, stderr, err := br.Run.Capture(
			br.Build.BuildGitCmd(
				filepath.Join(path, name), "rev-parse", gitDiffRefName), 0)
		if err != nil {
			return "", fmt.Errorf("%s %s", stderr, err)
		}

		lines := strings.Split(stdout, "\n")
		return lines[0], nil
	}
	return gitEmptyTree, nil
}

// Check whether or not a diff exists between the last reviewed diff and
// HEAD@{upstream}
func gitHasDiff(br BuildRun, path, name string) (bool, error) {
	if gitHasLastSeenRef(br, path, name) {
		stdout, stderr, err := br.Run.Capture(
			br.Build.BuildGitCmd(filepath.Join(path, name), "rev-parse", gitDiffRefName, "HEAD@{upstream}"), 0)
		if err != nil {
			return false, fmt.Errorf("%s%s", stderr, err)
		}

		lines := strings.Split(stdout, "\n")
		lastseen := lines[0]
		upstream := lines[1]
		return lastseen != upstream, nil
	}
	// If YAY_DIFF_REVIEW does not exists, we have never reviewed a diff for this package
	// and should display it.
	return true, nil
}

// TODO: yay-next passes args through the header, use that to unify ABS and AUR
func gitDownloadABS(br BuildRun, url, path, name string) (bool, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, err
	}

	if _, errExist := os.Stat(filepath.Join(path, name)); os.IsNotExist(errExist) {
		cmd := br.Build.BuildGitCmd(path, "clone", "--no-progress", "--single-branch",
			"-b", "packages/"+name, url, name)
		_, stderr, err := br.Run.Capture(cmd, 0)
		if err != nil {
			return false, fmt.Errorf(text.Tf("error cloning %s: %s", name, stderr))
		}

		return true, nil
	} else if errExist != nil {
		return false, fmt.Errorf(text.Tf("error reading %s", filepath.Join(path, name, ".git")))
	}

	cmd := br.Build.BuildGitCmd(filepath.Join(path, name), "pull", "--ff-only")
	_, stderr, err := br.Run.Capture(cmd, 0)
	if err != nil {
		return false, fmt.Errorf(text.Tf("error fetching %s: %s", name, stderr))
	}

	return true, nil
}

func gitDownload(br BuildRun, url, path, name string) (bool, error) {
	_, err := os.Stat(filepath.Join(path, name, ".git"))
	if os.IsNotExist(err) {
		cmd := br.Build.BuildGitCmd(path, "clone", "--no-progress", url, name)
		_, stderr, errCapture := br.Run.Capture(cmd, 0)
		if errCapture != nil {
			return false, fmt.Errorf(text.Tf("error cloning %s: %s", name, stderr))
		}

		return true, nil
	} else if err != nil {
		return false, fmt.Errorf(text.Tf("error reading %s", filepath.Join(path, name, ".git")))
	}

	cmd := br.Build.BuildGitCmd(filepath.Join(path, name), "fetch")
	_, stderr, err := br.Run.Capture(cmd, 0)
	if err != nil {
		return false, fmt.Errorf(text.Tf("error fetching %s: %s", name, stderr))
	}

	return false, nil
}

func gitMerge(br BuildRun, path, name string) error {
	_, stderr, err := br.Run.Capture(
		br.Build.BuildGitCmd(
			filepath.Join(path, name), "reset", "--hard", "HEAD"), 0)
	if err != nil {
		return fmt.Errorf(text.Tf("error resetting %s: %s", name, stderr))
	}

	_, stderr, err = br.Run.Capture(
		br.Build.BuildGitCmd(
			filepath.Join(path, name), "merge", "--no-edit", "--ff"), 0)
	if err != nil {
		return fmt.Errorf(text.Tf("error merging %s: %s", name, stderr))
	}

	return nil
}

func getPkgbuilds(pkgs []string, rt *runtime.Runtime, force bool) error {
	missing := false
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	pkgs = query.RemoveInvalidTargets(pkgs, rt.Config.Mode)
	aur, repo := packageSlices(pkgs, rt.DB, rt.Config.Mode)

	for n := range aur {
		_, pkg := text.SplitDBFromName(aur[n])
		aur[n] = pkg
	}

	info, err := query.AURInfoPrint(aur, rt.Config.Conf.RequestSplitN)
	if err != nil {
		return err
	}

	if len(repo) > 0 {
		missing, err = getPkgbuildsfromABS(repo, wd, force, rt)
		if err != nil {
			return err
		}
	}

	if len(aur) > 0 {
		allBases := dep.GetBases(info)
		bases := make([]dep.Base, 0)

		for _, base := range allBases {
			name := base.Pkgbase()
			pkgDest := filepath.Join(wd, name)
			_, err = os.Stat(pkgDest)
			if os.IsNotExist(err) {
				bases = append(bases, base)
			} else if err != nil {
				text.Errorln(err)
				continue
			} else {
				if force {
					if err = os.RemoveAll(pkgDest); err != nil {
						text.Errorln(err)
						continue
					}
					bases = append(bases, base)
				} else {
					text.Warnln(text.Tf("%s already exists. Use -f/--force to overwrite", pkgDest))
					continue
				}
			}
		}

		if _, err = downloadPkgbuilds(BuildRun{rt.CmdBuilder, rt.CmdRunner}, bases, nil, wd, rt.Config.Conf.AURURL); err != nil {
			return err
		}

		missing = missing || len(aur) != len(info)
	}

	if missing {
		err = ErrMissing
	}

	return err
}

// GetPkgbuild downloads pkgbuild from the ABS.
func getPkgbuildsfromABS(pkgs []string, path string, force bool, rt *runtime.Runtime) (bool, error) {
	var wg sync.WaitGroup
	var mux sync.Mutex
	var errs multierror.MultiError
	names := make(map[string]string)
	missing := make([]string, 0)
	downloaded := 0

	for _, pkgN := range pkgs {
		var pkg db.IPackage
		var err error
		var url string
		pkgDB, name := text.SplitDBFromName(pkgN)

		if pkgDB != "" {
			pkg = rt.DB.SatisfierFromDB(name, pkgDB)
		} else {
			pkg = rt.DB.SyncSatisfier(name)
		}

		if pkg == nil {
			missing = append(missing, name)
			continue
		}

		name = pkg.Base()
		if name == "" {
			name = pkg.Name()
		}

		// TODO: Check existence with ls-remote
		// https://git.archlinux.org/svntogit/packages.git
		switch pkg.DB().Name() {
		case "core", "extra", "testing":
			url = "https://git.archlinux.org/svntogit/packages.git"
		case "community", "multilib", "community-testing", "multilib-testing":
			url = "https://git.archlinux.org/svntogit/community.git"
		default:
			missing = append(missing, name)
			continue
		}

		_, err = os.Stat(filepath.Join(path, name))
		switch {
		case err != nil && !os.IsNotExist(err):
			text.Errorln(err)
			continue
		case os.IsNotExist(err), force:
			if err = os.RemoveAll(filepath.Join(path, name)); err != nil {
				text.Errorln(err)
				continue
			}
		default:
			text.Warn(text.Tf("%s already downloaded -- use -f to overwrite", text.Cyan(name)))
			continue
		}

		names[name] = url
	}

	if len(missing) != 0 {
		text.Warnln(text.T("Missing ABS packages:"),
			text.Cyan(strings.Join(missing, ", ")))
	}

	download := func(pkg string, url string) {
		defer wg.Done()
		if _, err := gitDownloadABS(BuildRun{rt.CmdBuilder, rt.CmdRunner}, url, rt.Config.Conf.ABSDir, pkg); err != nil {
			errs.Add(errors.New(text.Tf("failed to get pkgbuild: %s: %s", text.Cyan(pkg), err.Error())))
			return
		}

		_, stderr, err := rt.CmdRunner.Capture(
			exec.Command(
				"cp", "-r",
				filepath.Join(rt.Config.Conf.ABSDir, pkg, "trunk"),
				filepath.Join(path, pkg)), 0)
		mux.Lock()
		downloaded++
		if err != nil {
			errs.Add(errors.New(text.Tf("failed to link %s: %s", text.Cyan(pkg), stderr)))
		} else {
			text.EPrintln(text.Tf("(%d/%d) Downloaded PKGBUILD from ABS: %s", downloaded, len(names), text.Cyan(pkg)))
		}
		mux.Unlock()
	}

	count := 0
	for name, url := range names {
		wg.Add(1)
		go download(name, url)
		count++
		if count%25 == 0 {
			wg.Wait()
		}
	}

	wg.Wait()

	return len(missing) != 0, errs.Return()
}

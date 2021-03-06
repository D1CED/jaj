package dep

import (
	"errors"
	"strings"
	"sync"

	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
)

type mapStringSet = map[string]stringset.StringSet

func (dp *Pool) checkInnerConflict(name, conflict string, conflicts mapStringSet) {
	for _, pkg := range dp.Aur {
		if pkg.Name == name {
			continue
		}

		if satisfiesAur(conflict, pkg) {
			stringset.Add(conflicts, name, pkg.Name)
		}
	}

	for _, pkg := range dp.repo {
		if pkg.Name() == name {
			continue
		}

		if satisfiesRepo(conflict, pkg, dp.alpmExecutor) {
			stringset.Add(conflicts, name, pkg.Name())
		}
	}
}

func (dp *Pool) checkForwardConflict(name, conflict string, conflicts mapStringSet) {
	for _, pkg := range dp.alpmExecutor.LocalPackages() {
		if pkg.Name() == name || dp.hasPackage(pkg.Name()) {
			continue
		}

		if satisfiesRepo(conflict, pkg, dp.alpmExecutor) {
			n := pkg.Name()
			if n != conflict {
				n += " (" + conflict + ")"
			}
			stringset.Add(conflicts, name, n)
		}
	}
}

func (dp *Pool) checkReverseConflict(name, conflict string, conflicts mapStringSet) {
	for _, pkg := range dp.Aur {
		if pkg.Name == name {
			continue
		}

		if satisfiesAur(conflict, pkg) {
			if name != conflict {
				name += " (" + conflict + ")"
			}

			stringset.Add(conflicts, pkg.Name, name)
		}
	}

	for _, pkg := range dp.repo {
		if pkg.Name() == name {
			continue
		}

		if satisfiesRepo(conflict, pkg, dp.alpmExecutor) {
			if name != conflict {
				name += " (" + conflict + ")"
			}

			stringset.Add(conflicts, pkg.Name(), name)
		}
	}
}

func (dp *Pool) checkInnerConflicts(conflicts mapStringSet) {
	for _, pkg := range dp.Aur {
		for _, conflict := range pkg.Conflicts {
			dp.checkInnerConflict(pkg.Name, conflict, conflicts)
		}
	}

	for _, pkg := range dp.repo {
		for _, conflict := range dp.alpmExecutor.PackageConflicts(pkg) {
			dp.checkInnerConflict(pkg.Name(), conflict.String(), conflicts)
		}
	}
}

func (dp *Pool) checkForwardConflicts(conflicts mapStringSet) {
	for _, pkg := range dp.Aur {
		for _, conflict := range pkg.Conflicts {
			dp.checkForwardConflict(pkg.Name, conflict, conflicts)
		}
	}

	for _, pkg := range dp.repo {
		for _, conflict := range dp.alpmExecutor.PackageConflicts(pkg) {
			dp.checkForwardConflict(pkg.Name(), conflict.String(), conflicts)
		}
	}
}

func (dp *Pool) checkReverseConflicts(conflicts mapStringSet) {
	for _, pkg := range dp.alpmExecutor.LocalPackages() {
		if dp.hasPackage(pkg.Name()) {
			continue
		}
		for _, conflict := range dp.alpmExecutor.PackageConflicts(pkg) {
			dp.checkReverseConflict(pkg.Name(), conflict.String(), conflicts)
		}
	}
}

func (dp *Pool) CheckConflicts(useAsk, noConfirm bool) (map[string]stringset.StringSet, error) {
	var wg sync.WaitGroup
	innerConflicts := make(mapStringSet)
	conflicts := make(mapStringSet)
	wg.Add(2)

	text.OperationInfoln(text.T("Checking for conflicts..."))
	go func() {
		dp.checkForwardConflicts(conflicts)
		dp.checkReverseConflicts(conflicts)
		wg.Done()
	}()

	text.OperationInfoln(text.T("Checking for inner conflicts..."))
	go func() {
		dp.checkInnerConflicts(innerConflicts)
		wg.Done()
	}()

	wg.Wait()

	if len(innerConflicts) != 0 {
		text.Errorln(text.T("\nInner conflicts found:"))

		for name, pkgs := range innerConflicts {
			str := text.SprintError(name + ":")
			for pkg := range pkgs.Iter() {
				str += " " + text.Cyan(pkg) + ","
			}
			str = strings.TrimSuffix(str, ",")

			text.Println(str)
		}
	}

	if len(conflicts) != 0 {
		text.Errorln(text.T("\nPackage conflicts found:"))

		for name, pkgs := range conflicts {
			str := text.SprintError(text.Tf("Installing %s will remove:", text.Cyan(name)))
			for pkg := range pkgs.Iter() {
				str += " " + text.Cyan(pkg) + ","
			}
			str = strings.TrimSuffix(str, ",")

			text.Println(str)
		}
	}

	// Add the inner conflicts to the conflicts
	// These are used to decide what to pass --ask to (if set) or don't pass --noconfirm to
	// As we have no idea what the order is yet we add every inner conflict to the slice
	for name, pkgs := range innerConflicts {
		conflicts[name] = stringset.Make()
		for pkg := range pkgs.Iter() {
			conflicts[pkg] = stringset.Make()
		}
	}

	if len(conflicts) > 0 {
		if !useAsk {
			if noConfirm {
				return nil, text.ErrT("package conflicts can not be resolved with noconfirm, aborting")
			}

			text.Errorln(text.T("Conflicting packages will have to be confirmed manually"))
		}
	}

	return conflicts, nil
}

type missing struct {
	Good    stringset.StringSet
	Missing map[string][][]string
}

func (dp *Pool) _checkMissing(dep string, stack []string, missing *missing) {
	if missing.Good.Get(dep) {
		return
	}

	if trees, ok := missing.Missing[dep]; ok {
		for _, tree := range trees {
			if stringSliceEqual(tree, stack) {
				return
			}
		}
		missing.Missing[dep] = append(missing.Missing[dep], stack)
		return
	}

	aurPkg := dp.findSatisfierAur(dep)
	if aurPkg != nil {
		missing.Good.Set(dep)
		for _, deps := range [3][]string{aurPkg.Depends, aurPkg.MakeDepends, aurPkg.CheckDepends} {
			for _, aurDep := range deps {
				if dp.alpmExecutor.LocalSatisfierExists(aurDep) {
					missing.Good.Set(aurDep)
					continue
				}

				dp._checkMissing(aurDep, append(stack, aurPkg.Name), missing)
			}
		}

		return
	}

	repoPkg := dp.findSatisfierRepo(dep)
	if repoPkg != nil {
		missing.Good.Set(dep)
		for _, dep := range dp.alpmExecutor.PackageDepends(repoPkg) {
			if dp.alpmExecutor.LocalSatisfierExists(dep.String()) {
				missing.Good.Set(dep.String())
				continue
			}

			dp._checkMissing(dep.String(), append(stack, repoPkg.Name()), missing)
		}

		return
	}

	missing.Missing[dep] = [][]string{stack}
}

func stringSliceEqual(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (dp *Pool) CheckMissing() error {
	missing := &missing{
		stringset.Make(),
		make(map[string][][]string),
	}

	for _, target := range dp.targets {
		dp._checkMissing(target.DepString(), make([]string, 0), missing)
	}

	if len(missing.Missing) == 0 {
		return nil
	}

	text.Errorln(text.T("Could not find all required packages:"))
	for dep, trees := range missing.Missing {
		for _, tree := range trees {
			text.EPrintf("\t%s", text.Cyan(dep))

			if len(tree) == 0 {
				text.EPrint(text.T(" (Target"))
			} else {
				text.EPrint(text.T(" (Wanted by: "))
				for n := 0; n < len(tree)-1; n++ {
					text.EPrint(text.Cyan(tree[n]), " -> ")
				}
				text.EPrint(text.Cyan(tree[len(tree)-1]))
			}

			text.EPrintln(")")
		}
	}

	return errors.New("")
}

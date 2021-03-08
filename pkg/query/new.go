package query

// TODO: rename file

import (
	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/stringset"
)

// HangingPackages returns a list of packages installed as deps
// and unneeded by the system
// removeOptional decides whether optional dependencies are counted or not
func HangingPackages(removeOptional bool, dbExecutor db.Executor) (hanging []string) {
	// safePackages represents every package in the system in one of 3 states
	// State = 0 - Remove package from the system
	// State = 1 - Keep package in the system; need to iterate over dependencies
	// State = 2 - Keep package and have iterated over dependencies
	safePackages := make(map[string]uint8)
	// provides stores a mapping from the provides name back to the original package name
	provides := make(map[string]stringset.StringSet)

	packages := dbExecutor.LocalPackages()
	// Mark explicit dependencies and enumerate the provides list
	for _, pkg := range packages {
		if pkg.Reason() == db.PkgReasonExplicit {
			safePackages[pkg.Name()] = 1
		} else {
			safePackages[pkg.Name()] = 0
		}

		for _, dep := range dbExecutor.PackageProvides(pkg) {
			stringset.Add(provides, dep.Name, pkg.Name())
		}
	}

	iterateAgain := true

	for iterateAgain {
		iterateAgain = false
		for _, pkg := range packages {
			if state := safePackages[pkg.Name()]; state == 0 || state == 2 {
				continue
			}

			safePackages[pkg.Name()] = 2
			deps := dbExecutor.PackageDepends(pkg)
			if !removeOptional {
				deps = append(deps, dbExecutor.PackageOptionalDepends(pkg)...)
			}

			// Update state for dependencies
			for _, dep := range deps {
				// Don't assume a dependency is installed
				state, ok := safePackages[dep.Name]
				if !ok {
					// Check if dep is a provides rather than actual package name
					if pset, ok2 := provides[dep.Name]; ok2 {
						for p := range pset.Iter() {
							if safePackages[p] == 0 {
								iterateAgain = true
								safePackages[p] = 1
							}
						}
					}
					continue
				}

				if state == 0 {
					iterateAgain = true
					safePackages[dep.Name] = 1
				}
			}
		}
	}

	// Build list of packages to be removed
	for _, pkg := range packages {
		if safePackages[pkg.Name()] == 0 {
			hanging = append(hanging, pkg.Name())
		}
	}

	return hanging
}

// Statistics returns statistics about packages installed in system
func Statistics(dbExecutor db.Executor) struct {
	Totaln    int
	Expln     int
	TotalSize int64
} {
	var totalSize int64
	localPackages := dbExecutor.LocalPackages()
	totalInstalls := 0
	explicitInstalls := 0

	for _, pkg := range localPackages {
		totalSize += pkg.ISize()
		totalInstalls++
		if pkg.Reason() == db.PkgReasonExplicit {
			explicitInstalls++
		}
	}

	return struct {
		Totaln    int
		Expln     int
		TotalSize int64
	}{
		totalInstalls, explicitInstalls, totalSize,
	}
}

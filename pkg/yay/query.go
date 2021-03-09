package yay

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	rpc "github.com/mikkeloscar/aur"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/stringset"
	"github.com/Jguer/yay/v10/pkg/text"
	"github.com/Jguer/yay/v10/pkg/upgrade"
)

var errMissing = errors.New("missing")

type enum = int

// Query is a collection of Results
type aurQuery = []query.Pkg

// Query holds the results of a repository search.
type repoQuery = []db.IPackage

// PrintSearch handles printing search results in a given format
func printSearchAUR(q aurQuery, dbExecutor db.Executor, start int, searchMode settings.SearchMode, sortOrder enum) {
	for i := range q {
		var toprint string
		if searchMode == settings.NumberMenu {
			switch sortOrder {
			case settings.TopDown:
				toprint += text.Magenta(strconv.Itoa(start+i) + " ")
			case settings.BottomUp:
				toprint += text.Magenta(strconv.Itoa(len(q)+start-i-1) + " ")
			default:
				text.Warnln(text.T("invalid sort mode. Fix with yay -Y --bottomup --save"))
			}
		} else if searchMode == settings.Minimal {
			text.Println(q[i].Name)
			continue
		}

		toprint += text.Bold(text.ColorHash("aur")) + "/" + text.Bold(q[i].Name) +
			" " + text.Cyan(q[i].Version) +
			text.Bold(" (+"+strconv.Itoa(q[i].NumVotes)) +
			" " + text.Bold(strconv.FormatFloat(q[i].Popularity, 'f', 2, 64)+") ")

		if q[i].Maintainer == "" {
			toprint += text.Bold(text.Red(text.T("(Orphaned)"))) + " "
		}

		if q[i].OutOfDate != 0 {
			toprint += text.Bold(text.Red(text.Tf("(Out-of-date: %s)", text.FormatTime(q[i].OutOfDate)))) + " "
		}

		if pkg := dbExecutor.LocalPackage(q[i].Name); pkg != nil {
			if pkg.Version() != q[i].Version {
				toprint += text.Bold(text.Green(text.Tf("(Installed: %s)", pkg.Version())))
			} else {
				toprint += text.Bold(text.Green(text.T("(Installed)")))
			}
		}
		toprint += "\n    " + q[i].Description
		text.Println(toprint)
	}
}

// PrintSearch receives a RepoSearch type and outputs pretty text.
func printSearchRepo(s repoQuery, dbExecutor db.Executor, searchMode settings.SearchMode, sortMode enum) {
	for i, res := range s {
		var toprint string
		if searchMode == settings.NumberMenu {
			switch sortMode {
			case settings.TopDown:
				toprint += text.Magenta(strconv.Itoa(i+1) + " ")
			case settings.BottomUp:
				toprint += text.Magenta(strconv.Itoa(len(s)-i) + " ")
			default:
				text.Warnln(text.T("invalid sort mode. Fix with yay -Y --bottomup --save"))
			}
		} else if searchMode == settings.Minimal {
			text.Println(res.Name())
			continue
		}

		toprint += text.Bold(text.ColorHash(res.DB().Name())) + "/" + text.Bold(res.Name()) +
			" " + text.Cyan(res.Version()) +
			text.Bold(" ("+text.Human(res.Size())+
				" "+text.Human(res.ISize())+") ")

		packageGroups := dbExecutor.PackageGroups(res)
		if len(packageGroups) != 0 {
			toprint += fmt.Sprint(packageGroups, " ")
		}

		if pkg := dbExecutor.LocalPackage(res.Name()); pkg != nil {
			if pkg.Version() != res.Version() {
				toprint += text.Bold(text.Green(text.Tf("(Installed: %s)", pkg.Version())))
			} else {
				toprint += text.Bold(text.Green(text.T("(Installed)")))
			}
		}

		toprint += "\n    " + res.Description()
		text.Println(toprint)
	}
}

func reverse(s repoQuery) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func sortAURQuery(q aurQuery, sortBy string, sortMode enum) func(int, int) bool {
	return func(i, j int) bool {
		var result bool

		switch sortBy {
		case "votes":
			result = q[i].NumVotes > q[j].NumVotes
		case "popularity":
			result = q[i].Popularity > q[j].Popularity
		case "name":
			result = text.LessRunes([]rune(q[i].Name), []rune(q[j].Name))
		case "base":
			result = text.LessRunes([]rune(q[i].PackageBase), []rune(q[j].PackageBase))
		case "submitted":
			result = q[i].FirstSubmitted < q[j].FirstSubmitted
		case "modified":
			result = q[i].LastModified < q[j].LastModified
		case "id":
			result = q[i].ID < q[j].ID
		case "baseid":
			result = q[i].PackageBaseID < q[j].PackageBaseID
		}

		if sortMode == settings.BottomUp {
			return !result
		}

		return result
	}
}

func getSearchBy(value string) rpc.By {
	switch value {
	case "name":
		return rpc.Name
	case "maintainer":
		return rpc.Maintainer
	case "depends":
		return rpc.Depends
	case "makedepends":
		return rpc.MakeDepends
	case "optdepends":
		return rpc.OptDepends
	case "checkdepends":
		return rpc.CheckDepends
	default:
		return rpc.NameDesc
	}
}

// NarrowSearch searches AUR and narrows based on subarguments
func narrowSearch(aur *query.AUR, pkgS []string, sortS bool, searchBy string, sortBy string) (aurQuery, error) {
	var r []query.Pkg
	var err error
	var usedIndex int

	by := getSearchBy(searchBy)

	if len(pkgS) == 0 {
		return nil, nil
	}

	for i, word := range pkgS {
		r, err = aur.SearchBy(word, by)
		if err == nil {
			usedIndex = i
			break
		}
	}

	if err != nil {
		return nil, err
	}

	if len(pkgS) == 1 {
		if sortS {
			sort.Slice(aurQuery(r), sortAURQuery(r, sortBy, settings.TopDown))
		}
		return r, err
	}

	var aq aurQuery
	var n int

	for i := range r {
		match := true
		for j, pkgN := range pkgS {
			if usedIndex == j {
				continue
			}

			if !(strings.Contains(r[i].Name, pkgN) || strings.Contains(strings.ToLower(r[i].Description), pkgN)) {
				match = false
				break
			}
		}

		if match {
			n++
			aq = append(aq, r[i])
		}
	}

	if sortS {
		sort.Slice(aq, sortAURQuery(aq, sortBy, settings.TopDown))
	}

	return aq, err
}

// SyncSearch presents a query to the local repos and to the AUR.
func syncSearch(pkgS []string, rt *Runtime) (err error) {
	pkgS = query.RemoveInvalidTargets(pkgS, rt.Config.Mode)
	var aurErr error
	var aq aurQuery
	var pq repoQuery

	switch rt.Config.Mode {
	case settings.ModeAUR:
		aq, aurErr = narrowSearch(rt.AUR, pkgS, true, rt.Config.SearchBy, rt.Config.SortBy)
	case settings.ModeRepo:
		pq = queryRepo(pkgS, rt.DB, rt.Config.SortMode)
	case settings.ModeAny:
		aq, aurErr = narrowSearch(rt.AUR, pkgS, true, rt.Config.SearchBy, rt.Config.SortBy)
		pq = queryRepo(pkgS, rt.DB, rt.Config.SortMode)
	}

	switch rt.Config.SortMode {
	case settings.TopDown:
		if rt.Config.Mode == settings.ModeRepo || rt.Config.Mode == settings.ModeAny {
			printSearchRepo(pq, rt.DB, rt.Config.SearchMode, rt.Config.SortMode)
		}
		if rt.Config.Mode == settings.ModeAUR || rt.Config.Mode == settings.ModeAny {
			printSearchAUR(aq, rt.DB, 1, rt.Config.SearchMode, rt.Config.SortMode)
		}
	case settings.BottomUp:
		if rt.Config.Mode == settings.ModeAUR || rt.Config.Mode == settings.ModeAny {
			printSearchAUR(aq, rt.DB, 1, rt.Config.SearchMode, rt.Config.SortMode)
		}
		if rt.Config.Mode == settings.ModeRepo || rt.Config.Mode == settings.ModeAny {
			printSearchRepo(pq, rt.DB, rt.Config.SearchMode, rt.Config.SortMode)
		}
	default:
		return errors.New(text.T("invalid sort mode. Fix with yay -Y --bottomup --save"))
	}

	if aurErr != nil {
		text.Errorln(text.Tf("error during AUR search: %s", aurErr))
		text.Warnln(text.T("Showing repo packages only"))
	}

	return nil
}

// SyncInfo serves as a pacman -Si for repo packages and AUR packages.
func syncInfo(cmdArgs *settings.PacmanConf, pkgS []string, rt *Runtime) error {
	var info []*query.Pkg
	var err error
	missing := false
	pkgS = query.RemoveInvalidTargets(pkgS, rt.Config.Mode)
	aurS, repoS := packageSlices(pkgS, rt.DB, rt.Config.Mode)

	if len(aurS) != 0 {
		noDB := make([]string, 0, len(aurS))

		for _, pkg := range aurS {
			_, name := text.SplitDBFromName(pkg)
			noDB = append(noDB, name)
		}

		info, err = query.AURInfoPrint(rt.AUR, noDB, rt.Config.RequestSplitN)
		if err != nil {
			missing = true
			text.EPrintln(err)
		}
	}

	// Repo always goes first
	if len(repoS) != 0 {
		arguments := cmdArgs.DeepCopy()
		arguments.Targets = nil
		*arguments.Targets = append(*arguments.Targets, repoS...)
		err = rt.CmdRunner.Show(PassToPacman(rt.Config, arguments))

		if err != nil {
			return err
		}
	}

	if len(aurS) != len(info) {
		missing = true
	}

	if len(info) != 0 {
		for _, pkg := range info {
			printInfo(pkg, rt.Config.AURURL, cmdArgs.ModeConf.(*settings.QConf).Info > 1)
		}
	}

	if missing {
		err = errMissing
	}

	return err
}

// Search handles repo searches. Creates a RepoSearch struct.
func queryRepo(pkgInputN []string, dbExecutor db.Executor, sortOrder enum) repoQuery {
	s := repoQuery(dbExecutor.SyncPackages(pkgInputN...))

	if sortOrder == settings.BottomUp {
		reverse(s)
	}
	return s
}

// PackageSlices separates an input slice into aur and repo slices
func packageSlices(toCheck []string, dbExecutor db.Executor, mode settings.TargetMode) (aur, repo []string) {
	for _, _pkg := range toCheck {
		dbName, name := text.SplitDBFromName(_pkg)
		found := false

		if dbName == "aur" || mode == settings.ModeAUR {
			aur = append(aur, _pkg)
			continue
		} else if dbName != "" || mode == settings.ModeRepo {
			repo = append(repo, _pkg)
			continue
		}

		found = dbExecutor.SyncSatisfierExists(name)

		if !found {
			found = len(dbExecutor.PackagesFromGroup(name)) != 0
		}

		if found {
			repo = append(repo, _pkg)
		} else {
			aur = append(aur, _pkg)
		}
	}

	return aur, repo
}

func printNumberOfUpdates(rt *Runtime, enableDowngrade bool) error {
	warnings := query.NewWarnings()

	var (
		aurUp  []upgrade.Upgrade
		repoUp []upgrade.Upgrade
		err    error
	)

	text.CaptureOutput(nil, nil, func() {
		aurUp, repoUp, err = upList(warnings, rt, enableDowngrade)
	})

	if err != nil {
		return err
	}
	text.Println(len(aurUp) + len(repoUp))

	return nil
}

func printUpdateList(cmdArgs *settings.PacmanConf, rt *Runtime, enableDowngrade bool) error {
	targets := stringset.Make(*cmdArgs.Targets...)
	warnings := query.NewWarnings()

	var (
		err         error
		localNames  []string
		remoteNames []string
		aurUp       []upgrade.Upgrade
		repoUp      []upgrade.Upgrade
	)
	text.CaptureOutput(nil, nil, func() {
		localNames, remoteNames, err = query.GetPackageNamesBySource(rt.DB)
		if err != nil {
			return
		}

		aurUp, repoUp, err = upList(warnings, rt, enableDowngrade)
	})

	if err != nil {
		return err
	}

	noTargets := targets.Len() == 0

	qconf := cmdArgs.ModeConf.(*settings.QConf)

	if !qconf.Foreign {
		for _, pkg := range repoUp {
			if noTargets || targets.Get(pkg.Name) {
				if qconf.Quiet {
					text.Printf("%s\n", pkg.Name)
				} else {
					text.Printf("%s %s -> %s\n", text.Bold(pkg.Name), text.Green(pkg.LocalVersion), text.Green(pkg.RemoteVersion))
				}
				targets.Remove(pkg.Name)
			}
		}
	}

	if !qconf.Native {
		for _, pkg := range aurUp {
			if noTargets || targets.Get(pkg.Name) {
				if qconf.Quiet {
					text.Printf("%s\n", pkg.Name)
				} else {
					text.Printf("%s %s -> %s\n", text.Bold(pkg.Name), text.Green(pkg.LocalVersion), text.Green(pkg.RemoteVersion))
				}
				targets.Remove(pkg.Name)
			}
		}
	}

	missing := false

outer:
	for pkg := range targets.Iter() {
		for _, name := range localNames {
			if name == pkg {
				continue outer
			}
		}

		for _, name := range remoteNames {
			if name == pkg {
				continue outer
			}
		}

		text.Errorln(text.Tf("package '%s' was not found", pkg))
		missing = true
	}

	if missing {
		return fmt.Errorf("missing")
	}

	return nil
}

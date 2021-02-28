package settings

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jguer/yay/v10/pkg/settings/parser"
	"github.com/Jguer/yay/v10/pkg/text"
)

func mappingFunc() func(string) (parser.Enum, bool) {
	mainOp := OpYay
	mainOpSet := false

	return func(s string) (parser.Enum, bool) {
		enum := mapping(s, mainOp)

		switch enum {
		case database, files, query, remove, sync, depTest, upgrade, yay, show, getPkgbuild:

			if mainOpSet {
				return parser.InvalidOption, false
			} else {
				mainOpSet = true
			}
		}

		switch enum {
		case database:
			mainOp = OpDatabase
		case files:
			mainOp = OpFiles
		case query:
			mainOp = OpQuery
		case remove:
			mainOp = OpRemove
		case sync:
			mainOp = OpSync
		case depTest:
			mainOp = OpDepTest
		case upgrade:
			mainOp = OpUpgrade

		case yay:
			mainOp = OpYay
		case show:
			mainOp = OpShow
		case getPkgbuild:
			mainOp = OpGetPkgbuild

		case help:
			mainOp = OpHelp
		case version:
			mainOp = OpVersion
		}
		exists := takesParam(enum, mainOp)
		return enum, exists
	}
}

func ParseCommandLine(args []string) (*YayConfig, error) {
	if len(args) == 0 {
		args = []string{"-Syu"}
	}

	conf, err := newConfig()
	if err != nil {
		return nil, err
	}

	yay := &YayConfig{
		PersistentYayConfig: *conf,
		CompletionPath:      filepath.Join(getCacheHome(), completionFileName),
		ConfigPath:          getConfigPath(),
		Pacman:              new(PacmanConf),
	}
	yay.Pacman.Targets = &yay.Targets

	err = parseCommandLine(args, yay, text.InRef())
	if err != nil {
		return nil, err
	}
	return yay, err
}

func parseCommandLine(args []string, yay *YayConfig, r *io.Reader) error {

	a, err := parser.Parse(mappingFunc(), args, r)
	var uoErr parser.ErrUnknownOption
	if ok := errors.As(err, &uoErr); ok {
		switch uoErr {
		case "D", "Q", "R", "S", "T", "U", "F", "Y", "P", "G":
			fallthrough
		case "database", "query", "remove", "sync", "deptest", "upgrade", "files", "yay", "show", "getpkgbuild":
			return fmt.Errorf("main op may not be specified multiple times")
		}
	}
	if err != nil {
		return err
	}

	a.Iterate(handleConfig(yay, &err))

	if yay.MainOperation == 0 {
		yay.MainOperation = OpYay
		yay.ModeConf = &YConf{}
	}

	yay.Targets = a.Targets()

	return err
}

func handleConfig(conf *YayConfig, err *error) func(option parser.Enum, value []string) bool {

	last := func(s []string) string { return s[len(s)-1] }

	type pkgVerSl []struct {
		Package string
		Version string
	}

	toPackageVersion := func(s []string) pkgVerSl {
		var pvs = make(pkgVerSl, 0, len(s))
		for _, v := range s {
			pv := strings.SplitN(v, "=", 2)
			if len(pv) != 2 {
				continue
			}
			pvs = append(pvs, struct {
				Package string
				Version string
			}{pv[0], pv[1]})
		}
		return pvs
	}

	opModePacman := func() bool {
		switch conf.MainOperation {
		case OpDatabase, OpQuery, OpRemove, OpSync, OpDepTest, OpUpgrade, OpFiles:
			return true
		default:
			return false
		}
	}

	return func(option parser.Enum, value []string) (res bool) {
		// TODO(jmh): stronger validation
		// TODO(jmh): propper error reporting

		defer func() {
			r := recover()
			if re, ok := r.(error); ok && strings.HasPrefix(re.Error(), "interface conversion:") {
				res = false
				*err = fmt.Errorf("wrong argument order")
				return
			}
			if r != nil {
				panic(r)
			}
		}()

		switch option {

		default:
			panic("unknown enum argument")

		// -- Main Operations --

		case database:
			conf.Pacman.ModeConf = new(DConf)
			conf.MainOperation = OpDatabase
		case query:
			conf.Pacman.ModeConf = new(QConf)
			conf.MainOperation = OpQuery
		case remove:
			conf.Pacman.ModeConf = new(RConf)
			conf.MainOperation = OpRemove
		case sync:
			conf.Pacman.ModeConf = new(SConf)
			conf.MainOperation = OpSync
		case depTest:
			conf.Pacman.ModeConf = new(TConf)
			conf.MainOperation = OpDepTest
		case upgrade:
			conf.Pacman.ModeConf = new(UConf)
			conf.MainOperation = OpUpgrade
		case files:
			conf.Pacman.ModeConf = new(FConf)
			conf.MainOperation = OpFiles

		case version:
			if opModePacman() {
				conf.Pacman.ModeConf = new(VConf)
			}
			conf.MainOperation = OpVersion
		case help:
			if opModePacman() {
				conf.Pacman.ModeConf = new(HConf)
			}
			conf.MainOperation = OpHelp

		case yay:
			conf.ModeConf = new(YConf)
			conf.MainOperation = OpYay
		case show:
			conf.ModeConf = new(PConf)
			conf.MainOperation = OpShow
		case getPkgbuild:
			conf.ModeConf = new(GConf)
			conf.MainOperation = OpGetPkgbuild

		// -- Yay and Pacman Options --

		case config:
			if opModePacman() {
				conf.Pacman.Config = last(value)
			}
			conf.PacmanConf = last(value)

		case noConfirm:
			if opModePacman() {
				conf.Pacman.NoConfirm = true
			}
			userNoConfirm = true
		case confirm:
			if opModePacman() {
				conf.Pacman.NoConfirm = false
			}
			userNoConfirm = false
		case upgrades:
			if opModePacman() {
				conf.Pacman.ModeConf.(*QConf).Upgrades = Trilean(parser.GetCount(value))
			} else {
				conf.ModeConf.(*PConf).Upgrades = true
			}

		// -- Options (PQFDS) --

		case quiet:
			if pc, ok := conf.ModeConf.(*PConf); ok {
				pc.Quiet = true
				break
			}
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.Quiet = true
			case *FConf:
				t.Quiet = true
			case *DConf:
				t.Quiet = true
			case *SConf:
				t.Quiet = true
			}

		// -- Pacman Global Options --

		case dbPath:
			conf.Pacman.DBPath = last(value)
		case root:
			conf.Pacman.Root = last(value)
		case verbose:
			conf.Pacman.Verbose = true
		case arch:
			conf.Pacman.Arch = last(value)
		case cacheDir:
			conf.Pacman.CacheDir = last(value)
		case color:
			switch last(value) {
			case "always":
				conf.Pacman.Color = ColorAlways
			case "never":
				conf.Pacman.Color = ColorNever
			case "auto":
				conf.Pacman.Color = ColorAuto
			default:
				text.EPrintf("unknown value for color %q", last(value))
			}
		case debug:
			conf.Pacman.Debug = true
		case gpgDir:
			conf.Pacman.GPGDir = last(value)
		case hookDir:
			conf.Pacman.HookDir = last(value)
		case logFile:
			conf.Pacman.LogFile = last(value)
		case disableDownloadTimeout:
			conf.Pacman.DisableDownloadTimeout = true
		case sysRoot:
			conf.Pacman.SysRoot = last(value)

		case ask:
			conf.Pacman.Ask, _ = strconv.Atoi(last(value))

		// -- Pacman Transaction Options (SRU) --

		case noDeps:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.NoDeps = Trilean(parser.GetCount(value))
			case *RConf:
				t.NoDeps = Trilean(parser.GetCount(value))
			case *UConf:
				t.NoDeps = Trilean(parser.GetCount(value))
			}
		case assumeInstalled:
			pvs := toPackageVersion(value)
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.AssumeInstalled = pvs
			case *RConf:
				t.AssumeInstalled = pvs
			case *UConf:
				t.AssumeInstalled = pvs
			}
		case dbOnly:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.DBOnly = true
			case *RConf:
				t.DBOnly = true
			case *UConf:
				t.DBOnly = true
			}
		case noProgressbar:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.NoProgressbar = true
			case *RConf:
				t.NoProgressbar = true
			case *UConf:
				t.NoProgressbar = true
			}
		case noScriptlet:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.NoScriptlet = true
			case *RConf:
				t.NoScriptlet = true
			case *UConf:
				t.NoScriptlet = true
			}
		case printFormat:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.PrintFormat = last(value)
			case *RConf:
				t.PrintFormat = last(value)
			case *UConf:
				t.PrintFormat = last(value)
			}
			fallthrough
		case print:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.Print = true
			case *RConf:
				t.Print = true
			case *UConf:
				t.Print = true
			}

		// -- Pacman Options (SUD) --

		case asDeps:
			switch t := conf.Pacman.ModeConf.(type) {
			case *DConf:
				t.AsDeps = last(value)
			case *UConf:
				t.AsDeps = true
			case *SConf:
				t.AsDeps = true
			}
		case asExplicit:
			switch t := conf.Pacman.ModeConf.(type) {
			case *DConf:
				t.AsExplicit = last(value)
			case *UConf:
				t.AsExplicit = true
			case *SConf:
				t.AsExplicit = true
			}

		// -- Pacman Options (QFS) --

		case list:
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.List = true
			case *FConf:
				t.List = true
			case *SConf:
				t.List = true
			}

		// -- Pacman Upgrade Options (SU) --

		case ignore:
			switch t := conf.Pacman.ModeConf.(type) {
			case *UConf:
				t.Ignore = value // TODO(jmh): split at ',' same for below
			case *SConf:
				t.Ignore = value
			}
		case ignoreGroup:
			switch t := conf.Pacman.ModeConf.(type) {
			case *UConf:
				t.IgnoreGroup = value
			case *SConf:
				t.IgnoreGroup = value
			}
		case needed:
			switch t := conf.Pacman.ModeConf.(type) {
			case *UConf:
				t.Needed = true
			case *SConf:
				t.Needed = true
			}
		case overwrite:
			switch t := conf.Pacman.ModeConf.(type) {
			case *UConf:
				t.Overwrite = last(value)
			case *SConf:
				t.Overwrite = last(value)
			}

		// -- Pacman Options (QS) --

		case groups:
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.Groups = true
			case *SConf:
				t.Groups = Trilean(parser.GetCount(value))
			}
		case info:
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.Info = Trilean(parser.GetCount(value))
			case *SConf:
				t.Info = Trilean(parser.GetCount(value))
			}
		case search:
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.Search = last(value)
			case *SConf:
				t.Search = last(value)
			}

		// -- Pacman Options (SF) --

		case refresh:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.Refresh = Trilean(parser.GetCount(value))
			case *FConf:
				t.Refresh = Trilean(parser.GetCount(value))
			}

		// -- Pacman Options (QD) --

		case check:
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.Check = Trilean(parser.GetCount(value))
			case *DConf:
				t.Check = Trilean(parser.GetCount(value))
			}

		// -- Pacman Query Options --

		case changelog:
			conf.Pacman.ModeConf.(*QConf).Changelog = true
		case deps:
			conf.Pacman.ModeConf.(*QConf).Deps = true
		case explicit:
			conf.Pacman.ModeConf.(*QConf).Explicit = true
		case foreign:
			conf.Pacman.ModeConf.(*QConf).Foreign = true
		case native:
			conf.Pacman.ModeConf.(*QConf).Native = true
		case owns:
			conf.Pacman.ModeConf.(*QConf).Owns = last(value)
		case file:
			conf.Pacman.ModeConf.(*QConf).File = true
		case unrequired:
			conf.Pacman.ModeConf.(*QConf).Unrequired = Trilean(parser.GetCount(value))

		// -- Pacman Remove Options --

		case cascade:
			conf.Pacman.ModeConf.(*RConf).Cascade = true
		case noSave:
			conf.Pacman.ModeConf.(*RConf).NoSave = true
		case recursive:
			conf.Pacman.ModeConf.(*RConf).Recursive = Trilean(parser.GetCount(value))
		case unneeded:
			conf.Pacman.ModeConf.(*RConf).Unneeded = true

		// -- Pacman Sync Options --

		case clean:
			conf.Pacman.ModeConf.(*SConf).Clean = Trilean(parser.GetCount(value))
		case sysUpgrade:
			conf.Pacman.ModeConf.(*SConf).SysUpgrade = Trilean(parser.GetCount(value))
		case downloadOnly:
			conf.Pacman.ModeConf.(*SConf).DownloadOnly = true

		// -- Pacman File Options --

		case regex:
			conf.Pacman.ModeConf.(*FConf).Regex = true
		case machineReadable:
			conf.Pacman.ModeConf.(*FConf).MachineReadable = true

		// -- Persistent Yay Options --

		case aurURL:
			conf.AURURL = strings.TrimRight(last(value), "/")
		case buildDir:
			conf.BuildDir = last(value)
		case absDir:
			conf.ABSDir = last(value)

		case cleanAfter:
			conf.CleanAfter = true
		case noCleanAfter:
			conf.CleanAfter = false

		case devel:
			conf.Devel = true
		case noDevel:
			conf.Devel = false

		case timeUpdate:
			conf.TimeUpdate = true
		case noTimeUpdate:
			conf.TimeUpdate = false

		case topdown:
			conf.SortMode = TopDown
		case bottomup:
			conf.SortMode = BottomUp

		case completionInterval:
			n, err := strconv.Atoi(last(value))
			if err == nil {
				conf.CompletionInterval = n
			}

		case sortBy:
			conf.SortBy = last(value)

		case searchBy:
			conf.SearchBy = last(value)

		case redownload:
			conf.ReDownload = "yes"
		case redownloadAll:
			conf.ReDownload = "all"
		case noRedownload:
			conf.ReDownload = "no"

		case rebuild:
			conf.ReBuild = "yes"
		case rebuildAll:
			conf.ReBuild = "all"
		case rebuildTree:
			conf.ReBuild = "tree"
		case noRebuild:
			conf.ReBuild = "no"

		case batchInstall:
			conf.BatchInstall = true
		case noBatchInstall:
			conf.BatchInstall = false

		case answerClean:
			conf.AnswerClean = last(value)
		case noAnswerClean:
			conf.AnswerClean = ""
		case answerDiff:
			conf.AnswerDiff = last(value)
		case noAnswerDiff:
			conf.AnswerDiff = ""
		case answerEdit:
			conf.AnswerEdit = last(value)
		case noAnswerEdit:
			conf.AnswerEdit = ""
		case answerUpgrade:
			conf.AnswerUpgrade = last(value)
		case noAnswerUpgrade:
			conf.AnswerUpgrade = ""

		case gpg:
			conf.GpgBin = last(value)
		case gpgFlags:
			conf.GpgFlags = last(value)

		case git:
			conf.GitBin = last(value)
		case gitFlags:
			conf.GitFlags = last(value)

		case editor:
			conf.Editor = last(value)
		case editorFlags:
			conf.EditorFlags = last(value)

		case mFlags:
			conf.MFlags = last(value)
		case makepkg:
			conf.MakepkgBin = last(value)
		case makePkgconf:
			conf.MakepkgConf = last(value)
		case noMakePkgconf:
			conf.MakepkgConf = ""

		case tar:
			conf.Tar = last(value)

		case pacman:
			conf.PacmanBin = last(value)

		case sudo:
			conf.SudoBin = last(value)
		case sudoFlags:
			conf.SudoFlags = last(value)

		case sudoLoop:
			conf.SudoLoop = true
		case noSudoLoop:
			conf.SudoLoop = false

		case requestSplitN:
			n, _ := strconv.Atoi(last(value))
			if n > 0 {
				conf.RequestSplitN = n
			}

		case provides:
			conf.Provides = true
		case noProvides:
			conf.Provides = false

		case pgpFetch:
			conf.PGPFetch = true
		case noPGPFetch:
			conf.PGPFetch = false

		case upgradeMenu:
			conf.UpgradeMenu = true
		case noUpgradeMenu:
			conf.UpgradeMenu = false
		case cleanMenu:
			conf.CleanMenu = true
		case noCleanMenu:
			conf.CleanMenu = false
		case diffMenu:
			conf.DiffMenu = true
		case noDiffMenu:
			conf.DiffMenu = false
		case editMenu:
			conf.EditMenu = true
		case noEditMenu:
			conf.EditMenu = false

		case useAsk:
			conf.UseAsk = true
		case noUseAsk:
			conf.UseAsk = false

		case combinedUpgrade:
			conf.CombinedUpgrade = true
		case noCombinedUpgrade:
			conf.CombinedUpgrade = false

		case removeMake:
			conf.RemoveMake = "yes"
		case noRemoveMake:
			conf.RemoveMake = "no"
		case askRemoveMake:
			conf.RemoveMake = "ask"

		// -- Yay Show Options --

		case complete:
			conf.ModeConf.(*PConf).Complete = Trilean(parser.GetCount(value))
		case defaultConfig:
			conf.ModeConf.(*PConf).DefaultConfig = true
		case currentConfig:
			conf.ModeConf.(*PConf).CurrentConfig = true
		case stats:
			conf.ModeConf.(*PConf).LocalStats = true
		case news:
			conf.ModeConf.(*PConf).News = true
		case numberUpgrades:
			conf.ModeConf.(*PConf).NumberUpgrades = true
		case fish:
			conf.ModeConf.(*PConf).Fish = true

		// -- Yay yay-mode Options --

		case yayClean:
			conf.ModeConf.(*YConf).Clean = Trilean(parser.GetCount(value))
		case genDB:
			conf.ModeConf.(*YConf).GenDevDB = true

		// -- Yay GetPkgbuild Options --

		case force:
			if conf.MainOperation != OpGetPkgbuild {
				break
			}
			conf.ModeConf.(*GConf).Force = true

		// -- Other Yay Options --

		case aur:
			conf.Mode = ModeAUR
		case repo:
			conf.Mode = ModeRepo

		case save:
			conf.SaveConfig = true
		}
		return true
	}
}

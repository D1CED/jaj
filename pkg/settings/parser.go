package settings

import (
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
				return parser.InvalidFlag, false
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
		Conf:           *conf,
		CompletionPath: filepath.Join(GetCacheHome(), completionFileName),
		ConfigPath:     GetConfigPath(),
	}

	a, err := parser.Parse(mappingFunc(), args, text.In())
	if err != nil {
		return nil, err
	}

	a.Iterate(handleConfig(yay, &err))

	if yay.MainOperation == 0 {
		yay.MainOperation = OpYay
	}

	yay.Pacman.Targets = a.Targets() // TODO(jmh): move targets to yay

	return yay, err
}

func handleConfig(conf *YayConfig, err *error) func(option parser.Enum, value []string) bool {

	last := func(s []string) string { return s[len(s)-1] }

	toPackageVersion := func(s []string) []struct {
		Package string
		Version string
	} {
		var pvs = make([]struct {
			Package string
			Version string
		}, 0, len(s))
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

	return func(option parser.Enum, value []string) bool {
		// TODO(jmh): stronger validation
		// TODO(jmh): propper error reporting

		switch option {

		default:
			panic("unknown enum argument")

		// -- Main Operations --

		case database:
			conf.Pacman = new(PacmanConf)
			conf.Pacman.ModeConf = new(DConf)
			conf.MainOperation = OpDatabase
		case query:
			conf.Pacman = new(PacmanConf)
			conf.Pacman.ModeConf = new(QConf)
			conf.MainOperation = OpQuery
		case remove:
			conf.Pacman = new(PacmanConf)
			conf.Pacman.ModeConf = new(RConf)
			conf.MainOperation = OpRemove
		case sync:
			conf.Pacman = new(PacmanConf)
			conf.Pacman.ModeConf = new(SConf)
			conf.MainOperation = OpSync
		case depTest:
			conf.Pacman = new(PacmanConf)
			conf.Pacman.ModeConf = new(TConf)
			conf.MainOperation = OpDepTest
		case upgrade:
			conf.Pacman = new(PacmanConf)
			conf.Pacman.ModeConf = new(UConf)
			conf.MainOperation = OpUpgrade
		case files:
			conf.Pacman = new(PacmanConf)
			conf.Pacman.ModeConf = new(FConf)
			conf.MainOperation = OpFiles

		case version:
			if conf.Pacman != nil {
				conf.Pacman.ModeConf = new(VConf)
			}
			conf.MainOperation = OpVersion
		case help:
			if conf.Pacman != nil {
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
			if conf.Pacman != nil {
				conf.Pacman.Config = last(value)
			}
			conf.Conf.PacmanConf = last(value)

		case noConfirm:
			if conf.Pacman != nil {
				conf.Pacman.NoConfirm = true
			}
			userNoConfirm = true
		case confirm:
			if conf.Pacman != nil {
				conf.Pacman.NoConfirm = false
			}
			userNoConfirm = false

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

		// -- Pacman Transaction Options (SRU) --

		case noDeps:
			switch t := conf.Pacman.ModeConf.(type) {
			case *SConf:
				t.NoDeps = true
			case *RConf:
				t.NoDeps = true
			case *UConf:
				t.NoDeps = true
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
				t.Groups = true
			}
		case info:
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.Info = true
			case *SConf:
				t.Info = true
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
				t.Refresh = true
			case *FConf:
				t.Refresh = true
			}

		// -- Pacman Options (QD) --

		case check:
			switch t := conf.Pacman.ModeConf.(type) {
			case *QConf:
				t.Check = true
			case *DConf:
				t.Check = true
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
			conf.Pacman.ModeConf.(*QConf).Unrequired = true
		case upgrades:
			conf.Pacman.ModeConf.(*QConf).Upgrades = true

		// -- Pacman Remove Options --

		case cascade:
			conf.Pacman.ModeConf.(*RConf).Cascade = true
		case noSave:
			conf.Pacman.ModeConf.(*RConf).NoSave = true
		case recursive:
			conf.Pacman.ModeConf.(*RConf).Recursive = true
		case unneeded:
			conf.Pacman.ModeConf.(*RConf).Unneeded = true

		// -- Pacman Sync Options --

		case clean:
			conf.ModeConf.(*SConf).Clean = true
		case sysUpgrade:
			conf.ModeConf.(*SConf).SysUpgrade = true
		case downloadOnly:
			conf.ModeConf.(*SConf).DownloadOnly = true

		// -- Pacman File Options --

		case regex:
			conf.ModeConf.(*FConf).Regex = true
		case machineReadable:
			conf.ModeConf.(*FConf).MachineReadable = true

		// -- Persistent Yay Options --

		case aurURL:
			conf.Conf.AURURL = strings.TrimRight(last(value), "/")
		case buildDir:
			conf.Conf.BuildDir = last(value)
		case absDir:
			conf.Conf.ABSDir = last(value)

		case cleanAfter:
			conf.Conf.CleanAfter = true
		case noCleanAfter:
			conf.Conf.CleanAfter = false

		case devel:
			conf.Conf.Devel = true
		case noDevel:
			conf.Conf.Devel = false

		case timeUpdate:
			conf.Conf.TimeUpdate = true
		case noTimeUpdate:
			conf.Conf.TimeUpdate = false

		case topdown:
			conf.Conf.SortMode = TopDown
		case bottomup:
			conf.Conf.SortMode = BottomUp

		case completionInterval:
			n, err := strconv.Atoi(last(value))
			if err == nil {
				conf.Conf.CompletionInterval = n
			}

		case sortBy:
			conf.Conf.SortBy = last(value)

		case searchBy:
			conf.Conf.SearchBy = last(value)

		case redownload:
			conf.Conf.ReDownload = "yes"
		case redownloadAll:
			conf.Conf.ReDownload = "all"
		case noRedownload:
			conf.Conf.ReDownload = "no"

		case rebuild:
			conf.Conf.ReBuild = "yes"
		case rebuildAll:
			conf.Conf.ReBuild = "all"
		case rebuildTree:
			conf.Conf.ReBuild = "tree"
		case noRebuild:
			conf.Conf.ReBuild = "no"

		case batchInstall:
			conf.Conf.BatchInstall = true
		case noBatchInstall:
			conf.Conf.BatchInstall = false

		case answerClean:
			conf.Conf.AnswerClean = last(value)
		case noAnswerClean:
			conf.Conf.AnswerClean = ""
		case answerDiff:
			conf.Conf.AnswerDiff = last(value)
		case noAnswerDiff:
			conf.Conf.AnswerDiff = ""
		case answerEdit:
			conf.Conf.AnswerEdit = last(value)
		case noAnswerEdit:
			conf.Conf.AnswerEdit = ""
		case answerUpgrade:
			conf.Conf.AnswerUpgrade = last(value)
		case noAnswerUpgrade:
			conf.Conf.AnswerUpgrade = ""

		case gpg:
			conf.Conf.GpgBin = last(value)
		case gpgFlags:
			conf.Conf.GpgFlags = last(value)

		case git:
			conf.Conf.GitBin = last(value)
		case gitFlags:
			conf.Conf.GitFlags = last(value)

		case editor:
			conf.Conf.Editor = last(value)
		case editorFlags:
			conf.Conf.EditorFlags = last(value)

		case mFlags:
			conf.Conf.MFlags = last(value)
		case makepkg:
			conf.Conf.MakepkgBin = last(value)
		case makePkgconf:
			conf.Conf.MakepkgConf = last(value)
		case noMakePkgconf:
			conf.Conf.MakepkgConf = ""

		case pacman:
			conf.Conf.PacmanBin = last(value)

		case sudo:
			conf.Conf.SudoBin = last(value)
		case sudoFlags:
			conf.Conf.SudoFlags = last(value)

		case sudoLoop:
			conf.Conf.SudoLoop = true
		case noSudoLoop:
			conf.Conf.SudoLoop = false

		case requestSplitN:
			n, _ := strconv.Atoi(last(value))
			if n > 0 {
				conf.Conf.RequestSplitN = n
			}

		case provides:
			conf.Conf.Provides = true
		case noProvides:
			conf.Conf.Provides = false

		case pgpFetch:
			conf.Conf.PGPFetch = true
		case noPGPFetch:
			conf.Conf.PGPFetch = false

		case upgradeMenu:
			conf.Conf.UpgradeMenu = true
		case noUpgradeMenu:
			conf.Conf.UpgradeMenu = false
		case cleanMenu:
			conf.Conf.CleanMenu = true
		case noCleanMenu:
			conf.Conf.CleanMenu = false
		case diffMenu:
			conf.Conf.DiffMenu = true
		case noDiffMenu:
			conf.Conf.DiffMenu = false
		case editMenu:
			conf.Conf.EditMenu = true
		case noEditMenu:
			conf.Conf.EditMenu = false

		case useAsk:
			conf.Conf.UseAsk = true
		case noUseAsk:
			conf.Conf.UseAsk = false

		case combinedUpgrade:
			conf.Conf.CombinedUpgrade = true
		case noCombinedUpgrade:
			conf.Conf.CombinedUpgrade = false

		case removeMake:
			conf.Conf.RemoveMake = "yes"
		case noRemoveMake:
			conf.Conf.RemoveMake = "no"
		case askRemoveMake:
			conf.Conf.RemoveMake = "ask"

		// -- Yay Show Options --

		case complete:
			conf.ModeConf.(*PConf).Complete = true
		case defaultConfig:
			conf.ModeConf.(*PConf).DefaultConfig = true
		case currentConfig:
			conf.ModeConf.(*PConf).CurrentConfig = true
		case stats:
			conf.ModeConf.(*PConf).LocalStats = true
		case news:
			conf.ModeConf.(*PConf).News = true

		// -- Yay yay-mode Options --

		case yayClean:
			conf.ModeConf.(*YConf).Clean = true
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

		case ask: // empty
		}
		return true
	}
}

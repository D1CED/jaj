package settings

import (
	"strconv"
	"strings"

	rpc "github.com/mikkeloscar/aur"

	"github.com/Jguer/yay/v10/pkg/settings/parser"
)

var isArg = []string{
	"-", "--",
	"ask",
	"D", "database",
	"Q", "query",
	"R", "remove",
	"S", "sync",
	"T", "deptest",
	"U", "upgrade",
	"F", "files",
	"V", "version",
	"h", "help",
	"Y", "yay",
	"P", "show",
	"G", "getpkgbuild",
	"b", "dbpath",
	"r", "root",
	"v", "verbose",
	"arch",
	"cachedir",
	"color",
	"config",
	"debug",
	"gpgdir",
	"hookdir",
	"logfile",
	"noconfirm",
	"confirm",
	"disable-download-timeout",
	"sysroot",
	"d", "nodeps",
	"assume-installed",
	"dbonly",
	"absdir",
	"noprogressbar",
	"noscriptlet",
	"p", "print",
	"print-format",
	"asdeps",
	"asexplicit",
	"ignore",
	"ignoregroup",
	"needed",
	"overwrite",
	"f", "force",
	"c", "changelog",
	"deps",
	"e", "explicit",
	"g", "groups",
	"i", "info",
	"k", "check",
	"l", "list",
	"m", "foreign",
	"n", "native",
	"o", "owns",
	"file",
	"q", "quiet",
	"s", "search",
	"t", "unrequired",
	"u", "upgrades",
	"cascade",
	"nosave",
	"recursive",
	"unneeded",
	"clean",
	"sysupgrade",
	"w", "downloadonly",
	"y", "refresh",
	"x", "regex",
	"machinereadable",
	// yay options
	"aururl",
	"save",
	"afterclean", "cleanafter",
	"noafterclean", "nocleanafter",
	"devel",
	"nodevel",
	"timeupdate",
	"notimeupdate",
	"topdown",
	"bottomup",
	"completioninterval",
	"sortby",
	"searchby",
	"redownload",
	"redownloadall",
	"noredownload",
	"rebuild",
	"rebuildall",
	"rebuildtree",
	"norebuild",
	"batchinstall",
	"nobatchinstall",
	"answerclean",
	"noanswerclean",
	"answerdiff",
	"noanswerdiff",
	"answeredit",
	"noansweredit",
	"answerupgrade",
	"noanswerupgrade",
	"gpgflags",
	"mflags",
	"gitflags",
	"builddir",
	"editor",
	"editorflags",
	"makepkg",
	"makepkgconf",
	"nomakepkgconf",
	"pacman",
	"git",
	"gpg",
	"sudo",
	"sudoflags",
	"requestsplitn",
	"sudoloop",
	"nosudoloop",
	"provides",
	"noprovides",
	"pgpfetch",
	"nopgpfetch",
	"upgrademenu",
	"noupgrademenu",
	"cleanmenu",
	"nocleanmenu",
	"diffmenu",
	"nodiffmenu",
	"editmenu",
	"noeditmenu",
	"useask",
	"nouseask",
	"combinedupgrade",
	"nocombinedupgrade",
	"a", "aur",
	"repo",
	"removemake",
	"noremovemake",
	"askremovemake",
	"complete",
	"stats",
	"news",
	"gendb",
	"currentconfig",
}

var isOp = []string{
	"V", "version",
	"D", "database",
	"F", "files",
	"Q", "query",
	"R", "remove",
	"S", "sync",
	"T", "deptest",
	"U", "upgrade",
	// yay specific
	"Y", "yay",
	"P", "show",
	"G", "getpkgbuild",
}

var isGlobal = []string{
	"b", "dbpath",
	"r", "root",
	"v", "verbose",
	"arch",
	"cachedir",
	"color",
	"config",
	"debug",
	"gpgdir",
	"hookdir",
	"logfile",
	"noconfirm",
	"confirm",
}

var hasParam = []string{
	"dbpath", "b",
	"root", "r",
	"sysroot",
	"config",
	"ignore",
	"assume-installed",
	"overwrite",
	"ask",
	"cachedir",
	"hookdir",
	"logfile",
	"ignoregroup",
	"arch",
	"print-format",
	"gpgdir",
	"color",
	// yay params
	"aururl",
	"mflags",
	"gpgflags",
	"gitflags",
	"builddir",
	"absdir",
	"editor",
	"editorflags",
	"makepkg",
	"makepkgconf",
	"pacman",
	"git",
	"gpg",
	"sudo",
	"sudoflags",
	"requestsplitn",
	"answerclean",
	"answerdiff",
	"answeredit",
	"answerupgrade",
	"completioninterval",
	"sortby",
	"searchby",
}

func NewFlagParser() *parser.Arguments {
	return parser.New(isArg, isOp, isGlobal, hasParam)
}

func NeedRoot(a *parser.Arguments, mode TargetMode) bool {
	if a.ExistsArg("h", "help") {
		return false
	}

	switch a.Op {
	case "D", "database":
		if a.ExistsArg("k", "check") {
			return false
		}
		return true
	case "F", "files":
		if a.ExistsArg("y", "refresh") {
			return true
		}
		return false
	case "Q", "query":
		if a.ExistsArg("k", "check") {
			return true
		}
		return false
	case "R", "remove":
		if a.ExistsArg("p", "print", "print-format") {
			return false
		}
		return true
	case "S", "sync":
		if a.ExistsArg("y", "refresh") {
			return true
		}
		if a.ExistsArg("p", "print", "print-format") {
			return false
		}
		if a.ExistsArg("s", "search") {
			return false
		}
		if a.ExistsArg("l", "list") {
			return false
		}
		if a.ExistsArg("g", "groups") {
			return false
		}
		if a.ExistsArg("i", "info") {
			return false
		}
		if a.ExistsArg("c", "clean") && mode == ModeAUR {
			return false
		}
		return true
	case "U", "upgrade":
		return true
	default:
		return false
	}
}

func handleConfig(config *Configuration, addFlags *AdditionalFlags, option, value string) bool {
	switch option {
	case "aururl":
		config.AURURL = value
	case "save":
		addFlags.SaveConfig = true
	case "afterclean", "cleanafter":
		config.CleanAfter = true
	case "noafterclean", "nocleanafter":
		config.CleanAfter = false
	case "devel":
		config.Devel = true
	case "nodevel":
		config.Devel = false
	case "timeupdate":
		config.TimeUpdate = true
	case "notimeupdate":
		config.TimeUpdate = false
	case "topdown":
		config.SortMode = TopDown
	case "bottomup":
		config.SortMode = BottomUp
	case "completioninterval":
		n, err := strconv.Atoi(value)
		if err == nil {
			config.CompletionInterval = n
		}
	case "sortby":
		config.SortBy = value
	case "searchby":
		config.SearchBy = value
	case "noconfirm":
		NoConfirm = true
	case "config":
		config.PacmanConf = value
	case "redownload":
		config.ReDownload = "yes"
	case "redownloadall":
		config.ReDownload = "all"
	case "noredownload":
		config.ReDownload = "no"
	case "rebuild":
		config.ReBuild = "yes"
	case "rebuildall":
		config.ReBuild = "all"
	case "rebuildtree":
		config.ReBuild = "tree"
	case "norebuild":
		config.ReBuild = "no"
	case "batchinstall":
		config.BatchInstall = true
	case "nobatchinstall":
		config.BatchInstall = false
	case "answerclean":
		config.AnswerClean = value
	case "noanswerclean":
		config.AnswerClean = ""
	case "answerdiff":
		config.AnswerDiff = value
	case "noanswerdiff":
		config.AnswerDiff = ""
	case "answeredit":
		config.AnswerEdit = value
	case "noansweredit":
		config.AnswerEdit = ""
	case "answerupgrade":
		config.AnswerUpgrade = value
	case "noanswerupgrade":
		config.AnswerUpgrade = ""
	case "gpgflags":
		config.GpgFlags = value
	case "mflags":
		config.MFlags = value
	case "gitflags":
		config.GitFlags = value
	case "builddir":
		config.BuildDir = value
	case "absdir":
		config.ABSDir = value
	case "editor":
		config.Editor = value
	case "editorflags":
		config.EditorFlags = value
	case "makepkg":
		config.MakepkgBin = value
	case "makepkgconf":
		config.MakepkgConf = value
	case "nomakepkgconf":
		config.MakepkgConf = ""
	case "pacman":
		config.PacmanBin = value
	case "git":
		config.GitBin = value
	case "gpg":
		config.GpgBin = value
	case "sudo":
		config.SudoBin = value
	case "sudoflags":
		config.SudoFlags = value
	case "requestsplitn":
		n, err := strconv.Atoi(value)
		if err == nil && n > 0 {
			config.RequestSplitN = n
		}
	case "sudoloop":
		config.SudoLoop = true
	case "nosudoloop":
		config.SudoLoop = false
	case "provides":
		config.Provides = true
	case "noprovides":
		config.Provides = false
	case "pgpfetch":
		config.PGPFetch = true
	case "nopgpfetch":
		config.PGPFetch = false
	case "upgrademenu":
		config.UpgradeMenu = true
	case "noupgrademenu":
		config.UpgradeMenu = false
	case "cleanmenu":
		config.CleanMenu = true
	case "nocleanmenu":
		config.CleanMenu = false
	case "diffmenu":
		config.DiffMenu = true
	case "nodiffmenu":
		config.DiffMenu = false
	case "editmenu":
		config.EditMenu = true
	case "noeditmenu":
		config.EditMenu = false
	case "useask":
		config.UseAsk = true
	case "nouseask":
		config.UseAsk = false
	case "combinedupgrade":
		config.CombinedUpgrade = true
	case "nocombinedupgrade":
		config.CombinedUpgrade = false
	case "a", "aur":
		addFlags.Mode = ModeAUR
	case "repo":
		addFlags.Mode = ModeRepo
	case "removemake":
		config.RemoveMake = "yes"
	case "noremovemake":
		config.RemoveMake = "no"
	case "askremovemake":
		config.RemoveMake = "ask"
	default:
		return false
	}

	return true
}

func ParseCommandLine(a *parser.Arguments, args []string, config *Configuration, addFlags *AdditionalFlags) error {
	if len(args) == 0 {
		args = []string{"-Syu"}
	}

	err := a.Parse(args)
	if err != nil {
		return err
	}

	extractYayOptions(a, config, addFlags)

	return nil
}

func extractYayOptions(a *parser.Arguments, config *Configuration, addFlags *AdditionalFlags) {
	for option, value := range a.Options {
		if handleConfig(config, addFlags, option, parser.First(value)) {
			a.DelArg(option)
		}
	}

	rpc.AURURL = strings.TrimRight(config.AURURL, "/") + "/rpc.php?"
	config.AURURL = strings.TrimRight(config.AURURL, "/")
}

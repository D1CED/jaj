package settings

import (
	"os"
	"strconv"
	"strings"

	rpc "github.com/mikkeloscar/aur"
)

func NeedRoot(a *Arguments, mode TargetMode) bool {
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

func isArg(arg string) bool {
	switch arg {
	case "-", "--":
	case "ask":
	case "D", "database":
	case "Q", "query":
	case "R", "remove":
	case "S", "sync":
	case "T", "deptest":
	case "U", "upgrade":
	case "F", "files":
	case "V", "version":
	case "h", "help":
	case "Y", "yay":
	case "P", "show":
	case "G", "getpkgbuild":
	case "b", "dbpath":
	case "r", "root":
	case "v", "verbose":
	case "arch":
	case "cachedir":
	case "color":
	case "config":
	case "debug":
	case "gpgdir":
	case "hookdir":
	case "logfile":
	case "noconfirm":
	case "confirm":
	case "disable-download-timeout":
	case "sysroot":
	case "d", "nodeps":
	case "assume-installed":
	case "dbonly":
	case "absdir":
	case "noprogressbar":
	case "noscriptlet":
	case "p", "print":
	case "print-format":
	case "asdeps":
	case "asexplicit":
	case "ignore":
	case "ignoregroup":
	case "needed":
	case "overwrite":
	case "f", "force":
	case "c", "changelog":
	case "deps":
	case "e", "explicit":
	case "g", "groups":
	case "i", "info":
	case "k", "check":
	case "l", "list":
	case "m", "foreign":
	case "n", "native":
	case "o", "owns":
	case "file":
	case "q", "quiet":
	case "s", "search":
	case "t", "unrequired":
	case "u", "upgrades":
	case "cascade":
	case "nosave":
	case "recursive":
	case "unneeded":
	case "clean":
	case "sysupgrade":
	case "w", "downloadonly":
	case "y", "refresh":
	case "x", "regex":
	case "machinereadable":
	// yay options
	case "aururl":
	case "save":
	case "afterclean", "cleanafter":
	case "noafterclean", "nocleanafter":
	case "devel":
	case "nodevel":
	case "timeupdate":
	case "notimeupdate":
	case "topdown":
	case "bottomup":
	case "completioninterval":
	case "sortby":
	case "searchby":
	case "redownload":
	case "redownloadall":
	case "noredownload":
	case "rebuild":
	case "rebuildall":
	case "rebuildtree":
	case "norebuild":
	case "batchinstall":
	case "nobatchinstall":
	case "answerclean":
	case "noanswerclean":
	case "answerdiff":
	case "noanswerdiff":
	case "answeredit":
	case "noansweredit":
	case "answerupgrade":
	case "noanswerupgrade":
	case "gpgflags":
	case "mflags":
	case "gitflags":
	case "builddir":
	case "editor":
	case "editorflags":
	case "makepkg":
	case "makepkgconf":
	case "nomakepkgconf":
	case "pacman":
	case "git":
	case "gpg":
	case "sudo":
	case "sudoflags":
	case "requestsplitn":
	case "sudoloop":
	case "nosudoloop":
	case "provides":
	case "noprovides":
	case "pgpfetch":
	case "nopgpfetch":
	case "upgrademenu":
	case "noupgrademenu":
	case "cleanmenu":
	case "nocleanmenu":
	case "diffmenu":
	case "nodiffmenu":
	case "editmenu":
	case "noeditmenu":
	case "useask":
	case "nouseask":
	case "combinedupgrade":
	case "nocombinedupgrade":
	case "a", "aur":
	case "repo":
	case "removemake":
	case "noremovemake":
	case "askremovemake":
	case "complete":
	case "stats":
	case "news":
	case "gendb":
	case "currentconfig":
	default:
		return false
	}

	return true
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

func isOp(op string) bool {
	switch op {
	case "V", "version":
	case "D", "database":
	case "F", "files":
	case "Q", "query":
	case "R", "remove":
	case "S", "sync":
	case "T", "deptest":
	case "U", "upgrade":
	// yay specific
	case "Y", "yay":
	case "P", "show":
	case "G", "getpkgbuild":
	default:
		return false
	}

	return true
}

func isGlobal(op string) bool {
	switch op {
	case "b", "dbpath":
	case "r", "root":
	case "v", "verbose":
	case "arch":
	case "cachedir":
	case "color":
	case "config":
	case "debug":
	case "gpgdir":
	case "hookdir":
	case "logfile":
	case "noconfirm":
	case "confirm":
	default:
		return false
	}

	return true
}

func hasParam(arg string) bool {
	switch arg {
	case "dbpath", "b":
	case "root", "r":
	case "sysroot":
	case "config":
	case "ignore":
	case "assume-installed":
	case "overwrite":
	case "ask":
	case "cachedir":
	case "hookdir":
	case "logfile":
	case "ignoregroup":
	case "arch":
	case "print-format":
	case "gpgdir":
	case "color":
	// yay params
	case "aururl":
	case "mflags":
	case "gpgflags":
	case "gitflags":
	case "builddir":
	case "absdir":
	case "editor":
	case "editorflags":
	case "makepkg":
	case "makepkgconf":
	case "pacman":
	case "git":
	case "gpg":
	case "sudo":
	case "sudoflags":
	case "requestsplitn":
	case "answerclean":
	case "answerdiff":
	case "answeredit":
	case "answerupgrade":
	case "completioninterval":
	case "sortby":
	case "searchby":
	default:
		return false
	}

	return true
}

func (a *Arguments) ParseCommandLine(config *Configuration, addFlags *AdditionalFlags) error {
	args := os.Args[1:]
	usedNext := false

	if len(args) < 1 {
		if _, err := a.parseShortOption("-Syu", ""); err != nil {
			return err
		}
	} else {
		for k, arg := range args {
			var nextArg string

			if usedNext {
				usedNext = false
				continue
			}

			if k+1 < len(args) {
				nextArg = args[k+1]
			}

			var err error
			switch {
			case a.ExistsArg("--"):
				a.AddTarget(arg)
			case strings.HasPrefix(arg, "--"):
				usedNext, err = a.parseLongOption(arg, nextArg)
			case strings.HasPrefix(arg, "-"):
				usedNext, err = a.parseShortOption(arg, nextArg)
			default:
				a.AddTarget(arg)
			}

			if err != nil {
				return err
			}
		}
	}

	if a.Op == "" {
		a.Op = "Y"
	}

	if a.ExistsArg("-") {
		if err := a.parseStdin(); err != nil {
			return err
		}
		a.DelArg("-")

		file, err := os.Open("/dev/tty")
		if err != nil {
			return err
		}

		os.Stdin = file
	}

	extractYayOptions(a, config, addFlags)

	return nil
}

func extractYayOptions(a *Arguments, config *Configuration, addFlags *AdditionalFlags) {
	for option, value := range a.Options {
		if handleConfig(config, addFlags, option, value.First()) {
			a.DelArg(option)
		}
	}

	rpc.AURURL = strings.TrimRight(config.AURURL, "/") + "/rpc.php?"
	config.AURURL = strings.TrimRight(config.AURURL, "/")
}

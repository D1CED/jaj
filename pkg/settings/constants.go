package settings

import "github.com/Jguer/yay/v10/pkg/settings/parser"

const Usage = `Usage:
    yay
    yay <operation> [...]
    yay <package(s)>

operations:
    yay {-h --help}
    yay {-V --version}
    yay {-D --database}    <options> <package(s)>
    yay {-F --files}       [options] [package(s)]
    yay {-Q --query}       [options] [package(s)]
    yay {-R --remove}      [options] <package(s)>
    yay {-S --sync}        [options] [package(s)]
    yay {-T --deptest}     [options] [package(s)]
    yay {-U --upgrade}     [options] <file(s)>

New operations:
    yay {-Y --yay}         [options] [package(s)]
    yay {-P --show}        [options]
    yay {-G --getpkgbuild} [package(s)]

If no arguments are provided 'yay -Syu' will be performed
If no operation is provided -Y will be assumed

New options:
       --repo             Assume targets are from the repositories
    -a --aur              Assume targets are from the AUR

Permanent configuration options:
    --save                Causes the following options to be saved back to the
                          config file when used

    --aururl      <url>   Set an alternative AUR URL
    --builddir    <dir>   Directory used to download and run PKGBUILDS
    --absdir      <dir>   Directory used to store downloads from the ABS
    --editor      <file>  Editor to use when editing PKGBUILDs
    --editorflags <flags> Pass arguments to editor
    --makepkg     <file>  makepkg command to use
    --mflags      <flags> Pass arguments to makepkg
    --pacman      <file>  pacman command to use
    --git         <file>  git command to use
    --gitflags    <flags> Pass arguments to git
    --gpg         <file>  gpg command to use
    --gpgflags    <flags> Pass arguments to gpg
    --config      <file>  pacman.conf file to use
    --makepkgconf <file>  makepkg.conf file to use
    --nomakepkgconf       Use the default makepkg.conf

    --requestsplitn <n>   Max amount of packages to query per AUR request
    --completioninterval  <n> Time in days to refresh completion cache
    --sortby    <field>   Sort AUR results by a specific field during search
    --searchby  <field>   Search for packages using a specified field
    --answerclean   <a>   Set a predetermined answer for the clean build menu
    --answerdiff    <a>   Set a predetermined answer for the diff menu
    --answeredit    <a>   Set a predetermined answer for the edit pkgbuild menu
    --answerupgrade <a>   Set a predetermined answer for the upgrade menu
    --noanswerclean       Unset the answer for the clean build menu
    --noanswerdiff        Unset the answer for the edit diff menu
    --noansweredit        Unset the answer for the edit pkgbuild menu
    --noanswerupgrade     Unset the answer for the upgrade menu
    --cleanmenu           Give the option to clean build PKGBUILDS
    --diffmenu            Give the option to show diffs for build files
    --editmenu            Give the option to edit/view PKGBUILDS
    --upgrademenu         Show a detailed list of updates with the option to skip any
    --nocleanmenu         Don't clean build PKGBUILDS
    --nodiffmenu          Don't show diffs for build files
    --noeditmenu          Don't edit/view PKGBUILDS
    --noupgrademenu       Don't show the upgrade menu
    --askremovemake       Ask to remove makedepends after install
    --removemake          Remove makedepends after install
    --noremovemake        Don't remove makedepends after install

    --cleanafter          Remove package sources after successful install
    --nocleanafter        Do not remove package sources after successful build
    --bottomup            Shows AUR's packages first and then repository's
    --topdown             Shows repository's packages first and then AUR's

    --devel               Check development packages during sysupgrade
    --nodevel             Do not check development packages
    --rebuild             Always build target packages
    --rebuildall          Always build all AUR packages
    --norebuild           Skip package build if in cache and up to date
    --rebuildtree         Always build all AUR packages even if installed
    --redownload          Always download pkgbuilds of targets
    --noredownload        Skip pkgbuild download if in cache and up to date
    --redownloadall       Always download pkgbuilds of all AUR packages
    --provides            Look for matching providers when searching for packages
    --noprovides          Just look for packages by pkgname
    --pgpfetch            Prompt to import PGP keys from PKGBUILDs
    --nopgpfetch          Don't prompt to import PGP keys
    --useask              Automatically resolve conflicts using pacman's ask flag
    --nouseask            Confirm conflicts manually during the install
    --combinedupgrade     Refresh then perform the repo and AUR upgrade together
    --nocombinedupgrade   Perform the repo upgrade and AUR upgrade separately
    --batchinstall        Build multiple AUR packages then install them together
    --nobatchinstall      Build and install each AUR package one by one

    --sudo                <file>  sudo command to use
    --sudoflags           <flags> Pass arguments to sudo
    --sudoloop            Loop sudo calls in the background to avoid timeout
    --nosudoloop          Do not loop sudo calls in the background

    --timeupdate          Check packages' AUR page for changes during sysupgrade
    --notimeupdate        Do not check packages' AUR page for changes

show specific options:
    -c --complete         Used for completions
    -d --defaultconfig    Print default yay configuration
    -g --currentconfig    Print current yay configuration
    -s --stats            Display system package statistics
    -w --news             Print arch news

yay specific options:
    -c --clean            Remove unneeded dependencies
       --gendb            Generates development package DB used for updating

getpkgbuild specific options:
    -f --force            Force download for existing ABS packages
`

const (
	// Pacman Operations
	database parser.Enum = iota
	query
	remove
	sync
	depTest
	upgrade
	files

	// Both Operations
	version
	help

	// YAY Operations
	yay
	show
	getPkgbuild

	// Pacman Global Options
	dbPath
	root
	verbose
	arch
	cacheDir
	color
	config
	debug
	gpgDir
	hookDir
	logFile
	noConfirm
	confirm
	disableDownloadTimeout
	sysRoot

	// Pacman Transaction Options (SRU)
	noDeps
	assumeInstalled
	dbOnly
	noProgressbar
	noScriptlet
	print
	printFormat // implies Print

	// Pacman Upgrade Options (SU)
	asDeps
	asExplicit
	ignore
	ignoreGroup
	needed
	overwrite

	// Pacman Query Options (Q)
	changelog
	deps
	explicit
	groups
	info
	check
	list
	foreign
	native
	owns
	file
	quiet
	search
	unrequired
	upgrades

	// Pacman Remove Options (R)
	cascade
	noSave
	recursive
	unneeded

	// Pacman Sync Options (S)
	clean
	// groups
	// info
	// list
	// quiet
	// search
	sysUpgrade
	downloadOnly
	refresh

	// Pacman Database Options (D)
	// asDeps
	// asExplicit
	// check
	// quiet

	// Pacman File Options (F)
	// refresh
	// list
	regex
	// quiet
	machineReadable

	// Yay persistent options
	aurURL
	buildDir
	absDir
	editor
	editorFlags
	makepkg
	makePkgconf
	noMakePkgconf
	pacman
	// pacmanConf see Config
	redownload
	redownloadAll
	noRedownload
	rebuild
	rebuildAll
	rebuildTree
	noRebuild
	answerClean
	noAnswerClean
	answerDiff
	noAnswerDiff
	answerEdit
	noAnswerEdit
	answerUpgrade
	noAnswerUpgrade
	git
	gpg
	gpgFlags
	mFlags
	sortBy
	searchBy
	gitFlags
	removeMake
	noRemoveMake
	askRemoveMake
	sudo
	sudoFlags
	requestSplitN
	topdown  // sort mode
	bottomup // sort mode
	completionInterval
	sudoLoop
	noSudoLoop
	timeUpdate
	noTimeUpdate
	devel
	noDevel
	cleanAfter
	noCleanAfter
	provides
	noProvides
	pgpFetch
	noPGPFetch
	upgradeMenu
	noUpgradeMenu
	cleanMenu
	noCleanMenu
	diffMenu
	noDiffMenu
	editMenu
	noEditMenu
	combinedUpgrade
	noCombinedUpgrade
	useAsk
	noUseAsk
	batchInstall
	noBatchInstall

	// Yay Show options (P)
	complete
	defaultConfig
	currentConfig
	stats
	news
	numberUpgrades // deprecated

	// Yay yay-mode options (Y)
	yayClean
	genDB

	// Yay GetPkgbuild options (G)
	force

	// Mode
	aur
	repo
	// any

	// misc options
	save

	// unused yay op
	ask
)

func mapping(option string, mainOp OpMode) parser.Enum {
	switch option {
	default:
		return parser.InvalidFlag

	// ambiguous flags
	case "d":
		if mainOp == OpQuery {
			return deps
		}
		if mainOp == OpShow {
			return defaultConfig
		}
		return noDeps
	case "p":
		if mainOp == OpQuery {
			return file
		}
		return print
	case "c":
		if mainOp == OpQuery {
			return changelog
		}
		if mainOp == OpRemove {
			return cascade
		}
		if mainOp == OpSync {
			return clean
		}
		if mainOp == OpShow {
			return complete
		}
		if mainOp == OpYay {
			return clean
		}
		return parser.InvalidFlag
	case "u":
		if mainOp == OpQuery {
			return upgrades
		}
		if mainOp == OpRemove {
			return unneeded
		}
		if mainOp == OpSync {
			return sysUpgrade
		}
		return parser.InvalidFlag
	case "n":
		if mainOp == OpShow {
			return numberUpgrades
		}
		if mainOp == OpQuery {
			return native
		}
		if mainOp == OpRemove {
			return noSave
		}
		return parser.InvalidFlag
	case "s":
		if mainOp == OpQuery {
			return search
		}
		if mainOp == OpRemove {
			return recursive
		}
		if mainOp == OpShow {
			return stats
		}
		return parser.InvalidFlag
	case "g":
		if mainOp == OpShow {
			return currentConfig
		}
		return groups
	case "w":
		if mainOp == OpShow {
			return news
		}
		return downloadOnly
	case "clean":
		if mainOp == OpYay {
			return yayClean
		}
		return clean

	case "D", "database":
		return database
	case "Q", "query":
		return query
	case "R", "remove":
		return remove
	case "S", "sync":
		return sync
	case "T", "deptest":
		return depTest
	case "U", "upgrade":
		return upgrade
	case "F", "files":
		return files
	case "V", "version":
		return version
	case "h", "help":
		return help
	case "Y", "yay":
		return yay
	case "P", "show":
		return show
	case "G", "getpkgbuild":
		return getPkgbuild
	case "b", "dbpath":
		return dbPath
	case "r", "root":
		return root
	case "v", "verbose":
		return verbose
	case "arch":
		return arch
	case "cachedir":
		return cacheDir
	case "color":
		return color
	case "config":
		return config
	case "debug":
		return debug
	case "gpgdir":
		return gpgDir
	case "hookdir":
		return hookDir
	case "logfile":
		return logFile
	case "noconfirm":
		return noConfirm
	case "confirm":
		return confirm
	case "disable-download-timeout":
		return disableDownloadTimeout
	case "sysroot":
		return sysRoot
	case "nodeps":
		return noDeps
	case "assume-installed":
		return assumeInstalled
	case "dbonly":
		return dbOnly
	case "absdir":
		return absDir
	case "noprogressbar":
		return noProgressbar
	case "noscriptlet":
		return noScriptlet
	case "print":
		return print
	case "print-format":
		return printFormat
	case "asdeps":
		return asDeps
	case "asexplicit":
		return asExplicit
	case "ignore":
		return ignore
	case "ignoregroup":
		return ignoreGroup
	case "needed":
		return needed
	case "overwrite":
		return overwrite
	case "f", "force":
		return force
	case "changelog":
		return changelog
	case "deps":
		return deps
	case "e", "explicit":
		return explicit
	case "groups":
		return groups
	case "i", "info":
		return info
	case "k", "check":
		return check
	case "l", "list":
		return list
	case "m", "foreign":
		return foreign
	case "native":
		return native
	case "o", "owns":
		return owns
	case "file":
		return file
	case "q", "quiet":
		return quiet
	case "search":
		return search
	case "t", "unrequired":
		return unrequired
	case "upgrades":
		return upgrades
	case "cascade":
		return cascade
	case "nosave":
		return noSave
	case "recursive":
		return recursive
	case "unneeded":
		return unneeded
	case "sysupgrade":
		return sysUpgrade
	case "downloadonly":
		return downloadOnly
	case "y", "refresh":
		return refresh
	case "x", "regex":
		return regex
	case "machinereadable":
		return machineReadable
	// yay options
	case "aururl":
		return aurURL
	case "save":
		return save
	case "afterclean", "cleanafter":
		return cleanAfter
	case "noafterclean", "nocleanafter":
		return noCleanAfter
	case "devel":
		return devel
	case "nodevel":
		return noDevel
	case "timeupdate":
		return timeUpdate
	case "notimeupdate":
		return noTimeUpdate
	case "topdown":
		return topdown
	case "bottomup":
		return bottomup
	case "completioninterval":
		return completionInterval
	case "sortby":
		return sortBy
	case "searchby":
		return searchBy
	case "redownload":
		return redownload
	case "redownloadall":
		return redownloadAll
	case "noredownload":
		return noRedownload
	case "rebuild":
		return rebuild
	case "rebuildall":
		return rebuildAll
	case "rebuildtree":
		return rebuildTree
	case "norebuild":
		return noRebuild
	case "batchinstall":
		return batchInstall
	case "nobatchinstall":
		return noBatchInstall
	case "answerclean":
		return answerClean
	case "noanswerclean":
		return noAnswerClean
	case "answerdiff":
		return answerDiff
	case "noanswerdiff":
		return noAnswerDiff
	case "answeredit":
		return answerEdit
	case "noansweredit":
		return noAnswerEdit
	case "answerupgrade":
		return answerUpgrade
	case "noanswerupgrade":
		return noAnswerUpgrade
	case "gpgflags":
		return gpgFlags
	case "mflags":
		return mFlags
	case "gitflags":
		return gitFlags
	case "builddir":
		return buildDir
	case "editor":
		return editor
	case "editorflags":
		return editorFlags
	case "makepkg":
		return makepkg
	case "makepkgconf":
		return makePkgconf
	case "nomakepkgconf":
		return noMakePkgconf
	case "pacman":
		return pacman
	case "git":
		return git
	case "gpg":
		return gpg
	case "sudo":
		return sudo
	case "sudoflags":
		return sudoFlags
	case "requestsplitn":
		return requestSplitN
	case "sudoloop":
		return sudoLoop
	case "nosudoloop":
		return noSudoLoop
	case "provides":
		return provides
	case "noprovides":
		return noProvides
	case "pgpfetch":
		return pgpFetch
	case "nopgpfetch":
		return noPGPFetch
	case "upgrademenu":
		return upgradeMenu
	case "noupgrademenu":
		return noUpgradeMenu
	case "cleanmenu":
		return cleanMenu
	case "nocleanmenu":
		return noCleanMenu
	case "diffmenu":
		return diffMenu
	case "nodiffmenu":
		return noDiffMenu
	case "editmenu":
		return editMenu
	case "noeditmenu":
		return noEditMenu
	case "useask":
		return useAsk
	case "nouseask":
		return noUseAsk
	case "combinedupgrade":
		return combinedUpgrade
	case "nocombinedupgrade":
		return noCombinedUpgrade
	case "a", "aur":
		return aur
	case "repo":
		return repo
	case "removemake":
		return removeMake
	case "noremovemake":
		return noRemoveMake
	case "askremovemake":
		return askRemoveMake
	case "complete":
		return complete
	case "stats":
		return stats
	case "news":
		return news
	case "gendb":
		return genDB
	case "defaultconfig":
		return defaultConfig
	case "currentconfig":
		return currentConfig
	case "numberupgrades":
		return numberUpgrades
	case "ask":
		return ask
	}
}

func takesParam(e parser.Enum, mainOp OpMode) bool {
	if (e == asDeps || e == asExplicit) && (mainOp == OpSync || mainOp == OpUpgrade) {
		return false
	}
	for _, val := range hasParam {
		if val == e {
			return true
		}
	}
	return false
}

// options that take a parameter
var hasParam = []parser.Enum{
	dbPath,   // path
	root,     // path
	arch,     // arch
	cacheDir, // dir
	color,    // 'always' | 'never' | 'auto'
	config,   // file
	gpgDir,   // dir
	hookDir,  // dir
	logFile,  // file
	sysRoot,  // dir

	assumeInstalled, // package=version
	printFormat,     // format
	ignore,          // package
	ignoreGroup,     // group
	overwrite,       // glob
	owns,            // file
	search,          // regexp

	asDeps, // package | '' ; yes if -D no if -S/-U
	asExplicit,

	// yay params
	aurURL,             // url
	mFlags,             // makepkg flags
	gpgFlags,           // gpg flags
	gitFlags,           // git flags
	buildDir,           // dir
	absDir,             // dir
	editor,             // file
	editorFlags,        // flags
	makepkg,            // file
	makePkgconf,        // file
	pacman,             // file
	git,                // file
	gpg,                // file
	sudo,               // file
	sudoFlags,          // flags
	requestSplitN,      // int
	answerClean,        // answer
	answerDiff,         // answer
	answerEdit,         // answer
	answerUpgrade,      // answer
	completionInterval, // int (days)
	sortBy,             // field
	searchBy,           // field

	ask,
}

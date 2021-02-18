package main

import (
	"bufio"
	"net/http"

	alpm "github.com/Jguer/go-alpm/v2"

	"github.com/Jguer/yay/v10/pkg/completion"
	"github.com/Jguer/yay/v10/pkg/news"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/text"
)

func usage() {
	text.Println(`Usage:
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
    -f --force            Force download for existing ABS packages`)
}

func handleCmd(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	if cmdArgs.ExistsArg("h", "help") {
		return handleHelp(cmdArgs, rt)
	}

	if rt.Config.SudoLoop && settings.NeedRoot(cmdArgs, rt.Mode) {
		sudoLoopBackground(rt.CmdRunner, rt.Config)
	}

	switch cmdArgs.Op {
	case "V", "version":
		handleVersion()
		return nil
	case "D", "database":
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	case "F", "files":
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	case "Q", "query":
		return handleQuery(cmdArgs, rt)
	case "R", "remove":
		return handleRemove(cmdArgs, rt)
	case "S", "sync":
		return handleSync(cmdArgs, rt)
	case "T", "deptest":
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	case "U", "upgrade":
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	case "G", "getpkgbuild":
		return handleGetpkgbuild(cmdArgs, rt)
	case "P", "show":
		return handlePrint(cmdArgs, rt)
	case "Y", "--yay":
		return handleYay(cmdArgs, rt)
	}

	return text.ErrT("unhandled operation")
}

func handleQuery(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	if cmdArgs.ExistsArg("u", "upgrades") {
		return printUpdateList(cmdArgs, rt, cmdArgs.ExistsDouble("u", "sysupgrade"))
	}
	return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
}

func handleHelp(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	if cmdArgs.Op == "Y" || cmdArgs.Op == "yay" {
		usage()
		return nil
	}
	return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
}

func handleVersion() {
	text.Printf("yay v%s - libalpm v%s\n", yayVersion, alpm.Version())
}

func handlePrint(cmdArgs *settings.Arguments, rt *runtime.Runtime) (err error) {
	switch {
	case cmdArgs.ExistsArg("d", "defaultconfig"):
		tmpConfig := settings.DefaultConfig()
		text.Printf("%v", tmpConfig)
	case cmdArgs.ExistsArg("g", "currentconfig"):
		text.Printf("%v", rt.Config)
	case cmdArgs.ExistsArg("n", "numberupgrades"):
		err = printNumberOfUpdates(rt, cmdArgs.ExistsDouble("u", "sysupgrade"))
	case cmdArgs.ExistsArg("w", "news"):
		double := cmdArgs.ExistsDouble("w", "news")
		quiet := cmdArgs.ExistsArg("q", "quiet")
		err = news.PrintNewsFeed(rt.DB.LastBuildTime(), rt.Config.SortMode, double, quiet)
	case cmdArgs.ExistsDouble("c", "complete"):
		err = completion.Show(rt.DB, rt.Config.AURURL, rt.CompletionPath, rt.Config.CompletionInterval, true)
	case cmdArgs.ExistsArg("c", "complete"):
		err = completion.Show(rt.DB, rt.Config.AURURL, rt.CompletionPath, rt.Config.CompletionInterval, false)
	case cmdArgs.ExistsArg("s", "stats"):
		err = localStatistics(rt.DB, rt.Config.RequestSplitN)
	default:
		err = nil
	}
	return err
}

func handleYay(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	if cmdArgs.ExistsArg("gendb") {
		return createDevelDB(rt)
	}
	if cmdArgs.ExistsDouble("c") {
		return cleanDependencies(rt, cmdArgs, true)
	}
	if cmdArgs.ExistsArg("c", "clean") {
		return cleanDependencies(rt, cmdArgs, false)
	}
	if len(cmdArgs.Targets) > 0 {
		return handleYogurt(cmdArgs, rt)
	}
	return nil
}

func handleGetpkgbuild(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	return getPkgbuilds(cmdArgs.Targets, rt, cmdArgs.ExistsArg("f", "force"))
}

func handleYogurt(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	rt.Config.SearchMode = numberMenu
	return displayNumberMenu(cmdArgs.Targets, cmdArgs, rt)
}

func handleSync(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	targets := cmdArgs.Targets

	if cmdArgs.ExistsArg("s", "search") {
		if cmdArgs.ExistsArg("q", "quiet") {
			rt.Config.SearchMode = minimal
		} else {
			rt.Config.SearchMode = detailed
		}
		return syncSearch(targets, rt)
	}
	if cmdArgs.ExistsArg("p", "print", "print-format") {
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	}
	if cmdArgs.ExistsArg("c", "clean") {
		return syncClean(rt, cmdArgs)
	}
	if cmdArgs.ExistsArg("l", "list") {
		return syncList(cmdArgs, rt)
	}
	if cmdArgs.ExistsArg("g", "groups") {
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	}
	if cmdArgs.ExistsArg("i", "info") {
		return syncInfo(cmdArgs, targets, rt)
	}
	if cmdArgs.ExistsArg("u", "sysupgrade") {
		return install(cmdArgs, rt, false)
	}
	if len(cmdArgs.Targets) > 0 {
		return install(cmdArgs, rt, false)
	}
	if cmdArgs.ExistsArg("y", "refresh") {
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	}
	return nil
}

func handleRemove(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	err := rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	if err == nil {
		rt.VCSStore.RemovePackage(cmdArgs.Targets)
	}

	return err
}

// NumberMenu presents a CLI for selecting packages to install.
func displayNumberMenu(pkgS []string, cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	var (
		aurErr, repoErr error
		aq              aurQuery
		pq              repoQuery
		lenaq, lenpq    int
	)

	pkgS = query.RemoveInvalidTargets(pkgS, rt.Mode)

	if rt.Mode == settings.ModeAUR || rt.Mode == settings.ModeAny {
		aq, aurErr = narrowSearch(pkgS, true, rt.Config.SearchBy, rt.Config.SortBy)
		lenaq = len(aq)
	}
	if rt.Mode == settings.ModeRepo || rt.Mode == settings.ModeAny {
		pq = queryRepo(pkgS, rt.DB, rt.Config.SortMode)
		lenpq = len(pq)
		if repoErr != nil {
			return repoErr
		}
	}

	if lenpq == 0 && lenaq == 0 {
		return text.ErrT("no packages match search")
	}

	switch rt.Config.SortMode {
	case settings.TopDown:
		if rt.Mode == settings.ModeRepo || rt.Mode == settings.ModeAny {
			pq.printSearch(rt.DB, rt.Config.SearchMode, rt.Config.SortMode)
		}
		if rt.Mode == settings.ModeAUR || rt.Mode == settings.ModeAny {
			aq.printSearch(rt.DB, lenpq+1, rt.Config.SearchMode, rt.Config.SortMode)
		}
	case settings.BottomUp:
		if rt.Mode == settings.ModeAUR || rt.Mode == settings.ModeAny {
			aq.printSearch(rt.DB, lenpq+1, rt.Config.SearchMode, rt.Config.SortMode)
		}
		if rt.Mode == settings.ModeRepo || rt.Mode == settings.ModeAny {
			pq.printSearch(rt.DB, rt.Config.SearchMode, rt.Config.SortMode)
		}
	default:
		return text.ErrT("invalid sort mode. Fix with yay -Y --bottomup --save")
	}

	if aurErr != nil {
		text.Errorln(text.Tf("Error during AUR search: %s\n", aurErr))
		text.Warnln(text.T("Showing repo packages only"))
	}

	text.Infoln(text.T("Packages to install (eg: 1 2 3, 1-3 or ^4)"))
	text.Info()

	reader := bufio.NewReader(text.In())

	numberBuf, overflow, err := reader.ReadLine()
	if err != nil {
		return err
	}
	if overflow {
		return text.ErrT("input too long")
	}

	include, exclude, _, otherExclude := ParseNumberMenu(string(numberBuf))
	arguments := cmdArgs.CopyGlobal()

	isInclude := len(exclude) == 0 && len(otherExclude) == 0

	for i, pkg := range pq {
		var target int
		switch rt.Config.SortMode {
		case settings.TopDown:
			target = i + 1
		case settings.BottomUp:
			target = len(pq) - i
		default:
			return text.ErrT("invalid sort mode. Fix with yay -Y --bottomup --save")
		}

		if (isInclude && include.Get(target)) || (!isInclude && !exclude.Get(target)) {
			arguments.AddTarget(pkg.DB().Name() + "/" + pkg.Name())
		}
	}

	for i := range aq {
		var target int

		switch rt.Config.SortMode {
		case settings.TopDown:
			target = i + 1 + len(pq)
		case settings.BottomUp:
			target = len(aq) - i + len(pq)
		default:
			return text.ErrT("invalid sort mode. Fix with yay -Y --bottomup --save")
		}

		if (isInclude && include.Get(target)) || (!isInclude && !exclude.Get(target)) {
			arguments.AddTarget("aur/" + aq[i].Name)
		}
	}

	if len(arguments.Targets) == 0 {
		text.Println(text.T(" there is nothing to do"))
		return nil
	}

	if rt.Config.SudoLoop {
		sudoLoopBackground(rt.CmdRunner, rt.Config)
	}

	return install(arguments, rt, true)
}

func syncList(cmdArgs *settings.Arguments, rt *runtime.Runtime) error {
	aur := false

	for i := len(cmdArgs.Targets) - 1; i >= 0; i-- {
		if cmdArgs.Targets[i] == "aur" && (rt.Mode == settings.ModeAny || rt.Mode == settings.ModeAUR) {
			cmdArgs.Targets = append(cmdArgs.Targets[:i], cmdArgs.Targets[i+1:]...)
			aur = true
		}
	}

	if (rt.Mode == settings.ModeAny || rt.Mode == settings.ModeAUR) && (len(cmdArgs.Targets) == 0 || aur) {
		resp, err := http.Get(rt.Config.AURURL + "/packages.gz")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		scanner.Scan()
		for scanner.Scan() {
			name := scanner.Text()
			if cmdArgs.ExistsArg("q", "quiet") {
				text.Println(name)
			} else {
				text.Printf("%s %s %s", text.Magenta("aur"), text.Bold(name), text.Bold(text.Green(text.T("unknown-version"))))

				if rt.DB.LocalPackage(name) != nil {
					text.Print(text.Bold(text.Blue(text.T(" [Installed]"))))
				}

				text.Println()
			}
		}
	}

	if (rt.Mode == settings.ModeAny || rt.Mode == settings.ModeRepo) && (len(cmdArgs.Targets) != 0 || !aur) {
		return rt.CmdRunner.Show(passToPacman(rt, cmdArgs))
	}

	return nil
}

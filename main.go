package main // import "github.com/Jguer/yay"

import (
	"fmt"
	"os"

	pacmanconf "github.com/Morganamilo/go-pacmanconf"
	"golang.org/x/term"

	"github.com/Jguer/yay/v10/pkg/db/ialpm"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/parser"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/text"
)

// YayConf holds the current config values for yay.
// var rt *runtime.Runtime

func initAlpm(cmdArgs *parser.Arguments, pacmanConfigPath string) (*pacmanconf.Config, bool, error) {
	root := "/"
	if value, _, exists := cmdArgs.GetArg("root", "r"); exists {
		root = value
	}

	pacmanConf, stderr, err := pacmanconf.PacmanConf("--config", pacmanConfigPath, "--root", root)
	if err != nil {
		return nil, false, fmt.Errorf("%s", stderr)
	}

	if dbPath, _, exists := cmdArgs.GetArg("dbpath", "b"); exists {
		pacmanConf.DBPath = dbPath
	}

	if arch, _, exists := cmdArgs.GetArg("arch"); exists {
		pacmanConf.Architecture = arch
	}

	if ignoreArray := cmdArgs.GetArgs("ignore"); ignoreArray != nil {
		pacmanConf.IgnorePkg = append(pacmanConf.IgnorePkg, ignoreArray...)
	}

	if ignoreGroupsArray := cmdArgs.GetArgs("ignoregroup"); ignoreGroupsArray != nil {
		pacmanConf.IgnoreGroup = append(pacmanConf.IgnoreGroup, ignoreGroupsArray...)
	}

	if cacheArray := cmdArgs.GetArgs("cachedir"); cacheArray != nil {
		pacmanConf.CacheDir = cacheArray
	}

	if gpgDir, _, exists := cmdArgs.GetArg("gpgdir"); exists {
		pacmanConf.GPGDir = gpgDir
	}

	useColor := pacmanConf.Color && term.IsTerminal(int(os.Stdout.Fd()))
	switch value, _, _ := cmdArgs.GetArg("color"); value {
	case "always":
		useColor = true
	case "auto":
		useColor = term.IsTerminal(int(os.Stdout.Fd()))
	case "never":
		useColor = false
	}

	return pacmanConf, useColor, nil
}

func main() {
	text.Init(localePath)

	if os.Geteuid() == 0 {
		text.Warnln(text.T("Avoid running yay as root/sudo."))
	}

	rc, err := appMain()
	if err != nil && err.Error() != "" {
		text.EPrintln(err)
	}
	os.Exit(rc)
}

func appMain() (int, error) {
	config, err := settings.NewConfig()
	if err != nil {
		return 1, err
	}

	fl := &settings.AdditionalFlags{}
	cmdArgs := settings.NewFlagParser()
	err = settings.ParseCommandLine(cmdArgs, os.Args[1:], config, fl)
	if err != nil {
		return 1, err
	}

	pacmanConf, useColor, err := initAlpm(cmdArgs, config.PacmanConf)
	if err != nil {
		return 1, err
	}

	text.UseColor = useColor

	dbExecutor, err := ialpm.NewExecutor(pacmanConf)
	if err != nil {
		return 1, err
	}
	defer dbExecutor.Cleanup()

	runt, err := runtime.New(config, pacmanConf, dbExecutor, fl.SaveConfig, fl.Mode)
	if err != nil {
		return 1, err
	}

	if runt.SaveConfig {
		if errS := config.Save(runt.ConfigPath); errS != nil {
			text.EPrintln(err)
		}
	}

	err = handleCmd(cmdArgs, runt)
	if err != nil {
		return 1, err
	}

	return 0, nil
}

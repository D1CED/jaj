package main // import "github.com/Jguer/yay/v10"

import (
	"fmt"
	"os"

	pacmanconf "github.com/Morganamilo/go-pacmanconf"
	rpc "github.com/mikkeloscar/aur"

	"github.com/Jguer/yay/v10/pkg/db/ialpm"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/runtime"
	"github.com/Jguer/yay/v10/pkg/text"
)

func initAlpm(cmdArgs *settings.PacmanConf, pacmanConfigPath string) (*pacmanconf.Config, bool, error) {

	root := "/"
	if cmdArgs.Root != "" {
		root = cmdArgs.Root
	}

	pacmanConf, stderr, err := pacmanconf.PacmanConf("--config", pacmanConfigPath, "--root", root)
	if err != nil {
		return nil, false, fmt.Errorf("%s", stderr)
	}

	if cmdArgs.DBPath != "" {
		pacmanConf.DBPath = cmdArgs.DBPath
	}

	if cmdArgs.Arch != "" {
		pacmanConf.Architecture = cmdArgs.Arch
	}

	sconf, ok := cmdArgs.ModeConf.(*settings.SConf)
	if ok && len(sconf.Ignore) > 0 {
		pacmanConf.IgnorePkg = append(pacmanConf.IgnorePkg, sconf.Ignore...)
	}

	if ok && len(sconf.IgnoreGroup) > 0 {
		pacmanConf.IgnoreGroup = append(pacmanConf.IgnoreGroup, sconf.IgnoreGroup...)
	}

	if cmdArgs.CacheDir != "" {
		pacmanConf.CacheDir = []string{cmdArgs.CacheDir}
	}

	if cmdArgs.GPGDir != "" {
		pacmanConf.GPGDir = cmdArgs.GPGDir
	}

	outIsTerm := text.InIsTerminal()
	useColor := pacmanConf.Color && outIsTerm
	switch cmdArgs.Color {
	case settings.ColorAlways:
		useColor = true
	case settings.ColorAuto:
		useColor = outIsTerm
	case settings.ColorNever:
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

	config, err := settings.ParseCommandLine(os.Args[1:])
	if err != nil {
		return 1, err
	}
	rpc.AURURL = config.AURURL + "/rpc.php?"

	pacmanConf, useColor, err := initAlpm(config.Pacman, config.PacmanConf)
	if err != nil {
		return 1, err
	}

	text.UseColor = useColor

	dbExecutor, err := ialpm.NewExecutor(pacmanConf, false, config.Pacman.NoConfirm)
	if err != nil {
		return 1, err
	}
	defer dbExecutor.Cleanup()

	runt, err := runtime.New(config, pacmanConf, dbExecutor)
	if err != nil {
		return 1, err
	}

	if config.SaveConfig {
		if errS := config.Save(runt.Config.ConfigPath); errS != nil {
			text.EPrintln(err)
		}
	}

	err = handleCmd(runt)
	if err != nil {
		return 1, err
	}

	return 0, nil
}

package runtime

import (
	"path/filepath"
	"strings"

	"github.com/Morganamilo/go-pacmanconf"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/exe"
	"github.com/Jguer/yay/v10/pkg/vcs"
)

// vcsFileName holds the name of the vcs file.
const vcsFileName = "vcs.json"

const completionFileName = "completion.cache"

type Runtime struct {
	Mode           settings.TargetMode
	SaveConfig     bool
	CompletionPath string
	ConfigPath     string
	PacmanConf     *pacmanconf.Config
	VCSStore       *vcs.InfoStore
	CmdBuilder     *exe.CmdBuilder
	CmdRunner      exe.Runner
	DB             db.Executor
	Config         *settings.Configuration
}

func New(
	conf *settings.Configuration,
	pConf *pacmanconf.Config,
	db db.Executor, safeConfig bool,
	mode settings.TargetMode) (*Runtime, error) {

	cmdRunner := &exe.OSRunner{}
	cmdBuilder := &exe.CmdBuilder{
		GitBin:          conf.GitBin,
		GitFlags:        strings.Fields(conf.GitFlags),
		MakepkgFlags:    strings.Fields(conf.MFlags),
		MakepkgConfPath: conf.MakepkgConf,
		MakepkgBin:      conf.MakepkgBin,
	}

	vcsStore := vcs.NewInfoStore(filepath.Join(settings.GetCacheHome(), vcsFileName), cmdRunner, cmdBuilder)
	err := vcsStore.Load()

	r := &Runtime{
		Mode:           mode,
		SaveConfig:     safeConfig,
		CompletionPath: filepath.Join(settings.GetCacheHome(), completionFileName),
		ConfigPath:     settings.GetConfigPath(),

		PacmanConf: pConf,
		CmdRunner:  cmdRunner,
		CmdBuilder: cmdBuilder,
		VCSStore:   vcsStore,
		DB:         db,
		Config:     conf,
	}

	return r, err
}

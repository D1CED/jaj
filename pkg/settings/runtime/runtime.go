package runtime

import (
	"path/filepath"
	"strings"

	pacmanconf "github.com/Morganamilo/go-pacmanconf"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/settings/exe"
	"github.com/Jguer/yay/v10/pkg/vcs"
)

// vcsFileName holds the name of the vcs file.
const vcsFileName = "vcs.json"

type Runtime struct {
	VCSStore   *vcs.InfoStore
	CmdBuilder *exe.CmdBuilder
	CmdRunner  exe.Runner
	DB         db.Executor
	Pacman     *pacmanconf.Config
	Config     *settings.YayConfig
}

func New(conf *settings.YayConfig, pac *pacmanconf.Config, db db.Executor) (*Runtime, error) {

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
		VCSStore:   vcsStore,
		CmdBuilder: cmdBuilder,
		CmdRunner:  cmdRunner,
		DB:         db,
		Pacman:     pac,
		Config:     conf,
	}

	return r, err
}

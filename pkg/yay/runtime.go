package yay

import (
	"path/filepath"
	"strings"

	pacmanconf "github.com/Morganamilo/go-pacmanconf"

	"github.com/Jguer/yay/v10/pkg/db"
	"github.com/Jguer/yay/v10/pkg/exe"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/vcs"
)

// vcsFileName holds the name of the vcs file.
const vcsFileName = "vcs.json"

type Runtime struct {
	VCSStore       *vcs.InfoStore
	GitBuilder     *exe.GitBuilder
	MakepkgBuilder *exe.MakepkgBuilder
	CmdRunner      exe.Runner
	DB             db.Executor
	AUR            *query.AUR
	Pacman         *pacmanconf.Config
	Config         *settings.YayConfig
}

func New(conf *settings.YayConfig, pac *pacmanconf.Config, db db.Executor) (*Runtime, error) {

	cmdRunner := &exe.OSRunner{}
	gitBuilder := &exe.GitBuilder{
		GitBin:   conf.GitBin,
		GitFlags: strings.Fields(conf.GitFlags),
	}
	mkpkgBuilder := &exe.MakepkgBuilder{
		MakepkgFlags:    strings.Fields(conf.MFlags),
		MakepkgConfPath: conf.MakepkgConf,
		MakepkgBin:      conf.MakepkgBin,
	}

	vcsStore := vcs.NewInfoStore(filepath.Join(conf.BuildDir, vcsFileName), cmdRunner, gitBuilder)
	err := vcsStore.Load()

	r := &Runtime{
		VCSStore:       vcsStore,
		GitBuilder:     gitBuilder,
		MakepkgBuilder: mkpkgBuilder,
		CmdRunner:      cmdRunner,
		DB:             db,
		Pacman:         pac,
		Config:         conf,
		AUR:            &query.AUR{URL: conf.AURURL + "/rpc.php?"},
	}

	return r, err
}

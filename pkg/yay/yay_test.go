package yay_test

import (
	"testing"

	"github.com/Jguer/yay/v10/pkg/db/mock"
	"github.com/stretchr/testify/assert"

	"os/exec"

	"github.com/Jguer/yay/v10/pkg/exe"
	"github.com/Jguer/yay/v10/pkg/query"
	"github.com/Jguer/yay/v10/pkg/settings"
	"github.com/Jguer/yay/v10/pkg/vcs"

	pacmanconf "github.com/Morganamilo/go-pacmanconf"

	"github.com/Jguer/yay/v10/pkg/yay"
)

type mockRunner struct{ capture [][]string }

func (m *mockRunner) Capture(c *exec.Cmd, t int64) (string, string, error) {
	m.capture = append(m.capture, []string{c.Path})
	m.capture[len(m.capture)-1] = append(m.capture[len(m.capture)-1], c.Args...)
	return "", "", nil
}
func (m *mockRunner) Show(c *exec.Cmd) error {
	m.capture = append(m.capture, []string{c.Path})
	m.capture[len(m.capture)-1] = append(m.capture[len(m.capture)-1], c.Args...)
	return nil
}

func BuildRuntime() (*yay.Runtime, *mockRunner) {

	r := &yay.Runtime{
		GitBuilder:     &exe.GitBuilder{GitBin: "git-test"},
		MakepkgBuilder: &exe.MakepkgBuilder{MakepkgBin: "mkpkg-test"},
		DB:             &mock.DBMock{},
		AUR:            &query.AUR{URL: "example.com"},
		Pacman:         &pacmanconf.Config{},
		Config: &settings.YayConfig{
			Pacman:              &settings.PacmanConf{Targets: new([]string)},
			PersistentYayConfig: *settings.Defaults(),
		},
	}
	run := &mockRunner{}
	r.CmdRunner = run
	r.VCSStore = &vcs.InfoStore{
		GitBuilder: r.GitBuilder.(*exe.GitBuilder),
		Runner:     r.CmdRunner,
	}
	return r, run
}

func TestPassToPacman(t *testing.T) {
	r, run := BuildRuntime()
	r.Config.MainOperation = 'S'
	r.Config.Pacman.ModeConf = &settings.SConf{Refresh: 1}

	e := yay.PassToPacman(r.Config, r.Config.Pacman)
	r.CmdRunner.Show(e)

	assert.Equal(t, 1, len(run.capture))
}

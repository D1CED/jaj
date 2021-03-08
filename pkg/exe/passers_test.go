package exe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Jguer/yay/v10/pkg/exe"
)

func TestCmdBuilder_BuildGitCmd(t *testing.T) {
	cmdBuilder := &exe.CmdBuilder{GitBin: "git-bin", GitFlags: []string{"--git-flag"}}

	cmd := cmdBuilder.BuildGitCmd("my-directory", "--additional-argument")

	assert.ElementsMatch(t, []string{"git-bin", "--git-flag", "-C", "my-directory", "--additional-argument"}, cmd.Args)
}

func TestCmdBuilder_BuildMakepkgCmd(t *testing.T) {
	cmdBuilder := &exe.CmdBuilder{
		MakepkgBin:      "mkpkg-bin",
		MakepkgFlags:    []string{"--makepkg-flag"},
		MakepkgConfPath: "mkpkg-conf",
	}

	cmd := cmdBuilder.BuildMakepkgCmd("my-direcotry", "--additional-argument")

	assert.ElementsMatch(t, []string{"mkpkg-bin", "--makepkg-flag", "--config", "mkpkg-conf", "--additional-argument"}, cmd.Args)
}

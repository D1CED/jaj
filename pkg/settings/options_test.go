package settings

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jguer/yay/v10/pkg/settings/parser"
)

func TestParse(t *testing.T) {

	tt := []struct {
		args      string
		want      *YayConfig
		err       bool
		errPhase2 bool
	}{{
		args: "-Syu",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Once,
				Refresh:    Once,
			}},
		},
	}, {
		args: "-Syyuu",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, {
		args: "-Suuyy",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, {
		args: "-S -uuyy",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, {
		args: "-Suyyu",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, {
		args: "--sync -u --refresh -y --sysupgrade",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, {
		args: "-S some-pkg",
		want: &YayConfig{
			MainOperation: 'S',
			Targets:       []string{"some-pkg"},
			Pacman: &PacmanConf{
				ModeConf: &SConf{},
				Targets:  &[]string{"some-pkg"},
			},
		},
	}, {
		args: "some-pkg other-pkg",
		want: &YayConfig{
			MainOperation: 'Y',
			ModeConf:      &YConf{},
			Targets:       []string{"some-pkg", "other-pkg"},
		},
	}, {
		args: "-SY",
		err:  true,
	}, {
		args: "--noansweredit --answeredit=None",
		want: &YayConfig{
			MainOperation:       'Y',
			ModeConf:            &YConf{},
			PersistentYayConfig: PersistentYayConfig{AnswerEdit: "None"},
		},
	}, {
		args: "--answeredit=None --noansweredit",
		want: &YayConfig{
			MainOperation: 'Y',
			ModeConf:      &YConf{},
		},
	}, {
		args: "--answeredit=All some-pkg -P --fish --aur -r/test/ --color always",
		want: &YayConfig{
			MainOperation:       'P',
			ModeConf:            &PConf{Fish: true},
			Targets:             []string{"some-pkg"},
			PersistentYayConfig: PersistentYayConfig{AnswerEdit: "All"},
			Mode:                ModeAUR,
			Pacman:              &PacmanConf{Root: "/test/", Color: ColorAlways, Targets: &[]string{"some-pkg"}},
		},
	}, {
		args:      "--color always --aur --fish -Pr/test/ some-pkg --answeredit=All",
		errPhase2: true,
	}, {
		args: "--color always -r/test/ --aur --show --fish some-pkg --answeredit=All",
		want: &YayConfig{
			MainOperation:       'P',
			ModeConf:            &PConf{Fish: true},
			Targets:             []string{"some-pkg"},
			PersistentYayConfig: PersistentYayConfig{AnswerEdit: "All"},
			Mode:                ModeAUR,
			Pacman:              &PacmanConf{Root: "/test/", Color: ColorAlways, Targets: &[]string{"some-pkg"}},
		},
	}}

	compare := func(t *testing.T, expect *YayConfig, got *YayConfig, targets []string) {
		if len(targets) != 0 {
			got.Targets = targets
			got.Pacman.Targets = &got.Targets
		}

		if expect.Pacman == nil {
			expect.Pacman = new(PacmanConf)
			expect.Pacman.Targets = &expect.Targets
		}
		if expect.Pacman.Targets == nil {
			expect.Pacman.Targets = new([]string)
		}

		wantP := expect.Pacman
		expect.Pacman = nil
		yayP := got.Pacman
		got.Pacman = nil
		assert.Equal(t, expect, got)
		assert.Equal(t, wantP, yayP)
	}

	for i, test := range tt {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			yay := &YayConfig{Pacman: new(PacmanConf)}
			yay.Pacman.Targets = &yay.Targets

			a, err := parser.Parse(mappingFunc(), strings.Split(test.args, " "), nil)
			if test.err {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			a.Iterate(handleConfig(yay, &err))
			if test.errPhase2 {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if yay.MainOperation == 0 {
				yay.MainOperation = OpYay
				yay.ModeConf = &YConf{}
			}

			compare(t, test.want, yay, a.Targets())
		})
	}
}

package settings

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseCommandLine(t *testing.T) {

	tt := []struct {
		args string
		want *YayConfig
		err  bool
	}{0: {
		args: "-Syu",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Once,
				Refresh:    Once,
			}},
		},
	}, 1: {
		args: "-Syyuu",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, 2: {
		args: "-Suuyy",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, 3: {
		args: "-S -uuyy",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, 4: {
		args: "-Suyyu",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, 5: {
		args: "--sync -u --refresh -y --sysupgrade",
		want: &YayConfig{
			MainOperation: 'S',
			Pacman: &PacmanConf{ModeConf: &SConf{
				SysUpgrade: Twice,
				Refresh:    Twice,
			}},
		},
	}, 6: {
		args: "-S some-pkg",
		want: &YayConfig{
			MainOperation: 'S',
			Targets:       []string{"some-pkg"},
			Pacman: &PacmanConf{
				ModeConf: &SConf{},
				Targets:  &[]string{"some-pkg"},
			},
		},
	}, 7: {
		args: "some-pkg other-pkg",
		want: &YayConfig{
			MainOperation: 'Y',
			ModeConf:      &YConf{},
			Targets:       []string{"some-pkg", "other-pkg"},
		},
	}, 8: {
		args: "-SY",
		err:  true,
	}, 9: {
		args: "--noansweredit --answeredit=None",
		want: &YayConfig{
			MainOperation:       'Y',
			ModeConf:            &YConf{},
			PersistentYayConfig: PersistentYayConfig{AnswerEdit: "None"},
		},
	}, 10: {
		args: "--answeredit=None --noansweredit",
		want: &YayConfig{
			MainOperation: 'Y',
			ModeConf:      &YConf{},
		},
	}, 11: {
		args: "--answeredit=All some-pkg -P --fish --aur -r/test/ --color always",
		want: &YayConfig{
			MainOperation:       'P',
			ModeConf:            &PConf{Fish: true},
			Targets:             []string{"some-pkg"},
			PersistentYayConfig: PersistentYayConfig{AnswerEdit: "All"},
			Mode:                ModeAUR,
			Pacman:              &PacmanConf{Root: "/test/", Color: ColorAlways, Targets: &[]string{"some-pkg"}},
		},
	}, 12: {
		args: "--color always --aur --fish -Pr/test/ some-pkg --answeredit=All",
		err:  true,
	}, 13: {
		args: "--color always -r/test/ --aur --show --fish some-pkg --answeredit=All",
		want: &YayConfig{
			MainOperation:       'P',
			ModeConf:            &PConf{Fish: true},
			Targets:             []string{"some-pkg"},
			PersistentYayConfig: PersistentYayConfig{AnswerEdit: "All"},
			Mode:                ModeAUR,
			Pacman:              &PacmanConf{Root: "/test/", Color: ColorAlways, Targets: &[]string{"some-pkg"}},
		},
	}, 14: {
		args: "--color always --aur --fish -Pr/test/ some-pkg --answeredit=All --unknown-option=5",
		err:  true,
	}, 15: {
		args: "-Ss --repo --topdown racket ide",
		want: &YayConfig{
			MainOperation:       'S',
			Targets:             []string{"racket", "ide"},
			PersistentYayConfig: PersistentYayConfig{SortMode: TopDown},
			Mode:                ModeRepo,
			Pacman:              &PacmanConf{Targets: &[]string{"racket", "ide"}, ModeConf: &SConf{Search: true}},
		},
	}}

	compare := func(t *testing.T, expect *YayConfig, got *YayConfig, targets []string) {
		if len(targets) == 0 {
			got.Targets = nil
		}

		if expect.Pacman == nil {
			expect.Pacman = new(PacmanConf)
			if expect.Targets != nil {
				expect.Pacman.Targets = &expect.Targets
			}
		}
		if len(*got.Pacman.Targets) == 0 {
			got.Pacman.Targets = nil
		}

		wantP := expect.Pacman
		expect.Pacman = nil
		yayP := got.Pacman
		got.Pacman = nil
		assert.Equal(t, expect, got)
		assert.Equal(t, wantP, yayP)
	}

	for i, test := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			yay := &YayConfig{Pacman: new(PacmanConf)}
			yay.Pacman.Targets = &yay.Targets

			err := parseCommandLine(strings.Split(test.args, " "), yay, nil)
			if test.err {
				t.Log(err)
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			compare(t, test.want, yay, yay.Targets)
		})
	}
}

func TestDashInput(t *testing.T) {

	t.Run("file_attached", func(t *testing.T) {
		var r io.Reader = bytes.NewBufferString("hello\nworld\n")
		yay := &YayConfig{Pacman: new(PacmanConf)}
		yay.Pacman.Targets = &yay.Targets

		err := parseCommandLine([]string{"-"}, yay, &r)
		require.NoError(t, err)

		assert.Equal(t, []string{"hello", "world"}, yay.Targets)

		if term.IsTerminal(int(os.Stdin.Fd())) {
			_, ok := r.(*os.File)
			assert.True(t, ok)
		}
	})

	t.Run("term_attached", func(t *testing.T) {
		yay := &YayConfig{Pacman: new(PacmanConf)}
		yay.Pacman.Targets = &yay.Targets

		r := io.Reader(os.Stdin)
		err := parseCommandLine([]string{"-"}, yay, &r)
		require.NoError(t, err)

		assert.Equal(t, []string{}, yay.Targets)
	})
}

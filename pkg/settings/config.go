package settings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jguer/yay/v10/pkg/text"
)

// Describes Sorting method for numberdisplay
const (
	BottomUp int = iota
	TopDown
)

type OpMode byte

const (
	OpDatabase OpMode = 'D'
	OpFiles    OpMode = 'F'
	OpQuery    OpMode = 'Q'
	OpRemove   OpMode = 'R'
	OpSync     OpMode = 'S'
	OpDepTest  OpMode = 'T'
	OpUpgrade  OpMode = 'U'

	OpVersion OpMode = 'V'
	OpHelp    OpMode = 'h'

	OpYay         OpMode = 'Y'
	OpShow        OpMode = 'P'
	OpGetPkgbuild OpMode = 'G'
)

type TargetMode int

const (
	ModeAny TargetMode = iota
	ModeAUR
	ModeRepo
)

type SearchMode int

// Verbosity settings for search
const (
	NumberMenu SearchMode = iota
	Detailed
	Minimal
)
const completionFileName = "completion.cache"

// HideMenus indicates if pacman's provider menus must be hidden
var hideMenus = false

// NoConfirm indicates if user input should be skipped
var userNoConfirm = false // use PacmanConf.NoConfirm

type YayConfig struct {
	MainOperation OpMode
	ModeConf      interface{ mark() } // *(P|Y|G)Conf

	SaveConfig     bool
	Mode           TargetMode
	SearchMode     SearchMode
	CompletionPath string
	ConfigPath     string

	PersistentYayConfig
	Targets []string
	Pacman  *PacmanConf
}

type PConf struct {
	Complete      Trilean
	Fish          bool
	DefaultConfig bool
	CurrentConfig bool
	LocalStats    bool
	News          bool
	Quiet         bool

	Upgrades       bool
	NumberUpgrades bool
}

type YConf struct {
	GenDevDB bool
	Clean    Trilean
}

type GConf struct {
	Force bool
}

func (*PConf) mark() {}
func (*YConf) mark() {}
func (*GConf) mark() {}

// Configuration stores yay's config.
type PersistentYayConfig struct {
	AURURL             string `json:"aururl"`
	BuildDir           string `json:"buildDir"`
	ABSDir             string `json:"absdir"`
	Editor             string `json:"editor"`
	EditorFlags        string `json:"editorflags"`
	MakepkgBin         string `json:"makepkgbin"`
	MakepkgConf        string `json:"makepkgconf"`
	PacmanBin          string `json:"pacmanbin"`
	PacmanConf         string `json:"pacmanconf"`
	ReDownload         string `json:"redownload"`
	ReBuild            string `json:"rebuild"`
	AnswerClean        string `json:"answerclean"`
	AnswerDiff         string `json:"answerdiff"`
	AnswerEdit         string `json:"answeredit"`
	AnswerUpgrade      string `json:"answerupgrade"`
	GitBin             string `json:"gitbin"`
	GpgBin             string `json:"gpgbin"`
	GpgFlags           string `json:"gpgflags"`
	MFlags             string `json:"mflags"`
	SortBy             string `json:"sortby"`
	SearchBy           string `json:"searchby"`
	GitFlags           string `json:"gitflags"`
	RemoveMake         string `json:"removemake"`
	SudoBin            string `json:"sudobin"`
	SudoFlags          string `json:"sudoflags"`
	RequestSplitN      int    `json:"requestsplitn"`
	SortMode           int    `json:"sortmode"`
	CompletionInterval int    `json:"completionrefreshtime"`
	SudoLoop           bool   `json:"sudoloop"`
	TimeUpdate         bool   `json:"timeupdate"`
	Devel              bool   `json:"devel"`
	CleanAfter         bool   `json:"cleanAfter"`
	Provides           bool   `json:"provides"`
	PGPFetch           bool   `json:"pgpfetch"`
	UpgradeMenu        bool   `json:"upgrademenu"`
	CleanMenu          bool   `json:"cleanmenu"`
	DiffMenu           bool   `json:"diffmenu"`
	EditMenu           bool   `json:"editmenu"`
	CombinedUpgrade    bool   `json:"combinedupgrade"`
	UseAsk             bool   `json:"useask"`
	BatchInstall       bool   `json:"batchinstall"`

	Tar string `json:"tar"`
}

var defaultYayConfig = PersistentYayConfig{
	AURURL:             "https://aur.archlinux.org",
	BuildDir:           os.ExpandEnv("$HOME/.cache/yay"),
	ABSDir:             os.ExpandEnv("$HOME/.cache/yay/abs"),
	CleanAfter:         false,
	Editor:             "",
	EditorFlags:        "",
	Devel:              false,
	MakepkgBin:         "makepkg",
	MakepkgConf:        "",
	PacmanBin:          "pacman",
	PGPFetch:           true,
	PacmanConf:         "/etc/pacman.conf",
	GpgFlags:           "",
	MFlags:             "",
	GitFlags:           "",
	SortMode:           BottomUp,
	CompletionInterval: 7,
	SortBy:             "votes",
	SearchBy:           "name-desc",
	SudoLoop:           false,
	GitBin:             "git",
	GpgBin:             "gpg",
	SudoBin:            "sudo",
	SudoFlags:          "",
	TimeUpdate:         false,
	RequestSplitN:      150,
	ReDownload:         "no",
	ReBuild:            "no",
	BatchInstall:       false,
	AnswerClean:        "",
	AnswerDiff:         "",
	AnswerEdit:         "",
	AnswerUpgrade:      "",
	RemoveMake:         "ask",
	Provides:           true,
	UpgradeMenu:        true,
	CleanMenu:          true,
	DiffMenu:           true,
	EditMenu:           false,
	UseAsk:             false,
	CombinedUpgrade:    false,
}

func Defaults() *PersistentYayConfig {
	dc := new(PersistentYayConfig)
	*dc = defaultYayConfig
	return dc
}

func newConfig() (*PersistentYayConfig, error) {
	new := Defaults()

	cacheHome := getCacheHome()
	new.BuildDir = cacheHome

	configPath := getConfigPath()
	new.load(configPath)

	if aurdest := os.Getenv("AURDEST"); aurdest != "" {
		new.BuildDir = aurdest
	}

	new.expandEnv()

	err := initDir(new.BuildDir)

	return new, err
}

// SaveConfig writes yay config to file.
func (c *PersistentYayConfig) Save(configPath string) error {
	marshalledinfo, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	// https://github.com/Jguer/yay/issues/1325
	marshalledinfo = append(marshalledinfo, '\n')
	// https://github.com/Jguer/yay/issues/1399
	if _, err = os.Stat(filepath.Dir(configPath)); os.IsNotExist(err) && err != nil {
		if mkErr := os.MkdirAll(filepath.Dir(configPath), 0o755); mkErr != nil {
			return mkErr
		}
	}
	in, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer in.Close()
	if _, err = in.Write(marshalledinfo); err != nil {
		return err
	}
	return in.Sync()
}

func (c *PersistentYayConfig) AsJSONString() string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "\t")
	if err := enc.Encode(c); err != nil {
		text.EPrintln(err)
	}
	return buf.String()
}

func (c *PersistentYayConfig) expandEnv() {
	c.AURURL = os.ExpandEnv(c.AURURL)
	c.ABSDir = os.ExpandEnv(c.ABSDir)
	c.BuildDir = os.ExpandEnv(c.BuildDir)
	c.Editor = os.ExpandEnv(c.Editor)
	c.EditorFlags = os.ExpandEnv(c.EditorFlags)
	c.MakepkgBin = os.ExpandEnv(c.MakepkgBin)
	c.MakepkgConf = os.ExpandEnv(c.MakepkgConf)
	c.PacmanBin = os.ExpandEnv(c.PacmanBin)
	c.PacmanConf = os.ExpandEnv(c.PacmanConf)
	c.GpgFlags = os.ExpandEnv(c.GpgFlags)
	c.MFlags = os.ExpandEnv(c.MFlags)
	c.GitFlags = os.ExpandEnv(c.GitFlags)
	c.SortBy = os.ExpandEnv(c.SortBy)
	c.SearchBy = os.ExpandEnv(c.SearchBy)
	c.GitBin = os.ExpandEnv(c.GitBin)
	c.GpgBin = os.ExpandEnv(c.GpgBin)
	c.SudoBin = os.ExpandEnv(c.SudoBin)
	c.SudoFlags = os.ExpandEnv(c.SudoFlags)
	c.ReDownload = os.ExpandEnv(c.ReDownload)
	c.ReBuild = os.ExpandEnv(c.ReBuild)
	c.AnswerClean = os.ExpandEnv(c.AnswerClean)
	c.AnswerDiff = os.ExpandEnv(c.AnswerDiff)
	c.AnswerEdit = os.ExpandEnv(c.AnswerEdit)
	c.AnswerUpgrade = os.ExpandEnv(c.AnswerUpgrade)
	c.RemoveMake = os.ExpandEnv(c.RemoveMake)
}

func (c *PersistentYayConfig) load(configPath string) error {
	cfile, err := os.Open(configPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf(text.Tf("failed to open config file '%s': %s", configPath, err))
	}
	defer cfile.Close()

	err = json.NewDecoder(cfile).Decode(c)
	if err != nil {
		return fmt.Errorf(text.Tf("failed to read config file '%s': %s", configPath, err))
	}
	return nil
}

package settings

import (
	"fmt"
	"strings"
)

type Trilean uint8

const (
	Unset Trilean = 0
	Once  Trilean = 1
	Twice Trilean = 2
)

func (t Trilean) repeat(r rune) string {
	switch t {
	case Unset:
		return ""
	case Once:
		return string(r)
	default:
		fallthrough
	case Twice:
		return string(r) + string(r)
	}
}

type Transaction struct {
	NoDeps          Trilean
	AssumeInstalled []struct {
		Package string
		Version string
	}
	DBOnly        bool
	NoProgressbar bool
	NoScriptlet   bool
	Print         bool
	PrintFormat   string
}

type Upgrade struct {
	AsDeps      bool
	AsExplicit  bool
	Ignore      []string
	IgnoreGroup []string
	Needed      bool
	Overwrite   string
}

type ColorMode int

const (
	ColorAuto ColorMode = iota
	ColorAlways
	ColorNever
)

type PacmanConf struct {
	ModeConf interface{ formatAsArgs(s []string) []string } // *XConf struct
	Targets  *[]string

	DBPath                 string
	Root                   string
	Verbose                bool
	Arch                   string
	CacheDir               string
	Color                  ColorMode
	Config                 string
	Debug                  bool
	GPGDir                 string
	HookDir                string
	LogFile                string
	NoConfirm              bool
	DisableDownloadTimeout bool
	SysRoot                string

	Ask int
}

type DConf struct {
	AsDeps     string
	AsExplicit string
	Check      Trilean
	Quiet      bool
}
type QConf struct {
	Changelog  bool
	Deps       bool
	Explicit   bool
	Groups     bool
	Info       Trilean
	Check      Trilean
	List       bool
	Foreign    bool
	Native     bool
	Owns       string
	File       bool
	Quiet      bool
	Search     bool
	Unrequired Trilean
	Upgrades   Trilean // see main.handleQuery
}
type RConf struct {
	Transaction

	Cascade   bool
	NoSave    bool
	Recursive Trilean
	Unneeded  bool
}
type SConf struct {
	Transaction
	Upgrade

	Clean        Trilean
	Groups       Trilean
	Info         Trilean
	List         bool
	Quiet        bool
	Search       bool
	SysUpgrade   Trilean
	DownloadOnly bool
	Refresh      Trilean
}
type UConf struct {
	Transaction
	Upgrade
}
type FConf struct {
	Refresh         Trilean
	List            bool
	Regex           bool
	Quiet           bool
	MachineReadable bool
}
type TConf struct{}

// -------------------------------------------------------

func mainOp(p *PacmanConf) rune {
	var op rune
	switch p.ModeConf.(type) {
	case *DConf:
		op = rune(OpDatabase)
	case *FConf:
		op = rune(OpFiles)
	case *QConf:
		op = rune(OpQuery)
	case *RConf:
		op = rune(OpRemove)
	case *SConf:
		op = rune(OpSync)
	case *TConf:
		op = rune(OpDepTest)
	case *UConf:
		op = rune(OpUpgrade)
	}
	return op
}

func (p *PacmanConf) String() string {
	return fmt.Sprintf("Op:%c Options:%#v", mainOp(p), p)
}

func (p *PacmanConf) DeepCopy() *PacmanConf {
	var q = new(PacmanConf)
	*q = *p
	q.Targets = new([]string)
	*q.Targets = make([]string, len(*p.Targets))
	copy(*q.Targets, *p.Targets)

	switch t := p.ModeConf.(type) {
	case *DConf:
		q.ModeConf = &(*t)
	case *FConf:
		q.ModeConf = &(*t)
	case *QConf:
		q.ModeConf = &(*t)
	case *RConf:
		q.ModeConf = &(*t)
	case *SConf:
		q.ModeConf = &(*t)
	case *UConf:
		q.ModeConf = &(*t)
	case *TConf: // are empty anyways
	}

	return q
}

func (p *PacmanConf) FormatAsArgs(help, version bool) []string {
	if version && !help {
		return []string{"-V"}
	}
	s := make([]string, 0, 32)
	s = formatGlobalArgs(p, s)

	type fmtArgs interface{ formatAsArgs(s []string) []string }

	if m, ok := p.ModeConf.(fmtArgs); ok {
		s = m.formatAsArgs(s)
	}
	if help {
		s = append(s, "-h")
	}
	return s
}

func formatGlobalArgs(p *PacmanConf, s []string) []string {

	if p.DBPath != "" {
		s = append(s, "--dbpath="+p.DBPath)
	}
	if p.Root != "" {
		s = append(s, "--root="+p.Root)
	}
	if p.Verbose {
		s = append(s, "-v")
	}
	if p.Arch != "" {
		s = append(s, "--arch="+p.Arch)
	}
	if p.CacheDir != "" {
		s = append(s, "--cachedir="+p.CacheDir)
	}
	if p.Color != ColorAuto {
		switch p.Color {
		case ColorAlways:
			s = append(s, "--color=always")
		case ColorNever:
			s = append(s, "--color=never")
		}
	}
	if p.Config != "" {
		s = append(s, "--config="+p.Config)
	}
	if p.Debug {
		s = append(s, "--debug")
	}
	if p.GPGDir != "" {
		s = append(s, "--gpgdir="+p.GPGDir)
	}
	if p.HookDir != "" {
		s = append(s, "--hookdir="+p.HookDir)
	}
	if p.LogFile != "" {
		s = append(s, "--logfile="+p.LogFile)
	}
	if p.NoConfirm {
		s = append(s, "--noconfirm")
	}
	if p.DisableDownloadTimeout {
		s = append(s, "--disable-download-timeout")
	}
	if p.SysRoot != "" {
		s = append(s, "--sysroot="+p.SysRoot)
	}

	if op := mainOp(p); op != 0 {
		s = append(s, "-"+string(op))
	}

	return s
}

func (D *DConf) formatAsArgs(s []string) []string {
	if D.AsDeps != "" {
		s = append(s, "--asdeps="+D.AsDeps)
	}
	if D.AsExplicit != "" {
		s = append(s, "--asexplicit="+D.AsDeps)
	}
	if D.Check != 0 {
		s = append(s, "-"+D.Check.repeat('k'))
	}
	if D.Quiet {
		s = append(s, "-q")
	}
	return s
}
func (Q *QConf) formatAsArgs(s []string) []string {
	if Q.Changelog {
		s = append(s, "-c")
	}
	if Q.Deps {
		s = append(s, "-d")
	}
	if Q.Explicit {
		s = append(s, "-e")
	}
	if Q.Groups {
		s = append(s, "-g")
	}
	if Q.Info != 0 {
		s = append(s, "-"+Q.Info.repeat('i'))
	}
	if Q.Check != 0 {
		s = append(s, "-"+Q.Check.repeat('k'))
	}
	if Q.List {
		s = append(s, "-l")
	}
	if Q.Foreign {
		s = append(s, "-m")
	}
	if Q.Native {
		s = append(s, "-n")
	}
	if Q.Owns != "" {
		s = append(s, "--owns="+Q.Owns)
	}
	if Q.File {
		s = append(s, "-p")
	}
	if Q.Quiet {
		s = append(s, "-q")
	}
	if Q.Search {
		s = append(s, "-s")
	}
	if Q.Unrequired != 0 {
		s = append(s, "-"+Q.Unrequired.repeat('t'))
	}
	if Q.Upgrades != 0 {
		s = append(s, "-u") // no double u for pacman, see above
	}
	return s
}
func (R *RConf) formatAsArgs(s []string) []string {
	if R.NoDeps != 0 {
		s = append(s, "-"+R.NoDeps.repeat('d'))
	}
	for _, kv := range R.AssumeInstalled {
		s = append(s, "--assume-installed="+kv.Package+"="+kv.Version)
	}
	if R.DBOnly {
		s = append(s, "--dbonly")
	}
	if R.NoProgressbar {
		s = append(s, "--noprogressbar")
	}
	if R.NoScriptlet {
		s = append(s, "--noscriptlet")
	}
	if R.Print {
		s = append(s, "-p")
	}
	if R.PrintFormat != "" {
		s = append(s, "--print-format="+R.PrintFormat)
	}

	if R.Cascade {
		s = append(s, "-c")
	}
	if R.NoSave {
		s = append(s, "-n")
	}
	if R.Recursive != 0 {
		s = append(s, "-"+R.Recursive.repeat('s'))
	}
	if R.Unneeded {
		s = append(s, "-u")
	}
	return s
}
func (S *SConf) formatAsArgs(s []string) []string {
	if S.NoDeps != 0 {
		s = append(s, "-"+S.NoDeps.repeat('d'))
	}
	for _, kv := range S.AssumeInstalled {
		s = append(s, "--assume-installed="+kv.Package+"="+kv.Version)
	}
	if S.DBOnly {
		s = append(s, "--dbonly")
	}
	if S.NoProgressbar {
		s = append(s, "--noprogressbar")
	}
	if S.NoScriptlet {
		s = append(s, "--noscriptlet")
	}
	if S.Print {
		s = append(s, "-p")
	}
	if S.PrintFormat != "" {
		s = append(s, "--print-format="+S.PrintFormat)
	}

	if S.AsDeps {
		s = append(s, "--asdeps")
	}
	if S.AsExplicit {
		s = append(s, "--asexplicit")
	}
	if len(S.Ignore) != 0 {
		s = append(s, "--ignore="+strings.Join(S.Ignore, ","))
	}
	if len(S.IgnoreGroup) != 0 {
		s = append(s, "--ignoregroup="+strings.Join(S.IgnoreGroup, ","))
	}
	if S.Needed {
		s = append(s, "--needed")
	}
	if S.Overwrite != "" {
		s = append(s, "--overwrite="+S.Overwrite)
	}

	if S.Clean != 0 {
		s = append(s, "-"+S.Clean.repeat('c'))
	}
	if S.Groups != 0 {
		s = append(s, "-"+S.Groups.repeat('g'))
	}
	if S.Info != 0 {
		s = append(s, "-"+S.Info.repeat('i'))
	}
	if S.List {
		s = append(s, "-l")
	}
	if S.Quiet {
		s = append(s, "-q")
	}
	if S.Search {
		s = append(s, "-s")
	}
	if S.SysUpgrade != 0 {
		s = append(s, "-"+S.SysUpgrade.repeat('u'))
	}
	if S.DownloadOnly {
		s = append(s, "-w")
	}
	if S.Refresh != 0 {
		s = append(s, "-"+S.SysUpgrade.repeat('y'))
	}
	return s
}
func (U *UConf) formatAsArgs(s []string) []string {
	if U.NoDeps != 0 {
		s = append(s, "-"+U.NoDeps.repeat('d'))
	}
	for _, kv := range U.AssumeInstalled {
		s = append(s, "--assume-installed="+kv.Package+"="+kv.Version)
	}
	if U.DBOnly {
		s = append(s, "--dbonly")
	}
	if U.NoProgressbar {
		s = append(s, "--noprogressbar")
	}
	if U.NoScriptlet {
		s = append(s, "--noscriptlet")
	}
	if U.Print {
		s = append(s, "-p")
	}
	if U.PrintFormat != "" {
		s = append(s, "--print-format="+U.PrintFormat)
	}

	if U.AsDeps {
		s = append(s, "--asdeps")
	}
	if U.AsExplicit {
		s = append(s, "--asexplicit")
	}
	if len(U.Ignore) != 0 {
		s = append(s, "--ignore="+strings.Join(U.Ignore, ","))
	}
	if len(U.IgnoreGroup) != 0 {
		s = append(s, "--ignoregroup="+strings.Join(U.IgnoreGroup, ","))
	}
	if U.Needed {
		s = append(s, "--needed")
	}
	if U.Overwrite != "" {
		s = append(s, "--overwrite="+U.Overwrite)
	}
	return s
}
func (F *FConf) formatAsArgs(s []string) []string {
	if F.Refresh != 0 {
		s = append(s, "-"+F.Refresh.repeat('y'))
	}
	if F.List {
		s = append(s, "-l")
	}
	if F.Regex {
		s = append(s, "-x")
	}
	if F.Quiet {
		s = append(s, "-q")
	}
	if F.MachineReadable {
		s = append(s, "--machinereadable")
	}
	return s
}
func (*TConf) formatAsArgs(s []string) []string { return s }

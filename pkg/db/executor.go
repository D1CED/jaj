package db

import (
	"time"

	alpm "github.com/Jguer/go-alpm/v2"
)

type IPackage = alpm.IPackage
type Depend = alpm.Depend

const PkgReasonExplicit = alpm.PkgReasonExplicit

func VerCmp(a, b string) int {
	return alpm.VerCmp(a, b)
}

type Upgrade struct {
	Name          string
	Repository    string
	LocalVersion  string
	RemoteVersion string
}

type Executor interface {
	AlpmArch() (string, error)
	BiggestPackages() []IPackage
	Cleanup()
	IsCorrectVersionInstalled(string, string) bool
	LastBuildTime() time.Time
	LocalPackage(string) IPackage
	LocalPackages() []IPackage
	LocalSatisfierExists(string) bool
	PackageConflicts(IPackage) []Depend
	PackageDepends(IPackage) []Depend
	SatisfierFromDB(string, string) IPackage
	PackageGroups(IPackage) []string
	PackageOptionalDepends(IPackage) []Depend
	PackageProvides(IPackage) []Depend
	PackagesFromGroup(string) []IPackage
	RefreshHandle() error
	RepoUpgrades(bool) ([]Upgrade, error)
	SyncPackage(string) IPackage
	SyncPackages(...string) []IPackage
	SyncSatisfier(string) IPackage
	SyncSatisfierExists(string) bool

	NoConfirm() bool
	SetNoConfirm(bool)
	HideMenus() bool
	SetHideMenus(bool)
}

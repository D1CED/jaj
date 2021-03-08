package mock

import (
	"time"

	"github.com/Jguer/yay/v10/pkg/db"
)

type DBMock struct{}

var _ db.Executor = &DBMock{}

func (m *DBMock) AlpmArch() (string, error) { return "mock", nil }

func (m *DBMock) NoConfirm() bool   { return true }
func (m *DBMock) SetNoConfirm(bool) {}
func (m *DBMock) HideMenus() bool   { return true }
func (m *DBMock) SetHideMenus(bool) {}

func (m *DBMock) Cleanup() {}

func (m *DBMock) BiggestPackages() []db.IPackage                 { return nil }
func (m *DBMock) IsCorrectVersionInstalled(string, string) bool  { return false }
func (m *DBMock) LastBuildTime() time.Time                       { return time.Time{} }
func (m *DBMock) LocalPackage(string) db.IPackage                { return nil }
func (m *DBMock) LocalPackages() []db.IPackage                   { return nil }
func (m *DBMock) LocalSatisfierExists(string) bool               { return false }
func (m *DBMock) PackageConflicts(db.IPackage) []db.Depend       { return nil }
func (m *DBMock) PackageDepends(db.IPackage) []db.Depend         { return nil }
func (m *DBMock) SatisfierFromDB(string, string) db.IPackage     { return nil }
func (m *DBMock) PackageGroups(db.IPackage) []string             { return nil }
func (m *DBMock) PackageOptionalDepends(db.IPackage) []db.Depend { return nil }
func (m *DBMock) PackageProvides(db.IPackage) []db.Depend        { return nil }
func (m *DBMock) PackagesFromGroup(string) []db.IPackage         { return nil }
func (m *DBMock) RefreshHandle() error                           { return nil }
func (m *DBMock) RepoUpgrades(bool) ([]db.Upgrade, error)        { return nil, nil }
func (m *DBMock) SyncPackage(string) db.IPackage                 { return nil }
func (m *DBMock) SyncPackages(...string) []db.IPackage           { return nil }
func (m *DBMock) SyncSatisfier(string) db.IPackage               { return nil }
func (m *DBMock) SyncSatisfierExists(string) bool                { return false }

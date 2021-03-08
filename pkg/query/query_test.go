package query_test

import (
	"testing"

	dbmock "github.com/Jguer/yay/v10/pkg/db/mock"
	"github.com/Jguer/yay/v10/pkg/settings"

	"github.com/stretchr/testify/assert"

	"github.com/Jguer/yay/v10/pkg/query"
)

func TestGetRemotePackages_Empty(t *testing.T) {
	mock := &dbmock.DBMock{}

	locPkgs, names := query.GetRemotePackages(mock)

	assert.Empty(t, locPkgs)
	assert.Empty(t, names)
}

func TestGetPackageNamesBySource_Empty(t *testing.T) {
	mock := &dbmock.DBMock{}

	locPkgs, rmtPkgs, err := query.GetPackageNamesBySource(mock)

	assert.Empty(t, locPkgs)
	assert.Empty(t, rmtPkgs)
	assert.NoError(t, err)
}

func TestHangingPackages(t *testing.T) {
	mock := &dbmock.DBMock{}

	handNames := query.HangingPackages(false, mock)
	assert.Empty(t, handNames)
}

func TestRemoveInvalidTargets(t *testing.T) {
	ss := query.RemoveInvalidTargets([]string{}, settings.ModeAUR)

	assert.Empty(t, ss)
}

func TestStatistics(t *testing.T) {
	mock := &dbmock.DBMock{}

	res := query.Statistics(mock)

	assert.Equal(t, 0, res.Totaln)
}

type aurMock struct{}

func (a aurMock) Info([]string) ([]query.Pkg, error) { return nil, nil }

func TestAURInfo(t *testing.T) {

	_, _ = query.AURInfo(aurMock{}, []string{}, nil, 0)
}

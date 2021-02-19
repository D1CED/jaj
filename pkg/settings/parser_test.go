package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeArguments(t *testing.T) {
	args := NewFlagParser()
	assert.NotNil(t, args)
	assert.Equal(t, "", args.Op)
	assert.Empty(t, args.Options)
	assert.Empty(t, args.Targets)
}

func TestArguments_DelArg(t *testing.T) {
	args := NewFlagParser()
	args.AddArg("arch") // , "arg")
	args.AddArg("ask")  // , "arg")
	args.DelArg("arch", "ask")
	assert.Empty(t, args.Options)
}

func Test_isArg(t *testing.T) {
	assert.NotContains(t, isArg, "zorg")
	assert.Contains(t, isArg, "dbpath")
}

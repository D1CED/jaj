package multierror_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	merr "github.com/Jguer/yay/v10/pkg/multierror"
)

func TestMultiError(t *testing.T) {

	e := &merr.MultiError{}

	e.Add(nil)
	assert.NoError(t, e.Return())

	e.Add(fmt.Errorf("Oh, no!"))
	assert.Error(t, e.Return())

	assert.Equal(t, "Oh, no!", e.Error())
}

//go:build apache

package processmng

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApacheReload(t *testing.T) {
	apacheProcessManager, err := GetApacheProcessManager()
	assert.Nil(t, err)

	err = apacheProcessManager.Reload()
	assert.Nil(t, err)
}

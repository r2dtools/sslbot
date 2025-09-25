//go:build apache

package processmng

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindApacheProcessByName(t *testing.T) {
	apacheProcess, err := findProcessByName([]string{"apache2"})
	assert.Nilf(t, err, "find apache process failed: %v", err)
	assert.NotNilf(t, apacheProcess, "apache process is nil")

	name, err := apacheProcess.Name()
	assert.Nil(t, err)
	assert.Equal(t, "apache2", name)

	parent, err := apacheProcess.Parent()
	assert.Nil(t, err)
	assert.Equal(t, int32(1), parent.Pid)
}

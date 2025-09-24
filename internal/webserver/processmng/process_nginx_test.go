//go:build nginx

package processmng

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindProcessByName(t *testing.T) {
	nginxProcess, err := findProcessByName([]string{"nginx"})
	assert.Nilf(t, err, "find nginx process failed: %v", err)
	assert.NotNilf(t, nginxProcess, "nginx process is nil")

	name, err := nginxProcess.Name()
	assert.Nil(t, err)
	assert.Equal(t, "nginx", name)

	parent, err := nginxProcess.Parent()
	assert.Nil(t, err)
	assert.Equal(t, int32(1), parent.Pid)
}

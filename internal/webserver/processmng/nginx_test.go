package processmng

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNginxReload(t *testing.T) {
	nginxProcessManager, err := GetNginxProcessManager()
	assert.Nil(t, err)

	err = nginxProcessManager.Reload()
	assert.Nil(t, err)
}

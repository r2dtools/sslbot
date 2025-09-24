//go:build nginx

package commondir

import (
	"slices"
	"strings"
	"sync"
	"testing"

	nginxconfig "github.com/r2dtools/gonginxconf/config"
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
	"github.com/stretchr/testify/assert"
)

const nginxCommonDir = "/var/www/html/"

func TestNginxCommonDir(t *testing.T) {
	host := "example2.com"

	command, query, nginxWebServer, rv := getNginxCommonDir(t)
	defer rv.Rollback()

	commonDir := query.GetCommonDirStatus(host)
	assert.False(t, commonDir.Enabled)
	assert.Empty(t, commonDir.Root)

	err := command.EnableCommonDir(host)
	assert.Nil(t, err)
	commonDir = query.GetCommonDirStatus(host)
	assert.True(t, commonDir.Enabled)
	assert.Equal(t, nginxCommonDir, commonDir.Root)

	blocks := nginxWebServer.Config.FindServerBlocksByServerName(host)
	assert.Len(t, blocks, 1)

	block := blocks[0]
	locations := block.FindLocationBlocks()
	assert.Len(t, locations, 2)

	acmeBlockExists := slices.ContainsFunc(locations, func(block nginxconfig.LocationBlock) bool {
		return strings.Contains(block.GetLocationMatch(), acmeLocation)
	})
	assert.True(t, acmeBlockExists)

	err = command.DisableCommonDir(host)
	assert.Nil(t, err)
	commonDir = query.GetCommonDirStatus(host)
	assert.False(t, commonDir.Enabled)
	assert.Empty(t, commonDir.Root)

	locations = block.FindLocationBlocks()
	assert.Len(t, locations, 1)

	acmeBlockExists = slices.ContainsFunc(locations, func(block nginxconfig.LocationBlock) bool {
		return strings.Contains(block.GetLocationMatch(), acmeLocation)
	})
	assert.False(t, acmeBlockExists)
}

func getNginxCommonDir(t *testing.T) (*NginxCommonDirChangeCommand, *NginxCommonDirQuery, webserver.NginxWebServer, reverter.Reverter) {
	config, err := config.GetConfig()
	assert.Nil(t, err)
	options := config.ToMap()

	nginxWebServer, err := webserver.GetNginxWebServer(options)
	assert.Nil(t, err)

	rv, err := reverter.CreateReverter(nginxWebServer, &logger.TestLogger{T: t})
	assert.Nil(t, err)

	command, err := CreateCommonDirChangeCommand(config, nginxWebServer, rv, &logger.TestLogger{T: t}, &sync.Mutex{})
	assert.Nil(t, err)

	query, err := CreateCommonDirStatusQuery(nginxWebServer)
	assert.Nil(t, err)

	return command, query, *nginxWebServer, rv
}

//go:build apache

package webserver

import (
	"testing"

	"github.com/r2dtools/sslbot/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApacheGetVHosts(t *testing.T) {
	apacheWebServer := getApacheWebServer(t)
	hosts, err := apacheWebServer.GetVhosts()
	assert.Nil(t, err)
	assert.Len(t, hosts, 2)
}

func TestApacheGetVHost(t *testing.T) {
	apacheWebServer := getApacheWebServer(t)
	host, err := apacheWebServer.GetVhostByName("example2.com")
	assert.Nil(t, err)
	assert.NotNil(t, host)
	assert.Equal(t, []string{"www.example2.com", "ipv4.example2.com"}, host.Aliases)
	assert.Equal(t, "/etc/apache2/sites-enabled/example2.com.conf", host.FilePath)
	assert.False(t, host.Ssl)
	assert.Equal(t, "/var/www/html", host.DocRoot)
	assert.Len(t, host.Addresses, 1)
	assert.Equal(t, "example2.com", host.ServerName)
}

func getApacheWebServer(t *testing.T) *ApacheWebServer {
	config, err := config.GetConfig()
	assert.Nil(t, err)
	apacheWebServer, err := GetApacheWebServer(config.ToMap())
	require.Nil(t, err)

	return apacheWebServer
}

//go:build nginx

package webserver

import (
	"testing"

	"github.com/r2dtools/sslbot/config"
	"github.com/stretchr/testify/assert"
)

func TestNginxGetVHosts(t *testing.T) {
	nginxWebServer := getNginxWebServer(t)
	hosts, err := nginxWebServer.GetVhosts()
	assert.Nil(t, err)
	assert.Len(t, hosts, 5)
}

func TestNginxGetVHost(t *testing.T) {
	nginxWebServer := getNginxWebServer(t)
	host, err := nginxWebServer.GetVhostByName("example2.com")
	assert.Nil(t, err)
	assert.NotNil(t, host)
	assert.Equal(t, []string{"www.example2.com"}, host.Aliases)
	assert.Equal(t, "/etc/nginx/sites-enabled/example2.com.conf", host.FilePath)
	assert.True(t, host.Ssl)
	assert.Equal(t, "/var/www/html", host.DocRoot)
	assert.Len(t, host.Addresses, 4)
	assert.Equal(t, "example2.com", host.ServerName)

	host, err = nginxWebServer.GetVhostByName("example3.com")
	assert.Nil(t, err)
	assert.NotNil(t, host)
	assert.Len(t, host.Aliases, 0)
	assert.Equal(t, "/etc/nginx/sites-enabled/example3.com.conf", host.FilePath)
	assert.False(t, host.Ssl)
	assert.Equal(t, "/var/www/html", host.DocRoot)
	assert.Len(t, host.Addresses, 2)
	assert.Equal(t, "example3.com", host.ServerName)

	host, err = nginxWebServer.GetVhostByName("example4.com")
	assert.Nil(t, err)
	assert.NotNil(t, host)
	assert.Equal(t, []string{"www.example4.com", "ipv4.example4.com"}, host.Aliases)
	assert.Equal(t, "/etc/nginx/sites-enabled/example4.com.conf", host.FilePath)
	assert.True(t, host.Ssl)
	assert.Equal(t, "/var/www/html", host.DocRoot)
	assert.Len(t, host.Addresses, 2)
	assert.Equal(t, "example4.com", host.ServerName)

	host, err = nginxWebServer.GetVhostByName("webmail.r2dtools.work.gd")
	assert.Nil(t, err)
	assert.NotNil(t, host)
	assert.Len(t, host.Aliases, 0)
	assert.Equal(t, "/etc/nginx/sites-enabled/webmail.conf", host.FilePath)
	assert.True(t, host.Ssl)
	assert.Equal(t, "", host.DocRoot)
	assert.Len(t, host.Addresses, 2)
	assert.Equal(t, "webmail.r2dtools.work.gd", host.ServerName)
}

func getNginxWebServer(t *testing.T) *NginxWebServer {
	config, err := config.GetConfig()
	assert.Nil(t, err)
	nginxWebServer, err := GetNginxWebServer(config.ToMap())
	assert.Nil(t, err)

	return nginxWebServer
}

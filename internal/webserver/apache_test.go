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

func getApacheWebServer(t *testing.T) *ApacheWebServer {
	config, err := config.GetConfig()
	assert.Nil(t, err)
	apacheWebServer, err := GetApacheWebServer(config.ToMap())
	require.Nil(t, err)

	return apacheWebServer
}

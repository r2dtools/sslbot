//go:build common

package webserver

import (
	"testing"

	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/stretchr/testify/assert"
)

func TestMergeVHosts(t *testing.T) {
	hosts := []dto.VirtualHost{
		{
			FilePath:   "/path",
			ServerName: "example.com",
			DocRoot:    "/var/www/html",
			WebServer:  "nginx",
			Aliases:    []string{"www.example.com", "alias.example.com"},
			Ssl:        false,
			Addresses: []dto.VirtualHostAddress{
				{
					IsIpv6: false,
					Host:   "127.0.0.1",
					Port:   "80",
				},
				{
					IsIpv6: true,
					Host:   "[::]",
					Port:   "80",
				},
			},
		},
		{
			FilePath:   "/path2",
			ServerName: "example.com",
			DocRoot:    "/var/www/html",
			WebServer:  "nginx",
			Aliases:    []string{"www.example.com", "alias.example.com"},
			Ssl:        true,
			Addresses: []dto.VirtualHostAddress{
				{
					IsIpv6: false,
					Host:   "127.0.0.1",
					Port:   "443",
				},
				{
					IsIpv6: true,
					Host:   "[::]",
					Port:   "443",
				},
			},
		},
	}

	mergedHosts := mergeVhosts(hosts)
	assert.Len(t, mergedHosts, 1)

	mergedHost := mergedHosts[0]
	assert.True(t, mergedHost.Ssl)
	assert.Equal(t, []string{"www.example.com", "alias.example.com"}, mergedHost.Aliases)
	assert.Equal(t, "/var/www/html", mergedHost.DocRoot)
	assert.Equal(t, "example.com", mergedHost.ServerName)
	assert.Len(t, mergedHost.Addresses, 4)
}

//go:build apache

package deploy

import (
	"testing"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApacheDeployCertificateToNonSslHost(t *testing.T) {
	deployer, apacheWebServer, rv := getApacheDeployer(t)
	defer rv.Rollback()

	hosts, err := apacheWebServer.GetVhosts()
	require.Nilf(t, err, "get apache hosts error: %v", err)

	servername := "example2.com"
	host := findHost(servername, hosts)
	require.NotNilf(t, host, "host %s not found", servername)

	configPath, _, err := deployer.DeployCertificate(host, "/usr/local/r2dtools/var/default/certificates/example2.com.crt", "/usr/local/r2dtools/var/default/certificates/example2.com.key")
	require.Nilf(t, err, "deploy certificate error: %v", err)
	require.Equal(t, "/etc/apache2/sites-available/example2.com-ssl.conf", configPath)

	hosts, err = apacheWebServer.GetVhosts()
	require.Nilf(t, err, "get apache hosts after deploy error: %v", err)

	host = findHost("example2.com", hosts)
	require.True(t, host.Ssl)
}

func TestApacheDeployCertificateToSslHost(t *testing.T) {
	deployer, apacheWebServer, rv := getApacheDeployer(t)
	defer rv.Rollback()

	hosts, err := apacheWebServer.GetVhosts()
	require.Nilf(t, err, "get apache hosts error: %v", err)

	servername := "example.com"
	host := findHost(servername, hosts)
	require.NotNilf(t, host, "host %s not found", servername)

	configPath, _, err := deployer.DeployCertificate(host, "/usr/local/r2dtools/var/default/certificates/example.com.crt", "/usr/local/r2dtools/var/default/certificates/example.com.key")
	require.Nilf(t, err, "deploy certificate error: %v", err)
	require.Equal(t, "/etc/apache2/sites-enabled/example.com.conf", configPath)

	hosts, err = apacheWebServer.GetVhosts()
	require.Nilf(t, err, "get apache hosts after deploy error: %v", err)

	host = findHost("example.com", hosts)
	require.True(t, host.Ssl)
}

func TestApacheEnsureSslPortIsListened(t *testing.T) {
	deployer, apacheWebServer, rv := getApacheDeployer(t)
	defer rv.Rollback()

	expected := `
# If you just change the port or add more ports here, you will likely also
# have to change the VirtualHost statement in
# /etc/apache2/sites-enabled/000-default.conf

Listen 80

<IfModule ssl_module >
	Listen 443
</IfModule>

<IfModule mod_gnutls.c >
	Listen 443
</IfModule>
<IfModule mod_ssl.c >
	Listen 8443
</IfModule>

`
	deployer.ensureSslPortIsListened("8443")
	portsConfigFile := apacheWebServer.Config.GetConfigFile("ports.conf")
	content, err := portsConfigFile.Dump()
	require.Nil(t, err)
	require.Equal(t, expected, content)
}

func TestApacheGetIpFromListen(t *testing.T) {
	deployer, _, _ := getApacheDeployer(t)

	items := []struct {
		listen   string
		expected string
	}{
		{
			listen:   "443",
			expected: "",
		},
		{
			listen:   "8443 htts",
			expected: "",
		},
		{
			listen:   "1.1.1.1:8443 htts",
			expected: "1.1.1.1",
		},
		{
			listen:   "[2001:db8:cafe::1]:8443 https",
			expected: "[2001:db8:cafe::1]",
		},
	}

	for _, item := range items {
		port := deployer.getIPFromListen(item.listen)
		require.Equal(t, item.expected, port)
	}
}

func getApacheDeployer(t *testing.T) (*ApacheCertificateDeployer, webserver.ApacheWebServer, reverter.Reverter) {
	config, err := config.GetConfig()
	assert.Nil(t, err)
	apacheWebServer, err := webserver.GetApacheWebServer(config.ToMap())
	require.Nil(t, err)

	log := &logger.TestLogger{T: t}
	rv, err := reverter.CreateReverter(apacheWebServer, log)
	require.Nil(t, err)

	deployer, err := GetCertificateDeployer(apacheWebServer, rv, log)
	require.Nil(t, err)

	apacheDeployer, ok := deployer.(*ApacheCertificateDeployer)
	require.True(t, ok)

	return apacheDeployer, *apacheWebServer, rv
}

func findHost(servername string, hosts []dto.VirtualHost) *dto.VirtualHost {
	for _, host := range hosts {
		if host.ServerName == servername {
			return &host
		}
	}

	return nil
}

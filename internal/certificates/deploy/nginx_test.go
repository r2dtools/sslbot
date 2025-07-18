package deploy

import (
	"testing"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
	"github.com/stretchr/testify/assert"
)

func TestDeployCertificateToNonSslHost(t *testing.T) {
	deployer, nginxWebServer, rv := getNginxDeployer(t)
	defer rv.Rollback()

	hosts, err := nginxWebServer.GetVhosts()
	assert.Nilf(t, err, "get nginx hosts error: %v", err)

	servername := "example3.com"
	host := findHost(servername, hosts)
	assert.NotNilf(t, host, "host %s not found", servername)

	configPath, _, err := deployer.DeployCertificate(host, "test/certificate/example.com.key", "test/certificate/example.com.crt")
	assert.Nilf(t, err, "deploy certificate error: %v", err)
	assert.Equal(t, "/etc/nginx/sites-available/example3.com-ssl.conf", configPath)

	hosts, err = nginxWebServer.GetVhosts()
	assert.Nilf(t, err, "get nginx hosts after deploy error: %v", err)

	host = findHost("example3.com", hosts)
	assert.True(t, host.Ssl)
}

func TestDeployCertificateToSslHost(t *testing.T) {
	deployer, nginxWebServer, rv := getNginxDeployer(t)
	defer rv.Rollback()

	hosts, err := nginxWebServer.GetVhosts()
	assert.Nilf(t, err, "get nginx hosts error: %v", err)

	servername := "example2.com"
	host := findHost(servername, hosts)
	assert.NotNilf(t, host, "host %s not found", servername)
	assert.True(t, host.Ssl)

	configPath, _, err := deployer.DeployCertificate(host, "test/certificate/example2.com.key", "test/certificate/example2.com.crt")
	assert.Nilf(t, err, "deploy certificate error: %v", err)
	assert.Equal(t, "/etc/nginx/sites-enabled/example2.com.conf", configPath)

	hosts, err = nginxWebServer.GetVhosts()
	assert.Nilf(t, err, "get nginx hosts after deploy error: %v", err)

	host = findHost("example2.com", hosts)
	assert.True(t, host.Ssl)
}

func getNginxDeployer(t *testing.T) (CertificateDeployer, webserver.NginxWebServer, reverter.Reverter) {
	config, err := config.GetConfig()
	assert.Nil(t, err)
	nginxWebServer, err := webserver.GetNginxWebServer(config.ToMap())
	assert.Nil(t, err)

	log := &logger.TestLogger{T: t}
	rv, err := reverter.CreateReverter(nginxWebServer, log)
	assert.Nil(t, err)

	deployer, err := GetCertificateDeployer(nginxWebServer, rv, log)
	assert.Nil(t, err)

	return deployer, *nginxWebServer, rv
}

func findHost(servername string, hosts []dto.VirtualHost) *dto.VirtualHost {
	for _, host := range hosts {
		if host.ServerName == servername {
			return &host
		}
	}

	return nil
}

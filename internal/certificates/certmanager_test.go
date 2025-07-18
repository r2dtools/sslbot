package certificates

import (
	"path/filepath"
	"testing"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates/request"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
	"github.com/stretchr/testify/assert"
)

const (
	varDir         = "/usr/local/r2dtools/var"
	certBotWorkDir = "/etc/letsencrypt/live"
)

type testCertManager struct {
	*CertificateManager

	reverter reverter.Reverter
	wServer  webserver.WebServer
}

type testReverter struct {
	reverter.Reverter
}

func (r *testReverter) Commit() error {
	return nil
}

func TestGetStorageCertificates(t *testing.T) {
	wServer := createNginxWebserver(t)
	sReverter := createReverter(t, wServer)
	certManager := createCertManager(t, wServer, sReverter)

	certs, err := certManager.GetStorageCertificates()
	assert.Nil(t, err)
	assert.Len(t, certs, 6)
}

func TestGetStorageCertificate(t *testing.T) {
	wServer := createNginxWebserver(t)
	sReverter := createReverter(t, wServer)
	certManager := createCertManager(t, wServer, sReverter)

	cert, err := certManager.GetStorageCertificate("example.com", "default")
	assert.Nil(t, err)
	assert.Equal(t, "example.com", cert.CN)

	cert, err = certManager.GetStorageCertificate("example2.com", "certbot")
	assert.Nil(t, err)
	assert.Equal(t, "example2.com", cert.CN)

	cert, err = certManager.GetStorageCertificate("example2.com", "lego")
	assert.Nil(t, err)
	assert.Equal(t, "example2.com", cert.CN)
}

func TestAddRemoveStorageCertificate(t *testing.T) {
	wServer := createNginxWebserver(t)
	sReverter := createReverter(t, wServer)
	certManager := createCertManager(t, wServer, sReverter)

	certPath, certContent, err := certManager.GetStorageCertificateAsString("example.com", "default")
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(varDir, "default", "certificates", "example.com.pem"), certPath)

	certPath, err = certManager.AddStorageCertificate("example3.com", certContent)
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(varDir, "default", "certificates", "example3.com.pem"), certPath)

	cert, err := certManager.GetStorageCertificate("example3.com", "default")
	assert.Nil(t, err)
	assert.Equal(t, "example.com", cert.CN)

	err = certManager.RemoveStorageCertificate("example3.com", "default")
	assert.Nil(t, err)

	_, err = certManager.GetStorageCertificate("example3.com", "default")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestAssignCertificateToNginxSslHost(t *testing.T) {
	wServer := createNginxWebserver(t)
	sReverter := createReverter(t, wServer)
	certManager := createCertManager(t, wServer, sReverter)
	defer sReverter.Rollback()

	vhost, err := wServer.GetVhostByName("example4.com")
	assert.Nil(t, err)
	assert.NotNilf(t, vhost, "host not found")
	assert.Equal(t, "example.com", vhost.Certificate.CN)

	request := request.AssignRequest{
		ServerName:  "example4.com",
		WebServer:   "nginx",
		CertName:    "example2.com",
		StorageType: string(Lego),
	}

	cert, err := certManager.Assign(request)
	assert.Nil(t, err)
	assert.Equal(t, "example2.com", cert.CN)

	vhost, err = wServer.GetVhostByName("example4.com")
	assert.Nil(t, err)
	assert.NotNilf(t, vhost, "host not found")
	assert.Equal(t, "example2.com", vhost.Certificate.CN)
}

func TestUploadCertificateToNginxNonSslHost(t *testing.T) {
	wServer := createNginxWebserver(t)
	sReverter := createReverter(t, wServer)
	certManager := createCertManager(t, wServer, sReverter)
	defer sReverter.Rollback()
	defer certManager.RemoveStorageCertificate("example3.com", "default")

	_, certContent, err := certManager.GetStorageCertificateAsString("example2.com", "lego")
	assert.Nil(t, err)

	request := request.UploadRequest{
		ServerName:     "example3.com",
		WebServer:      "nginx",
		CertName:       "example3.com",
		PemCertificate: certContent,
	}
	vhost, err := wServer.GetVhostByName("example3.com")
	assert.Nil(t, err)
	assert.Nil(t, vhost.Certificate)

	cert, err := certManager.Upload(request)
	assert.Nil(t, err)
	assert.Equal(t, "example2.com", cert.CN)

	vhost, err = wServer.GetVhostByName("example3.com")
	assert.Nil(t, err)
	assert.Equal(t, "example2.com", vhost.Certificate.CN)
	assert.True(t, vhost.Ssl)
}

func createCertManager(
	t *testing.T,
	wServer webserver.WebServer,
	sReverter reverter.Reverter,
) *CertificateManager {
	config := &config.Config{
		CertBotWokrDir: certBotWorkDir,
		VarDir:         varDir,
	}
	log := &logger.TestLogger{T: t}
	certManager, err := CreateCertificateManager(
		config,
		func(code string, options map[string]string) (webserver.WebServer, error) {
			return wServer, nil
		},
		func(wServer webserver.WebServer, logger logger.Logger) (reverter.Reverter, error) {
			return sReverter, nil
		},
		log,
	)
	assert.Nil(t, err)

	return certManager
}

func createNginxWebserver(t *testing.T) *webserver.NginxWebServer {
	config, err := config.GetConfig()
	assert.Nil(t, err)
	nginxWebServer, err := webserver.GetNginxWebServer(config.ToMap())
	assert.Nil(t, err)

	return nginxWebServer
}

func createReverter(t *testing.T, wServer webserver.WebServer) reverter.Reverter {
	rev, err := reverter.CreateReverter(wServer, &logger.TestLogger{T: t})
	assert.Nil(t, err)

	return &testReverter{
		Reverter: rev,
	}
}

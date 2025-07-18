package certbot

import (
	"path/filepath"
	"testing"

	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/stretchr/testify/assert"
)

const workDir = "/etc/letsencrypt/live"

func TestGetCertificates(t *testing.T) {
	storage := getStorage(t)

	certs, err := storage.GetCertificates()
	assert.Nil(t, err)
	assert.Len(t, certs, 2)

	cert, ok := certs["example.com"]
	assert.True(t, ok)
	assert.Equal(t, "example.com", cert.CN)

	cert, ok = certs["example2.com"]
	assert.True(t, ok)
	assert.Equal(t, "example2.com", cert.CN)
}

func TestGetCertificate(t *testing.T) {
	storage := getStorage(t)

	cert, err := storage.GetCertificate("example.com")
	assert.Nil(t, err)

	assert.Equal(t, "example.com", cert.CN)
}

func TestGetCertificatePath(t *testing.T) {
	storage := getStorage(t)

	certPath, keyPath, err := storage.GetCertificatePath("example.com")
	assert.Nil(t, err)

	assert.Equal(t, filepath.Join(workDir, "example.com", "fullchain.pem"), certPath)
	assert.Equal(t, filepath.Join(workDir, "example.com", "privkey.pem"), keyPath)
}

func getStorage(t *testing.T) *CertBotStorage {
	return &CertBotStorage{
		path:   workDir,
		logger: &logger.TestLogger{T: t},
	}
}

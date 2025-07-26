package lego

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/stretchr/testify/assert"
)

const workDir = "/usr/local/r2dtools/var/lego/certificates"

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

	assert.Equal(t, filepath.Join(workDir, "example.com.pem"), certPath)
	assert.Equal(t, filepath.Join(workDir, "example.com.pem"), keyPath)
}

func TestRemoveCertificate(t *testing.T) {
	storage := getStorage(t)

	certPath, data, err := storage.GetCertificateAsString("example2.com")
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(storage.path, "example2.com.pem"), certPath)

	err = os.WriteFile(filepath.Join(workDir, "example3.com.pem"), []byte(data), 0644)
	assert.Nil(t, err)

	cert, err := storage.GetCertificate("example3.com")
	assert.Nil(t, err)
	assert.Equal(t, "example2.com", cert.CN)

	err = storage.RemoveCertificate("example3.com")
	assert.Nil(t, err)

	certs, err := storage.GetCertificates()
	assert.Nil(t, err)

	_, ok := certs["example3.com"]
	assert.False(t, ok)
}

func getStorage(t *testing.T) *LegoStorage {
	return &LegoStorage{
		RWMutex: &sync.RWMutex{},
		path:    workDir,
		logger:  &logger.TestLogger{T: t},
	}
}

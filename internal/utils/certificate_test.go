package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCertificateFromFile(t *testing.T) {
	cert, err := GetCertificateFromFile("../../test/certificate/example.com.crt")
	assert.Nil(t, err)
	assert.Equal(t, []string{"example.com", "www.example.com"}, cert.DNSNames)
}

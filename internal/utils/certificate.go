package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/unknwon/com"
)

func GetX509CertificateFromRequest(domain string) ([]*x509.Certificate, error) {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Minute}, "tcp", domain+":443", &tls.Config{InsecureSkipVerify: true})

	if err != nil {
		return nil, err
	}

	defer conn.Close()
	return conn.ConnectionState().PeerCertificates, nil
}

func ConvertX509CertificateToIntCert(certificate *x509.Certificate, roots []*x509.Certificate) *dto.Certificate {
	certPool := x509.NewCertPool()

	for _, root := range roots {
		certPool.AddCert(root)
	}

	opts := x509.VerifyOptions{
		Intermediates: certPool,
	}
	_, err := certificate.Verify(opts)
	isValid := err == nil
	cert := dto.Certificate{
		DNSNames:       certificate.DNSNames,
		CN:             certificate.Subject.CommonName,
		EmailAddresses: certificate.EmailAddresses,
		Organization:   certificate.Subject.Organization,
		Country:        certificate.Subject.Country,
		Province:       certificate.Subject.Province,
		Locality:       certificate.Subject.Locality,
		IsCA:           certificate.IsCA,
		ValidFrom:      certificate.NotBefore.Format(time.RFC822Z),
		ValidTo:        certificate.NotAfter.Format(time.RFC822Z),
		Issuer: dto.Issuer{
			CN:           certificate.Issuer.CommonName,
			Organization: certificate.Issuer.Organization,
		},
		IsValid: isValid,
	}

	return &cert
}

func GetCertificateForDomainFromRequest(domain string) (*dto.Certificate, error) {
	certs, err := GetX509CertificateFromRequest(domain)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 {
		return nil, nil
	}

	var roots []*x509.Certificate

	if len(certs) > 1 {
		roots = certs[1:]
	}

	return ConvertX509CertificateToIntCert(certs[0], roots), nil
}

func GetCertificateFromFile(path string) (*dto.Certificate, error) {
	if !com.IsFile(path) {
		return nil, fmt.Errorf("certificate file '%s' does not exist", path)
	}

	certContent, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("could not read certificate content: %v", err)
	}

	var bCerts []*x509.Certificate

	for {
		block, rest := pem.Decode([]byte(certContent))

		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			x509Cert, err := x509.ParseCertificate(block.Bytes)

			if err != nil {
				return nil, errors.New("could not parse certificate")
			}

			bCerts = append(bCerts, x509Cert)
		}

		certContent = rest
	}

	if len(bCerts) == 0 {
		return nil, errors.New("could not parse certificate")
	}

	cert := ConvertX509CertificateToIntCert(bCerts[0], bCerts[1:])

	return cert, nil
}

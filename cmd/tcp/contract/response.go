package contract

import (
	"github.com/r2dtools/agentintegration"
	"github.com/r2dtools/sslbot/internal/dto"
)

func ConvertVirtualHost(vhost *dto.VirtualHost) *agentintegration.VirtualHost {
	addresses := []agentintegration.VirtualHostAddress{}

	for _, address := range vhost.Addresses {
		addresses = append(addresses, agentintegration.VirtualHostAddress{
			IsIpv6: address.IsIpv6,
			Host:   address.Host,
			Port:   address.Port,
		})
	}

	var certificate *agentintegration.Certificate

	if vhost.Certificate != nil {
		certificate = ConvertCertificate(vhost.Certificate)
	}

	return &agentintegration.VirtualHost{
		FilePath:    vhost.FilePath,
		ServerName:  vhost.ServerName,
		DocRoot:     vhost.DocRoot,
		WebServer:   vhost.WebServer,
		Aliases:     vhost.Aliases,
		Ssl:         vhost.Ssl,
		Addresses:   addresses,
		Certificate: certificate,
	}
}

func ConvertVirtualHosts(vhosts []dto.VirtualHost) []agentintegration.VirtualHost {
	cVhosts := []agentintegration.VirtualHost{}

	for _, vhost := range vhosts {
		cVhosts = append(cVhosts, *ConvertVirtualHost(&vhost))
	}

	return cVhosts
}

func ConvertCertificate(cert *dto.Certificate) *agentintegration.Certificate {
	issuer := agentintegration.Issuer{
		CN:           cert.Issuer.CN,
		Organization: cert.Issuer.Organization,
	}

	return &agentintegration.Certificate{
		CN:             cert.CN,
		ValidFrom:      cert.ValidFrom,
		ValidTo:        cert.ValidTo,
		DNSNames:       cert.DNSNames,
		EmailAddresses: cert.EmailAddresses,
		Organization:   cert.Organization,
		Province:       cert.Province,
		Country:        cert.Country,
		Locality:       cert.Locality,
		IsCA:           cert.IsCA,
		IsValid:        cert.IsValid,
		Issuer:         issuer,
	}
}

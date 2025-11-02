package webserver

import (
	"fmt"
	"strings"

	"github.com/r2dtools/goapacheconf"
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/utils"
	"github.com/r2dtools/sslbot/internal/webserver/processmng"
)

const (
	ApacheCertKeyDirective = "SSLCertificateKeyFile"
	ApacheCertDirective    = "SSLCertificateFile"
)

type ApacheWebServer struct {
	Config  *goapacheconf.Config
	root    string
	options map[string]string
}

func (a *ApacheWebServer) GetCode() string {
	return WebServerApacheCode
}

func (a *ApacheWebServer) GetVhosts() ([]dto.VirtualHost, error) {
	var vhosts []dto.VirtualHost

	aVhosts := a.Config.FindVirtualHostBlocks()

	for _, aVhost := range aVhosts {
		var addresses []dto.VirtualHostAddress

		for _, address := range aVhost.GetAddresses() {
			addresses = append(addresses, dto.VirtualHostAddress{
				IsIpv6: address.IsIpv6,
				Host:   address.Host,
				Port:   address.Port,
			})
		}

		serverNames := aVhost.GetServerNames()

		if len(serverNames) == 0 {
			continue
		}

		vhost := dto.VirtualHost{
			FilePath:    strings.Trim(aVhost.FilePath, "\""),
			ServerName:  strings.Trim(serverNames[0], "\""),
			DocRoot:     strings.Trim(aVhost.GetDocumentRoot(), "\""),
			Aliases:     aVhost.GetServerAliases(),
			Ssl:         aVhost.HasSSL(),
			WebServer:   WebServerApacheCode,
			Addresses:   addresses,
			Certificate: getApacheCertificate(aVhost),
		}
		vhosts = append(vhosts, vhost)
	}

	vhosts = filterVhosts(vhosts)
	vhosts = mergeVhosts(vhosts)

	return vhosts, nil
}

func (a *ApacheWebServer) GetVhostByName(serverName string) (*dto.VirtualHost, error) {
	vhosts, err := a.GetVhosts()

	if err != nil {
		return nil, err
	}

	return getVhostByName(vhosts, serverName), nil
}

func (a *ApacheWebServer) GetProcessManager() (ProcessManager, error) {
	return processmng.GetApacheProcessManager()
}

func getApacheCertificate(virtualHostBlock goapacheconf.VirtualHostBlock) *dto.Certificate {
	certDirectives := virtualHostBlock.FindDirectives(ApacheCertDirective)

	if len(certDirectives) == 0 {
		return nil
	}

	certDirective := certDirectives[len(certDirectives)-1]
	cert, _ := utils.GetCertificateFromFile(certDirective.GetFirstValue())

	return cert
}

func GetApacheWebServer(options map[string]string) (*ApacheWebServer, error) {
	root := options[config.ApacheRootOpt]
	config, err := goapacheconf.GetConfig(root, "")

	if err != nil {
		return nil, fmt.Errorf("could not parse apache config: %v", err)
	}

	return &ApacheWebServer{
		Config:  config,
		root:    root,
		options: options,
	}, nil
}

package webserver

import (
	"fmt"
	"strings"

	nginxConfig "github.com/r2dtools/gonginxconf/config"
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/utils"
	"github.com/r2dtools/sslbot/internal/webserver/processmng"
)

const (
	NginxCertKeyDirective = "ssl_certificate_key"
	NginxCertDirective    = "ssl_certificate"
)

type NginxWebServer struct {
	Config  *nginxConfig.Config
	root    string
	options map[string]string
}

func (nws *NginxWebServer) GetCode() string {
	return WebServerNginxCode
}

func (nws *NginxWebServer) GetVhostByName(serverName string) (*dto.VirtualHost, error) {
	vhosts, err := nws.GetVhosts()

	if err != nil {
		return nil, err
	}

	return getVhostByName(vhosts, serverName), nil
}

func (nws *NginxWebServer) GetVhosts() ([]dto.VirtualHost, error) {
	var vhosts []dto.VirtualHost

	nVhosts := nws.Config.FindServerBlocks()

	for _, nVhost := range nVhosts {
		var addresses []dto.VirtualHostAddress

		for _, address := range nVhost.GetAddresses() {
			addresses = append(addresses, dto.VirtualHostAddress{
				IsIpv6: address.IsIpv6,
				Host:   address.Host,
				Port:   address.Port,
			})
		}

		serverNames := nVhost.GetServerNames()

		if len(serverNames) == 0 {
			continue
		}

		aliases := []string{}

		if len(serverNames) > 1 {
			aliases = serverNames[1:]
		}

		vhost := dto.VirtualHost{
			FilePath:    strings.Trim(nVhost.FilePath, "\""),
			ServerName:  strings.Trim(serverNames[0], "\""),
			DocRoot:     strings.Trim(nVhost.GetDocumentRoot(), "\""),
			Aliases:     aliases,
			Ssl:         nVhost.HasSSL(),
			WebServer:   WebServerNginxCode,
			Addresses:   addresses,
			Certificate: getNginxCertificate(nVhost),
		}
		vhosts = append(vhosts, vhost)
	}

	vhosts = filterVhosts(vhosts)
	vhosts = mergeVhosts(vhosts)

	return vhosts, nil
}

func (nws *NginxWebServer) GetProcessManager() (ProcessManager, error) {
	return processmng.GetNginxProcessManager()
}

func GetNginxWebServer(options map[string]string) (*NginxWebServer, error) {
	root := options[config.NginxRootOpt]
	config, err := nginxConfig.GetConfig(root, "", false)

	if err != nil {
		return nil, fmt.Errorf("could not parse nginx config: %v", err)
	}

	return &NginxWebServer{
		Config:  config,
		root:    root,
		options: options,
	}, nil
}

func getNginxCertificate(serverBlock nginxConfig.ServerBlock) *dto.Certificate {
	certDirectives := serverBlock.FindDirectives(NginxCertDirective)

	if len(certDirectives) == 0 {
		return nil
	}

	certDirective := certDirectives[len(certDirectives)-1]
	cert, _ := utils.GetCertificateFromFile(certDirective.GetFirstValue())

	return cert
}

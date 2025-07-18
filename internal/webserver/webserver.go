package webserver

import (
	"fmt"

	"github.com/r2dtools/sslbot/internal/dto"
)

const (
	WebServerNginxCode  = "nginx"
	WebServerApacheCode = "apache"
)

type HostManager interface {
	Enable(configFilePath, originConfigFilePath string) (string, error)
	Disable(configFilePath string) error
}

type ProcessManager interface {
	Reload() error
}

func GetSupportedWebServers() []string {
	return []string{WebServerNginxCode}
}

type WebServer interface {
	GetVhostByName(serverName string) (*dto.VirtualHost, error)
	GetVhosts() ([]dto.VirtualHost, error)
	GetCode() string
	GetProcessManager() (ProcessManager, error)
}

func GetWebServer(webServerCode string, options map[string]string) (WebServer, error) {
	var webServer WebServer
	var err error

	switch webServerCode {
	case WebServerNginxCode:
		webServer, err = GetNginxWebServer(options)
	default:
		err = fmt.Errorf("webserver '%s' is not supported", webServerCode)
	}

	return webServer, err
}

func getVhostByName(vhosts []dto.VirtualHost, serverName string) *dto.VirtualHost {
	for _, vhost := range vhosts {
		if vhost.ServerName == serverName {
			return &vhost
		}
	}

	return nil
}

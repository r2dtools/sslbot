package hostmng

import (
	"fmt"

	"github.com/r2dtools/sslbot/internal/webserver"
)

type HostManager interface {
	Enable(configFilePath, originConfigFilePath string) (string, error)
	Disable(configFilePath string) error
}

func CreateHostManager(webServer webserver.WebServer) (HostManager, error) {
	switch webServer.GetCode() {
	case webserver.WebServerNginxCode:
		return &NginxHostManager{}, nil
	default:
		return nil, fmt.Errorf("webserver %s not supported", webServer.GetCode())
	}
}

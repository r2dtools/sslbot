package commondir

import (
	"fmt"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

type CommonDir struct {
	Enabled bool
	Root    string
}

type CommonDirQuery interface {
	GetCommonDirStatus(serverName string) CommonDir
}

type CommonDirChangeCommand interface {
	EnableCommonDir(serverName string) error
	DisableCommonDir(serverName string) error
}

func CreateCommonDirStatusQuery(webServer webserver.WebServer) (*NginxCommonDirQuery, error) {
	switch w := webServer.(type) {
	case *webserver.NginxWebServer:
		return &NginxCommonDirQuery{webServer: w}, nil
	default:
		return nil, fmt.Errorf("webserver %s is not supported", webServer.GetCode())
	}
}

func CreateCommonDirChangeCommand(
	webServer webserver.WebServer,
	reverter reverter.Reverter,
	logger logger.Logger,
	options map[string]string,
) (*NginxCommonDirChangeCommand, error) {
	nginxCommonDir := options[config.NginxAcmeCommonDirOpt]

	switch w := webServer.(type) {
	case *webserver.NginxWebServer:
		return &NginxCommonDirChangeCommand{
			logger:    logger,
			webServer: w,
			reverter:  reverter,
			commonDir: nginxCommonDir,
		}, nil
	default:
		return nil, fmt.Errorf("webserver %s is not supported", webServer.GetCode())
	}
}

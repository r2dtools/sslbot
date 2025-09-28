package deploy

import (
	"fmt"

	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

type CertificateDeployer interface {
	DeployCertificate(vhost *dto.VirtualHost, certPath, certKeyPath string) (string, string, error)
}

func GetCertificateDeployer(webServer webserver.WebServer, reverter reverter.Reverter, logger logger.Logger) (CertificateDeployer, error) {
	switch w := webServer.(type) {
	case *webserver.NginxWebServer:
		return &NginxCertificateDeployer{
			logger:    logger,
			webServer: w,
			reverter:  reverter,
		}, nil
	case *webserver.ApacheWebServer:
		return &ApacheCertificateDeployer{
			logger:    logger,
			webServer: w,
			reverter:  reverter,
		}, nil
	default:
		return nil, fmt.Errorf("could not create deployer: webserver '%s' is not supported", webServer.GetCode())
	}
}

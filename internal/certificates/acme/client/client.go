package client

import (
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates/acme/client/certbot"
	"github.com/r2dtools/sslbot/internal/certificates/acme/client/lego"
	"github.com/r2dtools/sslbot/internal/certificates/request"
	"github.com/r2dtools/sslbot/internal/logger"
)

type AcmeClient interface {
	Issue(docRoot string, request request.IssueRequest) (string, string, error)
}

func CreateAcmeClient(config *config.Config, logger logger.Logger) (AcmeClient, error) {
	if config.CertBotEnabled {
		return certbot.CreateCertBot(config, logger)
	}

	return lego.CreateClient(config, logger)
}

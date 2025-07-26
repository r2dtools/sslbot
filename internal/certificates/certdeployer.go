package certificates

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates/deploy"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/hostmng"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

type CertificateDeployer interface {
	DeployCertificate(
		serverName string,
		certPath string,
		keyPath string,
		preventReload bool,
	) error
}

type NilCertificateDeployer struct {
}

func (d *NilCertificateDeployer) DeployCertificate(
	serverName string,
	certPath string,
	keyPath string,
	preventReload bool,
) error {
	return nil
}

type DefaultCertificateDeployer struct {
	mx        *sync.Mutex
	webServer webserver.WebServer
	reverter  reverter.Reverter
	logger    logger.Logger
}

func (d *DefaultCertificateDeployer) DeployCertificate(
	serverName string,
	certPath string,
	keyPath string,
	preventReload bool,
) error {
	d.mx.Lock()
	defer d.mx.Unlock()

	vhost, err := d.webServer.GetVhostByName(serverName)

	if err != nil {
		return err
	}

	if vhost == nil {
		return fmt.Errorf("could not find virtual host %s", serverName)
	}

	deployer, err := deploy.GetCertificateDeployer(d.webServer, d.reverter, d.logger)

	if err != nil {
		return err
	}

	sslConfigFilePath, originEnabledConfigFilePath, err := deployer.DeployCertificate(vhost, certPath, keyPath)

	if err != nil {
		if rErr := d.reverter.Rollback(); rErr != nil {
			d.logger.Error(fmt.Sprintf("rollback failed on cert deploy: %v", rErr))
		}

		return err
	}

	err = d.enableHost(sslConfigFilePath, originEnabledConfigFilePath)

	if err != nil {
		return err
	}

	if !preventReload {
		if err := d.reloadWebServer(); err != nil {
			return err
		}
	}

	if err = d.reverter.Commit(); err != nil {
		if rErr := d.reverter.Rollback(); rErr != nil {
			d.logger.Error(fmt.Sprintf("failed to commit webserver configuration: %v", rErr))
		}
	}

	return nil
}

func (d *DefaultCertificateDeployer) enableHost(sslConfigFilePath, originEnabledConfigFilePath string) error {
	hostManager, err := hostmng.CreateHostManager(d.webServer)

	if err != nil {
		return err
	}

	enabledConfigPath, err := hostManager.Enable(sslConfigFilePath, filepath.Dir(originEnabledConfigFilePath))

	if err == nil {
		d.reverter.AddConfigToDisable(enabledConfigPath)

		return nil
	}

	if rErr := d.reverter.Rollback(); rErr != nil {
		d.logger.Error(fmt.Sprintf("rollback failed on host enabling: %v", rErr))
	}

	return err

}

func (d *DefaultCertificateDeployer) reloadWebServer() error {
	processManager, err := d.webServer.GetProcessManager()

	if err != nil {
		return err
	}

	err = processManager.Reload()

	if err == nil {
		return nil
	}

	if rErr := d.reverter.Rollback(); rErr != nil {
		d.logger.Error(fmt.Sprintf("rollback failed on webserver reload: %v", rErr))
	}

	return err
}

func createCertificateDeployer(
	config *config.Config,
	webServer webserver.WebServer,
	reverter reverter.Reverter,
	logger logger.Logger,
	mx *sync.Mutex,
) CertificateDeployer {
	if config.CertBotEnabled {
		// certbot has own deployer
		return &NilCertificateDeployer{}
	}

	return &DefaultCertificateDeployer{
		mx:        mx,
		webServer: webServer,
		reverter:  reverter,
		logger:    logger,
	}
}

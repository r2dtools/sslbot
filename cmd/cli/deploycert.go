package cli

import (
	"fmt"
	"path/filepath"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates/deploy"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/hostmng"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var DeployCertificateCmd = &cobra.Command{
	Use:   "deploy-cert",
	Short: "Deploy certificate to a domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.GetConfig()

		if err != nil {
			return err
		}

		log, err := logger.NewLogger(config)

		if err != nil {
			return err
		}

		supportedWebServerCodes := webserver.GetSupportedWebServers()

		if webServerCode == "" {
			return fmt.Errorf("webserver is not specified")
		}

		if serverName == "" {
			return fmt.Errorf("domain is not specified")
		}

		if !slices.Contains(supportedWebServerCodes, webServerCode) {
			return fmt.Errorf("invalid webserver %s", webServerCode)
		}

		webServer, err := webserver.GetWebServer(webServerCode, config.ToMap())

		if err != nil {
			return err
		}

		processManager, err := webServer.GetProcessManager()

		if err != nil {
			return err
		}

		vhost, err := webServer.GetVhostByName(serverName)

		if err != nil {
			return err
		}

		sReverter, err := reverter.CreateReverter(webServer, log)

		if err != nil {
			return err
		}

		if vhost == nil {
			return fmt.Errorf("could not find virtual host '%s'", serverName)
		}

		deployer, err := deploy.GetCertificateDeployer(webServer, sReverter, log)

		if err != nil {
			return err
		}

		sslConfigFilePath, originEnabledConfigFilePath, err := deployer.DeployCertificate(vhost, certPath, certKeyPath)

		if err != nil {
			if rErr := sReverter.Rollback(); rErr != nil {
				log.Error(fmt.Sprintf("failed to rallback webserver configuration on cert deploy: %v", rErr))
			}

			return err
		}

		hostManager, err := hostmng.CreateHostManager(webServer)

		if err != nil {
			return err
		}

		enabledConfigPath, err := hostManager.Enable(sslConfigFilePath, filepath.Dir(originEnabledConfigFilePath))

		if err != nil {
			if rErr := sReverter.Rollback(); rErr != nil {
				log.Error(fmt.Sprintf("failed to rallback webserver configuration on host enabling: %v", rErr))
			}

			return err
		}

		sReverter.AddConfigToDisable(enabledConfigPath)

		if err = processManager.Reload(); err != nil {
			if rErr := sReverter.Rollback(); rErr != nil {
				log.Error(fmt.Sprintf("failed to rallback webserver configuration on webserver reload: %v", rErr))
			}

			return err
		}

		if err = sReverter.Commit(); err != nil {
			if rErr := sReverter.Rollback(); rErr != nil {
				log.Error(fmt.Sprintf("failed to commit webserver configuration: %v", rErr))
			}
		}

		return nil
	},
}

var certPath string
var certKeyPath string

func init() {
	DeployCertificateCmd.PersistentFlags().StringVarP(&serverName, "domain", "d", "", "domain to deploy a certificate")
	DeployCertificateCmd.PersistentFlags().StringVarP(&certPath, "cert", "c", "", "path to a certificate file")
	DeployCertificateCmd.PersistentFlags().StringVarP(&certKeyPath, "key", "k", "", "path to a certificate key path")
}

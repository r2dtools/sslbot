package server

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/r2dtools/agentintegration"
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/modules/certificates"
	"github.com/r2dtools/sslbot/internal/modules/certificates/acme"
	"github.com/r2dtools/sslbot/internal/pkg/logger"
	"github.com/r2dtools/sslbot/internal/pkg/webserver"
	"github.com/spf13/cobra"
)

var IssueCertificateCmd = &cobra.Command{
	Use:   "issue-cert",
	Short: "Secure domain with a certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.GetConfig()

		if err != nil {
			return err
		}

		log, err := logger.NewLogger(config)

		if err != nil {
			return err
		}

		if email == "" {
			return fmt.Errorf("email is not specified")
		}

		if serverName == "" {
			return fmt.Errorf("domain is not specified")
		}

		supportedWebServerCodes := webserver.GetSupportedWebServers()

		if webServerCode == "" {
			return fmt.Errorf("webserver is not specified")
		}

		if !slices.Contains(supportedWebServerCodes, webServerCode) {
			return fmt.Errorf("invalid webserver %s", webServerCode)
		}

		certManager, err := certificates.GetCertificateManager(config, log)

		if err != nil {
			return err
		}

		certData := agentintegration.CertificateIssueRequestData{
			Email:         email,
			ServerName:    serverName,
			WebServer:     webServerCode,
			ChallengeType: acme.HttpChallengeTypeCode,
			Subjects:      aliases,
			Assign:        assign,
		}
		cert, err := certManager.Issue(certData)

		if err != nil {
			return err
		}

		data, err := json.MarshalIndent(cert, "", " ")

		if err != nil {
			return err
		}

		fmt.Println(string(data))

		return nil
	},
}

var email string
var assign bool
var aliases []string

func init() {
	aliases = make([]string, 0)
	IssueCertificateCmd.PersistentFlags().StringVarP(&serverName, "domain", "d", "", "domain to secure")
	IssueCertificateCmd.PersistentFlags().StringVarP(&email, "email", "e", "", "certificate email address")
	IssueCertificateCmd.PersistentFlags().BoolVarP(&assign, "assign", "s", true, "assignt certificate to the domain")
	IssueCertificateCmd.PersistentFlags().StringSliceVarP(&aliases, "alias", "a", nil, "domain aliases that need to be included in the certificate")
}

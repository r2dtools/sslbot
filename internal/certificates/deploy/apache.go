package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/r2dtools/goapacheconf"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

type ApacheCertificateDeployer struct {
	logger    logger.Logger
	webServer *webserver.ApacheWebServer
	reverter  reverter.Reverter
}

func (d *ApacheCertificateDeployer) DeployCertificate(vhost *dto.VirtualHost, certPath, certKeyPath string) (string, string, error) {
	wConfig := d.webServer.Config

	if !wConfig.IsModuleEnabled("ssl") {
		return "", "", errors.New("ssl module is not enabled")
	}

	vHostBlocks := wConfig.FindVirtualHostBlocksByServerName(vhost.ServerName)

	if len(vHostBlocks) == 0 {
		return "", "", fmt.Errorf("apache host %s does not exixst", vhost.ServerName)
	}

	var sslVHostBlock *goapacheconf.VirtualHostBlock
	var err error
	vHostBlock := vHostBlocks[0]

	for _, vHostBlock := range vHostBlocks {
		if vHostBlock.HasSSL() {
			sslVHostBlock = &vHostBlock
		}
	}

	if sslVHostBlock == nil {
		sslVHostBlock, err = d.createSslHost(vhost, vHostBlock)

		if err != nil {
			return "", "", err
		}

		d.reverter.AddConfigToDeletion(sslVHostBlock.FilePath)
	} else {
		d.reverter.BackupConfig(sslVHostBlock.FilePath)
	}

	sslVHostBlock.DeleteDirectiveByName(goapacheconf.SSLCertificateChainFile)
	d.createOrUpdateSingleDirective(sslVHostBlock, goapacheconf.SSLEngine, "on")
	d.createOrUpdateSingleDirective(sslVHostBlock, goapacheconf.SSLCertificateKeyFile, certKeyPath)
	d.createOrUpdateSingleDirective(sslVHostBlock, goapacheconf.SSLCertificateFile, certPath)
	d.removeDangerousForSslRewriteRules(sslVHostBlock)

	sslServerBlockFileName := filepath.Base(sslVHostBlock.FilePath)
	configFile := wConfig.GetConfigFile(sslServerBlockFileName)

	if err = configFile.Dump(); err != nil {
		return "", "", err
	}

	return sslVHostBlock.FilePath, vHostBlock.FilePath, nil
}

func (d *ApacheCertificateDeployer) createSslHost(
	vhost *dto.VirtualHost,
	vHostBlock goapacheconf.VirtualHostBlock,
) (*goapacheconf.VirtualHostBlock, error) {
	content := fmt.Sprintf(
		`
		<IfModule mod_ssl.c>
			%s
		</IfModule>`,
		vHostBlock.Dump(),
	)

	filePath, err := filepath.EvalSymlinks(vHostBlock.FilePath)

	if err != nil {
		return nil, err
	}

	extension := filepath.Ext(filePath)
	fileName := strings.TrimSuffix(filepath.Base(filePath), extension)
	directory := filepath.Dir(filePath)

	sslFileName := fmt.Sprintf("%s-ssl%s", fileName, extension)
	sslFilePath := filepath.Join(directory, sslFileName)

	if _, err := os.Stat(sslFilePath); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(sslFilePath)

		if err != nil {
			return nil, err
		}

		_, err = file.Write([]byte(content))

		if err != nil {
			return nil, err
		}

		err = d.webServer.Config.ParseFile(sslFilePath)

		if err != nil {
			return nil, err
		}

		configFile := d.webServer.Config.GetConfigFile(sslFileName)
		vHostBlocks := configFile.FindVirtualHostBlocksByServerName(vhost.ServerName)

		if len(vHostBlocks) == 0 {
			return nil, fmt.Errorf("apache ssl host %s not found", vhost.ServerName)
		}

		vHostBlock := vHostBlocks[0]

		var sslAddresses []goapacheconf.Address
		addresses := vHostBlock.GetAddresses()

		for _, address := range addresses {
			sslAddresses = append(sslAddresses, address.GetAddressWithNewPort("443"))
		}

		vHostBlock.SetAddresses(sslAddresses)

		return &vHostBlock, nil
	}

	return nil, fmt.Errorf("config file already exists %s", filePath)
}

func (d *ApacheCertificateDeployer) createOrUpdateSingleDirective(block *goapacheconf.VirtualHostBlock, name goapacheconf.DirectiveName, value string) {
	directives := block.FindDirectives(name)

	if len(directives) > 1 {
		block.DeleteDirectiveByName(string(name))
		directives = nil
	}

	if len(directives) == 0 {
		block.AddDirective(string(name), []string{value}, false, true)
	} else {
		directive := directives[0]
		directive.SetValue(value)
	}
}

func (d *ApacheCertificateDeployer) removeDangerousForSslRewriteRules(vHostBlock *goapacheconf.VirtualHostBlock) {
	directives := vHostBlock.FindRewriteRuleDirectives()

	for _, directive := range directives {
		if !d.isRewriteRuleDangerousForSsl(directive) {
			continue
		}

		rcDirectives := directive.GetRelatedRewiteCondDirectives()
		vHostBlock.DeleteDirective(directive.Directive)

		for _, rcDirective := range rcDirectives {
			vHostBlock.DeleteDirective(rcDirective)
		}
	}
}

// isRewriteRuleDangerousForSsl checks if provided rewrite rule potentially can not be used for the virtual host with ssl
// e.g:
// RewriteRule ^ https://%{SERVER_NAME}%{REQUEST_URI} [L,QSA,R=permanent]
// Copying the above line to the ssl vhost would cause a
// redirection loop.
func (d *ApacheCertificateDeployer) isRewriteRuleDangerousForSsl(directive goapacheconf.RewriteRuleDirective) bool {
	values := directive.GetValues()

	// According to: https://httpd.apache.org/docs/2.4/rewrite/flags.html
	// The syntax of a RewriteRule is:
	// RewriteRule pattern target [Flag1,Flag2,Flag3]
	// i.e. target is required, so it must exist.

	if len(values) < 2 {
		return false
	}

	target := strings.TrimSpace(values[1])
	target = strings.Trim(target, "'\"")

	return strings.HasPrefix(target, "https://")
}

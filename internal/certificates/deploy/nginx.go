package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	nginxConfig "github.com/r2dtools/gonginxconf/config"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

type NginxCertificateDeployer struct {
	logger    logger.Logger
	webServer *webserver.NginxWebServer
	reverter  reverter.Reverter
}

func (d *NginxCertificateDeployer) DeployCertificate(vhost *dto.VirtualHost, certPath, certKeyPath string) (string, string, error) {
	wConfig := d.webServer.Config
	serverBlocks := wConfig.FindServerBlocksByServerName(vhost.ServerName)

	if len(serverBlocks) == 0 {
		return "", "", fmt.Errorf("nginx host %s does not exixst", vhost.ServerName)
	}

	var sslServerBlock *nginxConfig.ServerBlock
	var err error
	serverBlock := serverBlocks[0]

	for _, serverBlock := range serverBlocks {
		if serverBlock.HasSSL() {
			sslServerBlock = &serverBlock
		}
	}

	if sslServerBlock == nil {
		sslServerBlock, err = d.createSslHost(vhost, serverBlock)

		if err != nil {
			return "", "", err
		}

		d.reverter.AddConfigToDeletion(sslServerBlock.FilePath)
	} else {
		d.reverter.BackupConfig(sslServerBlock.FilePath)
	}

	d.createOrUpdateSingleDirective(sslServerBlock, webserver.NginxCertKeyDirective, certKeyPath)
	d.createOrUpdateSingleDirective(sslServerBlock, webserver.NginxCertDirective, certPath)

	sslServerBlockFileName := filepath.Base(sslServerBlock.FilePath)
	configFile := wConfig.GetConfigFile(sslServerBlockFileName)

	if err = configFile.Dump(); err != nil {
		return "", "", err
	}

	return sslServerBlock.FilePath, serverBlock.FilePath, nil
}

func (d *NginxCertificateDeployer) createSslHost(
	vhost *dto.VirtualHost,
	serverBlock nginxConfig.ServerBlock,
) (*nginxConfig.ServerBlock, error) {
	content := serverBlock.Dump()

	filePath, err := filepath.EvalSymlinks(serverBlock.FilePath)

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
		serverBlocks := configFile.FindServerBlocksByServerName(vhost.ServerName)

		if len(serverBlocks) == 0 {
			return nil, fmt.Errorf("nginx ssl host %s not found", vhost.ServerName)
		}

		serverBlock := serverBlocks[0]
		isIpv4Enabled := serverBlock.IsIpv4Enabled()
		isIpv6Enabled := serverBlock.IsIpv6Enabled()
		serverBlock.DeleteDirectiveByName("listen")

		if isIpv6Enabled {
			listenDirective := nginxConfig.NewDirective("listen", []string{"[::]:443", "ssl"})
			serverBlock.AddDirective(listenDirective, true, true)
		}

		if isIpv4Enabled {
			listenDirective := nginxConfig.NewDirective("listen", []string{"443", "ssl"})
			serverBlock.AddDirective(listenDirective, true, false)
		}

		return &serverBlock, nil
	}

	return nil, fmt.Errorf("config file already exists %s", filePath)
}

func (d *NginxCertificateDeployer) createOrUpdateSingleDirective(block *nginxConfig.ServerBlock, name, value string) {
	directives := block.FindDirectives(name)

	if len(directives) > 1 {
		block.DeleteDirectiveByName(name)
		directives = nil
	}

	if len(directives) == 0 {
		directive := nginxConfig.NewDirective(name, []string{value})
		block.AddDirective(directive, false, true)
	} else {
		directive := directives[0]
		directive.SetValue(value)
	}
}

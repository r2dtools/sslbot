package commondir

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/r2dtools/gonginxconf/config"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

const (
	acmeLocation = "/.well-known/acme-challenge/"
)

type NginxCommonDirQuery struct {
	webServer *webserver.NginxWebServer
}

func (q *NginxCommonDirQuery) GetCommonDirStatus(serverName string) CommonDir {
	var commonDir CommonDir
	serverBlock := findServerBlock(q.webServer.Config, serverName)

	if serverBlock == nil {
		return commonDir
	}

	commonDirBlock := findCommonDirBlock(serverBlock)

	if commonDirBlock == nil {
		return commonDir
	}

	directives := commonDirBlock.FindDirectives("root")

	if len(directives) == 0 {
		return commonDir
	}

	commonDir.Enabled = true
	commonDir.Root = strings.Trim(directives[0].GetFirstValue(), " \"")

	return commonDir
}

type NginxCommonDirChangeCommand struct {
	webServer *webserver.NginxWebServer
	reverter  reverter.Reverter
	logger    logger.Logger
	commonDir string
	mx        *sync.Mutex
}

func (c *NginxCommonDirChangeCommand) EnableCommonDir(serverName string) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	wConfig := c.webServer.Config
	serverBlock := findServerBlock(wConfig, serverName)

	if serverBlock == nil {
		return fmt.Errorf("nginx host %s on 80 or 443 port does not exist", serverName)
	}

	serverBlockFileName := filepath.Base(serverBlock.FilePath)
	configFile := wConfig.GetConfigFile(serverBlockFileName)

	if configFile == nil {
		return fmt.Errorf("failed to find config file for host %s", serverName)
	}

	processManager, err := c.webServer.GetProcessManager()

	if err != nil {
		return err
	}

	if findCommonDirBlock(serverBlock) != nil {
		c.logger.Info("common directory is already enabled for %s host", serverName)

		return nil
	}

	commonDirLocationBlock := serverBlock.AddLocationBlock("^~", acmeLocation, true)
	commonDirLocationBlock.AddDirective(config.NewDirective("root", []string{c.commonDir}), true, false)
	commonDirLocationBlock.AddDirective(config.NewDirective("default_type", []string{`"text/plain"`}), true, false)

	if err := c.reverter.BackupConfig(serverBlock.FilePath); err != nil {
		return err
	}

	err = configFile.Dump()

	if err != nil {
		if rErr := c.reverter.Rollback(); rErr != nil {
			c.logger.Error(fmt.Sprintf("failed to rollback webserver configuration on common directory switching: %v", rErr))

		}

		return err
	}

	if err := processManager.Reload(); err != nil {
		if rErr := c.reverter.Rollback(); rErr != nil {
			c.logger.Error(fmt.Sprintf("failed to rollback webserver configuration on webserver reload: %v", rErr))
		}

		return err
	}

	c.reverter.Commit()

	return nil
}

func (c *NginxCommonDirChangeCommand) DisableCommonDir(serverName string) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	wConfig := c.webServer.Config
	serverBlock := findServerBlock(wConfig, serverName)

	if serverBlock == nil {
		return fmt.Errorf("nginx host %s on 80 or 443 port does not exist", serverName)
	}

	serverBlockFileName := filepath.Base(serverBlock.FilePath)
	configFile := wConfig.GetConfigFile(serverBlockFileName)

	if configFile == nil {
		return fmt.Errorf("failed to find config file for host %s", serverName)
	}

	processManager, err := c.webServer.GetProcessManager()

	if err != nil {
		return err
	}

	commonDirBlock := findCommonDirBlock(serverBlock)

	if commonDirBlock == nil {
		return nil
	}

	serverBlock.DeleteLocationBlock(*commonDirBlock)

	if err := c.reverter.BackupConfig(serverBlock.FilePath); err != nil {
		return err
	}

	err = configFile.Dump()

	if err != nil {
		if rErr := c.reverter.Rollback(); rErr != nil {
			c.logger.Error(fmt.Sprintf("failed to rollback webserver configuration on common directory switching: %v", rErr))

		}

		return err
	}

	if err := processManager.Reload(); err != nil {
		if rErr := c.reverter.Rollback(); rErr != nil {
			c.logger.Error(fmt.Sprintf("failed to rollback webserver configuration on webserver reload: %v", rErr))
		}

		return err
	}

	c.reverter.Commit()

	return nil
}

func findCommonDirBlock(serverBlock *config.ServerBlock) *config.LocationBlock {
	locationBlocks := serverBlock.FindLocationBlocks()

	for _, locationBlock := range locationBlocks {
		if locationBlock.GetLocationMatch() == acmeLocation {
			return &locationBlock
		}
	}

	return nil
}

func findServerBlock(nginxConfig *config.Config, serverName string) *config.ServerBlock {
	serverBlocks := nginxConfig.FindServerBlocksByServerName(serverName)

	if len(serverBlocks) == 0 {
		return nil
	}

	var nonSslServerBlock *config.ServerBlock

	for _, serverBlock := range serverBlocks {
		for _, address := range serverBlock.GetAddresses() {
			if address.Port == "80" && nonSslServerBlock == nil {
				nonSslServerBlock = &serverBlock

				continue
			}

			if address.Port == "443" {
				return &serverBlock
			}
		}
	}

	return nonSslServerBlock
}

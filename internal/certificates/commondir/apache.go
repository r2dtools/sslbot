package commondir

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/r2dtools/goapacheconf"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

const (
	apacheAcmeLocation      = "/.well-known/acme-challenge"
	apacheAcmeLocationMatch = "^/.well-known/acme-challenge"
)

type ApacheCommonDirQuery struct {
	webServer *webserver.ApacheWebServer
}

func (q *ApacheCommonDirQuery) GetCommonDirStatus(serverName string) CommonDir {
	var commonDir CommonDir
	vHostBlock := findVirtualHostBlock(q.webServer.Config, serverName)

	if vHostBlock == nil {
		return commonDir
	}

	commonDirAlias := findApacheCommonDirAlias(vHostBlock)

	if commonDirAlias == nil {
		return commonDir
	}

	commonDir.Enabled = true
	toLocation := strings.Trim(commonDirAlias.GetToLocation(), " \"")
	commonDir.Root = strings.TrimSuffix(toLocation, apacheAcmeLocation)

	return commonDir
}

func findApacheCommonDirAlias(vHostBlock *goapacheconf.VirtualHostBlock) *goapacheconf.AliasDirective {
	aliasDirectives := vHostBlock.FindAlliasDirectives()

	for _, aliasDirective := range aliasDirectives {
		if strings.HasPrefix(aliasDirective.GetFromLocation(), apacheAcmeLocation) {
			return &aliasDirective
		}
	}

	return nil
}

type ApacheCommonDirChangeCommand struct {
	webServer *webserver.ApacheWebServer
	reverter  reverter.Reverter
	logger    logger.Logger
	commonDir string
	mx        *sync.Mutex
}

func (c *ApacheCommonDirChangeCommand) EnableCommonDir(serverName string) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	wConfig := c.webServer.Config
	virtualHostBlock := findVirtualHostBlock(wConfig, serverName)

	if virtualHostBlock == nil {
		return fmt.Errorf("apache host %s on 80 or 443 port does not exist", serverName)
	}

	virtualHostBlockFileName := filepath.Base(virtualHostBlock.FilePath)
	configFile := wConfig.GetConfigFile(virtualHostBlockFileName)

	if configFile == nil {
		return fmt.Errorf("failed to find config file for host %s", serverName)
	}

	processManager, err := c.webServer.GetProcessManager()

	if err != nil {
		return err
	}

	if findApacheCommonDirAlias(virtualHostBlock) != nil {
		c.logger.Info("common directory is already enabled for %s host", serverName)

		return nil
	}

	aliasTo := filepath.Join(c.commonDir, apacheAcmeLocation)
	aliasTo = filepath.Clean(aliasTo)

	aliasDirective := goapacheconf.NewAliasDirective(apacheAcmeLocation, aliasTo)
	aliasDirective.AppendNewLine()
	aliasDirective = virtualHostBlock.AppendAliasDirective(aliasDirective)
	commonDirLocation := findAcmeCommonDirLocationBlock(virtualHostBlock)

	if commonDirLocation == nil {
		location := virtualHostBlock.AddLocationBlock(fmt.Sprintf("%s/", apacheAcmeLocation), false)
		directive := goapacheconf.NewDirective(goapacheconf.Order, []string{"Allow,Deny"})
		location.AppendDirective(directive)

		directive = goapacheconf.NewDirective(goapacheconf.Allow, []string{"from", "all"})
		location.AppendDirective(directive)

		directive = goapacheconf.NewDirective(goapacheconf.Satisfy, []string{"any"})
		location.AppendDirective(directive)

		commonDirLocation = &location
	}

	if err := c.reverter.BackupConfig(virtualHostBlock.FilePath); err != nil {
		return err
	}

	_, err = configFile.Dump()

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

func (c *ApacheCommonDirChangeCommand) DisableCommonDir(serverName string) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	wConfig := c.webServer.Config
	virtualHostBlock := findVirtualHostBlock(wConfig, serverName)

	if virtualHostBlock == nil {
		return fmt.Errorf("apache host %s on 80 or 443 port does not exist", serverName)
	}

	virtualHostBlockFileName := filepath.Base(virtualHostBlock.FilePath)
	configFile := wConfig.GetConfigFile(virtualHostBlockFileName)

	if configFile == nil {
		return fmt.Errorf("failed to find config file for host %s", serverName)
	}

	processManager, err := c.webServer.GetProcessManager()

	if err != nil {
		return err
	}

	commonDirAlias := findApacheCommonDirAlias(virtualHostBlock)

	if commonDirAlias == nil {
		return nil
	}

	virtualHostBlock.DeleteAliasDirective(*commonDirAlias)

	commonDirLocation := findAcmeCommonDirLocationBlock(virtualHostBlock)

	if commonDirLocation != nil {
		virtualHostBlock.DeleteLocationBlock(*commonDirLocation)
	}

	commonDirLocationMatch := findAcmeCommonDirLocationMatchBlock(virtualHostBlock)

	if commonDirLocationMatch != nil {
		virtualHostBlock.DeleteLocationMatchBlock(*commonDirLocationMatch)
	}

	if err := c.reverter.BackupConfig(virtualHostBlock.FilePath); err != nil {
		return err
	}

	_, err = configFile.Dump()

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

func findAcmeCommonDirLocationBlock(vHostBlock *goapacheconf.VirtualHostBlock) *goapacheconf.LocationBlock {
	locationBlocks := vHostBlock.FindLocationBlocks()

	for _, locationBlock := range locationBlocks {
		location := strings.Trim(locationBlock.GetLocation(), "\"")

		if strings.HasPrefix(location, apacheAcmeLocation) {
			return &locationBlock
		}
	}

	return nil
}

func findAcmeCommonDirLocationMatchBlock(vHostBlock *goapacheconf.VirtualHostBlock) *goapacheconf.LocationMatchBlock {
	locationMatchBlocks := vHostBlock.FindLocationMatchBlocks()

	for _, locationMatchBlock := range locationMatchBlocks {
		match := strings.Trim(locationMatchBlock.GetLocationMatch(), "\"")

		if strings.HasPrefix(match, apacheAcmeLocationMatch) {
			return &locationMatchBlock
		}
	}

	return nil
}

func findVirtualHostBlock(apacheConfig *goapacheconf.Config, serverName string) *goapacheconf.VirtualHostBlock {
	vHostBlocks := apacheConfig.FindVirtualHostBlocksByServerName(serverName)

	if len(vHostBlocks) == 0 {
		return nil
	}

	var nonSslVirtualHostBlock *goapacheconf.VirtualHostBlock

	for _, vHostBlock := range vHostBlocks {
		for _, address := range vHostBlock.GetAddresses() {
			if address.Port == "80" && nonSslVirtualHostBlock == nil {
				nonSslVirtualHostBlock = &vHostBlock

				continue
			}

			if address.Port == "443" {
				return &vHostBlock
			}
		}
	}

	return nonSslVirtualHostBlock
}

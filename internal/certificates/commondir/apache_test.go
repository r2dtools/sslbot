//go:build apache

package commondir

import (
	"sync"
	"testing"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApacheCommonDir(t *testing.T) {
	host := "example2.com"

	command, query, apacheWebServer, rv := getApacheCommonDir(t)
	defer rv.Rollback()

	commonDir := query.GetCommonDirStatus(host)
	require.True(t, commonDir.Enabled)
	require.Equal(t, "/var/www/vhosts/default/htdocs", commonDir.Root)

	err := command.DisableCommonDir(host)
	assert.Nil(t, err)
	commonDir = query.GetCommonDirStatus(host)
	assert.False(t, commonDir.Enabled)
	assert.Empty(t, commonDir.Root)

	blocks := apacheWebServer.Config.FindVirtualHostBlocksByServerName(host)
	assert.Len(t, blocks, 1)

	block := blocks[0]
	aliasDirectives := block.FindAlliasDirectives()
	assert.Empty(t, aliasDirectives)

	locationBlocks := block.FindLocationBlocks()
	locationMatchBlocks := block.FindLocationMatchBlocks()

	require.Empty(t, locationBlocks)
	require.Empty(t, locationMatchBlocks)

	err = command.EnableCommonDir(host)
	require.Nil(t, err)

	configFile := apacheWebServer.Config.GetConfigFile("example2.com.conf")
	require.NotNil(t, configFile)

	content, err := configFile.Dump()

	expectedContent := `#ATTENTION!
#
#DO NOT MODIFY THIS FILE BECAUSE IT WAS GENERATED AUTOMATICALLY,
#SO ALL YOUR CHANGES WILL BE LOST THE NEXT TIME THE FILE IS GENERATED.
#IF YOU REQUIRE TO APPLY CUSTOM MODIFICATIONS, PERFORM THEM IN THE FOLLOWING FILES:
<VirtualHost 127.0.0.1:80 >
	ServerName "example2.com"
	ServerAlias "www.example2.com"
	ServerAlias "ipv4.example2.com"
	UseCanonicalName Off

	DocumentRoot "/var/www/html"

	<IfModule mod_suexec.c >
		SuexecUserGroup "example.com_bc8fzauq68v" "psacln"
	</IfModule>

	<IfModule mod_sysenv.c >
		SetSysEnv PP_VHOST_ID "08fc3720-4f9b-4648-b1b2-4362a6319f4f"
	</IfModule>

	<Directory /var/www/html >
		# test inline comment
		Options -Includes -ExecCGI
	</Directory>

	DirectoryIndex "index.html" "index.cgi" "index.pl" "index.php" "index.xhtml" "index.htm" "index.shtml"

	<Directory /var/www/html >
		Options -FollowSymLinks
		AllowOverride AuthConfig FileInfo Indexes Limit Options=Indexes,SymLinksIfOwnerMatch,MultiViews,ExecCGI,Includes,IncludesNOEXEC
	</Directory>

	#extension letsencrypt begin
	#extension letsencrypt end
	Alias /.well-known/acme-challenge /var/www/html/.well-known/acme-challenge
	<Location /.well-known/acme-challenge/ >
		Order Allow,Deny
		Allow from all
		Satisfy any
	</Location>

</VirtualHost>
`
	require.Equal(t, expectedContent, content)
}

func getApacheCommonDir(t *testing.T) (CommonDirChangeCommand, CommonDirQuery, webserver.ApacheWebServer, reverter.Reverter) {
	config, err := config.GetConfig()
	require.Nil(t, err)
	options := config.ToMap()

	apacheWebServer, err := webserver.GetApacheWebServer(options)
	require.Nil(t, err)

	rv, err := reverter.CreateReverter(apacheWebServer, &logger.TestLogger{T: t})
	require.Nil(t, err)

	command, err := CreateCommonDirChangeCommand(config, apacheWebServer, rv, &logger.TestLogger{T: t}, &sync.Mutex{})
	assert.Nil(t, err)

	query, err := CreateCommonDirStatusQuery(apacheWebServer)
	require.Nil(t, err)

	return command, query, *apacheWebServer, rv
}

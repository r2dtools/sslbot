package webserver

import (
	"regexp"
	"slices"

	"github.com/google/go-cmp/cmp"
	"github.com/unknwon/com"

	"github.com/r2dtools/sslbot/internal/dto"
)

func filterVhosts(vhosts []dto.VirtualHost) []dto.VirtualHost {
	var fVhosts []dto.VirtualHost

	for _, vhost := range vhosts {
		if !isValidDomain(vhost.ServerName) || !checkVhostPorts(vhost.Addresses, []string{"80", "443"}) {
			continue
		}

		fVhosts = append(fVhosts, vhost)
	}

	return fVhosts
}

// MergeVhosts merge similar vhosts. For example, vhost:443 will be merged with vhost:80
func mergeVhosts(vhosts []dto.VirtualHost) []dto.VirtualHost {
	var fVhosts []dto.VirtualHost
	vhostsMap := make(map[string]dto.VirtualHost)

	for _, vhost := range vhosts {
		if existedVhost, ok := vhostsMap[vhost.ServerName]; ok {
			existedVhost.Ssl = existedVhost.Ssl || vhost.Ssl

			if existedVhost.DocRoot == "" {
				existedVhost.DocRoot = vhost.DocRoot
			}

			// merge addresses (for example ipv4 + ipv6)
			for _, address := range vhost.Addresses {
				var addressExists bool
				for _, eAddress := range existedVhost.Addresses {
					if cmp.Equal(address, eAddress) {
						addressExists = true
						break
					}
				}

				if !addressExists {
					existedVhost.Addresses = append(existedVhost.Addresses, address)
				}
			}

			// merge aliases
			for _, alias := range vhost.Aliases {
				existedVhost.Aliases = com.AppendStr(existedVhost.Aliases, alias)
			}

			vhostsMap[vhost.ServerName] = existedVhost
		} else {
			vhostsMap[vhost.ServerName] = vhost
		}
	}

	for _, vhost := range vhostsMap {
		fVhosts = append(fVhosts, vhost)
	}

	return fVhosts
}

func checkVhostPorts(addresses []dto.VirtualHostAddress, ports []string) bool {
	for _, address := range addresses {
		if slices.Contains(ports, address.Port) {
			return true
		}
	}

	return false
}

func isValidDomain(domain string) bool {
	// Regular expression to validate domain name
	regex := `^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(regex, domain)

	return match
}

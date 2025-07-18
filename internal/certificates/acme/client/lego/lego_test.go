package lego

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOutputError(t *testing.T) {
	type testData struct {
		input, output string
	}
	items := []testData{
		{
			`20:33:11 2021/01/03 20:32:54 [INFO] [example2.com] acme: Obtaining bundled SAN certificate
			2021/01/03 20:32:54 [INFO] [example2.com] AuthURL: http://localhost:4001/acme/authz-v3/6
			2021/01/03 20:32:54 [INFO] [example2.com] acme: Could not find solver for: tls-alpn-01
			2021/01/03 20:32:54 [INFO] [example2.com] acme: use http-01 solver
			2021/01/03 20:32:54 [INFO] [example2.com] acme: Trying to solve HTTP-01
			2021/01/03 20:33:11 [INFO] Deactivating auth: http://localhost:4001/acme/authz-v3/6
			2021/01/03 20:33:11 [INFO] Unable to deactivate the authorization: http://localhost:4001/acme/authz-v3/6
			2021/01/03 20:33:11 Could not obtain certificates:
				error: one or more domains had a problem:
			[example2.com] acme: error: 400 :: urn:ietf:params:acme:error:connection :: Fetching http://example2.com/.well-known/acme-challenge/p6s2hgfl5reHgZ4feROA4dxsOzH6EaR1701xEQriV94: Timeout during connect (likely firewall problem), url: 
			
			`,
			`one or more domains had a problem:
[example2.com] acme: 400 :: urn:ietf:params:acme:error:connection :: Fetching http://example2.com/.well-known/acme-challenge/p6s2hgfl5reHgZ4feROA4dxsOzH6EaR1701xEQriV94: Timeout during connect (likely firewall problem)`,
		},
	}

	for _, item := range items {
		output := getOutputError(item.input)
		assert.Equal(t, item.output, output)
	}
}

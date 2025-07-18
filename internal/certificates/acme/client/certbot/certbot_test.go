package certbot

import (
	"strings"
	"testing"

	"github.com/r2dtools/sslbot/internal/certificates/request"
	"github.com/stretchr/testify/assert"
)

func TestBuildCmdParams(t *testing.T) {
	challengeType := HTTPChallengeType{WebRoot: "path"}
	request := request.IssueRequest{
		Email:         "test@email.com",
		ServerName:    "example.com",
		WebServer:     "nginx",
		ChallengeType: "http",
		Subjects:      []string{"www.example.com"},
	}

	params := buildCmdParams(request, challengeType)
	cmd := strings.Join(params, " ")

	assert.Equal(t, "certonly -w path -d example.com -d www.example.com -m test@email.com -n --agree-tos", cmd)

	request.Assign = true
	params = buildCmdParams(request, challengeType)
	cmd = strings.Join(params, " ")
	assert.Equal(t, "run -a webroot -i nginx -w path -d example.com -d www.example.com -m test@email.com -n --agree-tos", cmd)
}

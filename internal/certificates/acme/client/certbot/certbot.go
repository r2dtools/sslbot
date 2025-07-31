package certbot

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates/acme"
	"github.com/r2dtools/sslbot/internal/certificates/request"
	"github.com/r2dtools/sslbot/internal/logger"
)

type CertBot struct {
	bin     string
	storage *CertBotStorage
}

func (b *CertBot) Issue(docRoot string, request request.IssueRequest) (string, string, error) {
	var challengeType acme.ChallengeType
	serverName := request.ServerName

	switch request.ChallengeType {
	case acme.HttpChallengeTypeCode:
		challengeType = HTTPChallengeType{WebRoot: docRoot}
	default:
		return "", "", fmt.Errorf("unsupported challenge type: %s", request.ChallengeType)
	}

	params := buildCmdParams(request, challengeType)
	cmd := exec.Command(b.bin, params...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) == 0 {
			return "", "", err
		}

		return "", "", fmt.Errorf("%s\n%s", output, err.Error())
	}

	return b.storage.GetCertificatePath(serverName)
}

func buildCmdParams(request request.IssueRequest, challengeType acme.ChallengeType) []string {
	serverName := request.ServerName
	params := []string{}

	if request.Assign {
		params = append(params, "run", "-a webroot", "-i "+request.WebServer)
	} else {
		params = append(params, "certonly")
	}

	params = append(params, challengeType.GetParams()...)
	params = append(params, "-d "+serverName)

	for _, subject := range request.Subjects {
		if subject != serverName {
			params = append(params, "-d "+subject)
		}
	}

	params = append(params, "-m "+request.Email, "-n", "--agree-tos")

	return params
}

func CreateCertBot(config *config.Config, logger logger.Logger) (*CertBot, error) {
	storage, err := CreateCertStorage(config, logger)

	if err != nil {
		return nil, err
	}

	return &CertBot{bin: config.CertBotBin, storage: storage}, nil
}

func GetVersion(config *config.Config) (string, error) {
	cmd := exec.Command(config.CertBotBin, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) == 0 {
			return "", err
		}

		return "", errors.New(string(output))
	}

	parts := strings.Split(string(output), " ")

	if len(parts) != 2 {
		return "", errors.New("failed to detect certbot version")
	}

	version := strings.TrimSpace(parts[1])

	return version, nil
}

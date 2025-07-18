package lego

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates/acme"
	"github.com/r2dtools/sslbot/internal/certificates/request"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/unknwon/com"
)

const (
	httpPort = 80
	tlsPort  = 443
)

type Lego struct {
	bin      string
	caServer string
	dataDir  string
	storage  *LegoStorage
}

func (l *Lego) Issue(docRoot string, request request.IssueRequest) (string, string, error) {
	var challengeType acme.ChallengeType
	serverName := request.ServerName

	switch request.ChallengeType {
	case acme.HttpChallengeTypeCode:
		challengeType = &HTTPChallengeType{
			HTTPPort: httpPort,
			TLSPort:  tlsPort,
			WebRoot:  docRoot,
		}
	default:
		return "", "", fmt.Errorf("unsupported challenge type: %s", request.ChallengeType)
	}

	params := []string{"--email=" + request.Email, "--domains=" + serverName}

	for _, subject := range request.Subjects {
		if subject != serverName {
			params = append(params, "--domains="+subject)
		}
	}

	params = append(params, challengeType.GetParams()...)
	err := l.execCmd("run", params)

	if err != nil {
		return "", "", err
	}

	return l.storage.GetCertificatePath(serverName)
}

func (l *Lego) execCmd(command string, params []string) error {
	aParams := []string{"--server=" + l.caServer, "--accept-tos", "--path=" + l.dataDir, "--pem"}
	params = append(params, aParams...)
	params = append(params, command)
	cmd := exec.Command(l.bin, params...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) == 0 {
			return err
		}

		return errors.New(getOutputError(string(output)))
	}

	return nil
}

func getOutputError(output string) string {
	errIndex := strings.Index(output, "error: ")

	if errIndex != -1 {
		output = output[errIndex:]
	}

	output = strings.ReplaceAll(output, "error: ", "")
	parts := strings.Split(output, "\n")
	var errorParts []string

	for _, part := range parts {
		if strings.Contains(part, "[INFO]") || strings.Contains(part, "[WARN]") {
			continue
		}

		// Skip log time: xxxx/xx/xx xx:xx:xx
		part = removeRegexString(part, `^[0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} (.*)`)

		if part == "" {
			continue
		}

		errorParts = append(errorParts, part)
	}

	output = strings.Join(errorParts, "\n")

	// Skip ", url:" string. Seems it is a bug in lego library
	// https://github.com/go-acme/lego/blob/master/acme/errors.go#L47
	return removeRegexString(output, `(?s)(.*), url:$`)
}

func removeRegexString(str string, regex string) string {
	re, err := regexp.Compile(regex)

	if err == nil {
		rParts := re.FindStringSubmatch(str)

		if len(rParts) > 1 {
			str = rParts[1]
		}
	}

	return strings.TrimSpace(str)
}

func CreateClient(config *config.Config, logger logger.Logger) (*Lego, error) {
	dataDir := config.GetPathInsideVarDir("lego")

	if !com.IsExist(dataDir) {
		err := os.MkdirAll(dataDir, 0755)

		if err != nil {
			return nil, err
		}
	}

	storage, err := CreateCertStorage(config, logger)

	if err != nil {
		return nil, err
	}

	client := &Lego{
		bin:      config.LegoBin,
		caServer: config.CaServer,
		dataDir:  dataDir,
		storage:  storage,
	}

	return client, nil
}

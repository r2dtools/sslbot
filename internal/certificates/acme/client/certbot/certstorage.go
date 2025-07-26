package certbot

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/utils"
	"github.com/unknwon/com"
)

var ErrStorageDirNotExists = errors.New("storage directory not exists")

type CertBotStorage struct {
	*sync.RWMutex
	bin    string
	path   string
	logger logger.Logger
}

func (s *CertBotStorage) RemoveCertificate(certName string) error {
	s.Lock()
	defer s.Unlock()

	params := []string{"delete", "--cert-name " + certName}
	cmd := exec.Command(s.bin, params...)
	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to delete certificate %s: %v", certName, err)
	}

	return nil
}

func (s *CertBotStorage) GetCertificate(certName string) (*dto.Certificate, error) {
	s.RLock()
	defer s.RUnlock()

	certPath, _, err := s.GetCertificatePath(certName)

	if err != nil {
		return nil, err
	}

	return utils.GetCertificateFromFile(certPath)
}

func (s *CertBotStorage) GetCertificateAsString(certName string) (certPath string, certContent string, err error) {
	s.RLock()
	defer s.RUnlock()

	certPath = s.getCertificatePath(certName)
	certContentBytes, err := os.ReadFile(certPath)

	if err != nil {
		return "", "", fmt.Errorf("could not read certificate content: %v", err)
	}

	certContent = string(certContentBytes)

	return
}

func (s *CertBotStorage) GetCertificates() (map[string]*dto.Certificate, error) {
	s.RLock()
	defer s.RUnlock()

	certNameMap, err := s.getStorageCertNameMap()

	if err != nil {
		return nil, err
	}

	certsMap := map[string]*dto.Certificate{}

	for certName := range certNameMap {
		certPath := s.getCertificatePath(certName)
		cert, err := utils.GetCertificateFromFile(certPath)

		if err != nil {
			s.logger.Error("failed to parse certificate %s: %v", certName, err)

			continue
		}

		certsMap[certName] = cert
	}

	return certsMap, nil
}

func (s *CertBotStorage) GetCertificatePath(certName string) (certPath string, keyPath string, err error) {
	certPath = s.getCertificatePath(certName)
	keyPath = s.getPrivateKeyPath(certName)

	return
}

func (s *CertBotStorage) getStorageCertNameMap() (map[string]struct{}, error) {
	certNameMap := make(map[string]struct{})
	entries, err := os.ReadDir(s.path)

	if err != nil {
		return nil, fmt.Errorf("could not get certificate list in the storage: %v", err)
	}

	for _, entry := range entries {
		// directories like example.com/
		if !entry.IsDir() {
			continue
		}

		certNameMap[entry.Name()] = struct{}{}
	}

	return certNameMap, nil
}

func (s *CertBotStorage) getCertificatePath(certName string) string {
	return filepath.Join(s.path, certName, "fullchain.pem")
}

func (s *CertBotStorage) getPrivateKeyPath(certName string) string {
	return filepath.Join(s.path, certName, "privkey.pem")
}

func CreateCertStorage(config *config.Config, logger logger.Logger) (*CertBotStorage, error) {
	workDir := config.CertBotWokrDir

	if !com.IsExist(workDir) {
		return nil, ErrStorageDirNotExists
	}

	return &CertBotStorage{
		RWMutex: &sync.RWMutex{},
		path:    workDir,
		bin:     config.CertBotBin,
		logger:  logger,
	}, nil
}

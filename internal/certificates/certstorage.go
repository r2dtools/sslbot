package certificates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/utils"
	"github.com/unknwon/com"
)

type DefaultStorage struct {
	*sync.RWMutex
	path   string
	logger logger.Logger
}

func (s *DefaultStorage) AddPemCertificate(certName, pemData string) (certPath string, err error) {
	s.Lock()
	defer s.Unlock()

	certPath = s.getCertificatePath(certName)

	if err := os.WriteFile(certPath, []byte(pemData), 0644); err != nil {
		return "", fmt.Errorf("could not save certificate data to the storage: %v", err)
	}

	return certPath, nil
}

func (s *DefaultStorage) RemoveCertificate(certName string) error {
	s.Lock()
	defer s.Unlock()

	certPath := s.getCertificatePath(certName)

	if com.IsFile(certPath) {
		if err := os.Remove(certPath); err != nil {
			return fmt.Errorf("could not remove certificate %s: %v", certName, err)
		}
	}

	return nil
}

func (s *DefaultStorage) GetCertificate(certName string) (*dto.Certificate, error) {
	s.RLock()
	defer s.RUnlock()

	certPath := s.getCertificatePath(certName)

	return utils.GetCertificateFromFile(certPath)
}

func (s *DefaultStorage) GetCertificateAsString(certName string) (certPath string, certContent string, err error) {
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

func (s *DefaultStorage) GetCertificates() (map[string]*dto.Certificate, error) {
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

func (s *DefaultStorage) GetCertificatePath(certName string) (certPath string, keyPath string, err error) {
	certPath = s.getCertificatePath(certName)
	keyPath = certPath

	return
}

func (s *DefaultStorage) getStorageCertNameMap() (map[string]struct{}, error) {
	certNameMap := make(map[string]struct{})
	entries, err := os.ReadDir(s.path)

	if err != nil {
		return nil, fmt.Errorf("could not get certificate list in the storage: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		certExt := filepath.Ext(name)

		if strings.Trim(certExt, ".") != "pem" {
			continue
		}

		certNameMap[name[:(len(name)-len(certExt))]] = struct{}{}
	}

	return certNameMap, nil
}

func (s *DefaultStorage) getCertificatePath(certName string) string {
	return filepath.Join(s.path, certName+".pem")
}

func CreateCertStorage(config *config.Config, logger logger.Logger) (*DefaultStorage, error) {
	path := config.GetPathInsideVarDir("default", "certificates")

	if !com.IsExist(path) {
		err := os.MkdirAll(path, 0755)

		if err != nil {
			return nil, err
		}
	}

	return &DefaultStorage{RWMutex: &sync.RWMutex{}, path: path, logger: logger}, nil
}

package lego

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

type LegoStorage struct {
	*sync.RWMutex
	path   string
	logger logger.Logger
}

func (s *LegoStorage) RemoveCertificate(certName string) error {
	s.Lock()
	defer s.Unlock()

	certPemPath := s.getCertificatePath(certName)
	certCrtPath := s.getFilePathByNameWithExt(certName, "crt")
	certIssuerCrtPath := s.getFilePathByNameWithExt(certName, "issuer.crt")
	certJsonData := s.getFilePathByNameWithExt(certName, "json")
	keyPath := s.getFilePathByNameWithExt(certName, "key")
	rPaths := []string{certPemPath, certCrtPath}
	nrPaths := []string{certIssuerCrtPath, keyPath, certJsonData}

	for _, path := range rPaths {
		if com.IsFile(path) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("could not remove certificate %s: %v", certName, err)
			}
		}
	}

	for _, path := range nrPaths {
		if com.IsFile(path) {
			os.Remove(path)
		}
	}

	return nil
}

func (s *LegoStorage) GetCertificate(certName string) (*dto.Certificate, error) {
	s.RLock()
	defer s.RUnlock()

	certPath, _, err := s.GetCertificatePath(certName)

	if err != nil {
		return nil, err
	}

	return utils.GetCertificateFromFile(certPath)
}

func (s *LegoStorage) GetCertificateAsString(certName string) (certPath string, certContent string, err error) {
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

func (s *LegoStorage) GetCertificates() (map[string]*dto.Certificate, error) {
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

func (s *LegoStorage) GetCertificatePath(certName string) (certPath string, keyPath string, err error) {
	certPath = s.getCertificatePath(certName)
	keyPath = certPath

	return
}

func (s *LegoStorage) getFilePathByNameWithExt(fileName, extension string) string {
	return filepath.Join(s.path, fileName+"."+extension)
}

func (s *LegoStorage) getCertificatePath(certName string) string {
	return s.getFilePathByNameWithExt(certName, "pem")
}

func (s *LegoStorage) getStorageCertNameMap() (map[string]struct{}, error) {
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

func CreateCertStorage(config *config.Config, logger logger.Logger) (*LegoStorage, error) {
	dataPath := config.GetPathInsideVarDir("lego", "certificates")

	if !com.IsExist(dataPath) {
		err := os.MkdirAll(dataPath, 0755)

		if err != nil {
			return nil, err
		}
	}

	return &LegoStorage{RWMutex: &sync.RWMutex{}, path: dataPath, logger: logger}, nil
}

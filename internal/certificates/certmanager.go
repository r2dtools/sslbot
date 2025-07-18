package certificates

import (
	"errors"
	"fmt"

	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates/acme/client"
	"github.com/r2dtools/sslbot/internal/certificates/acme/client/certbot"
	"github.com/r2dtools/sslbot/internal/certificates/acme/client/lego"
	"github.com/r2dtools/sslbot/internal/certificates/commondir"
	"github.com/r2dtools/sslbot/internal/certificates/request"
	"github.com/r2dtools/sslbot/internal/dto"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/utils"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

type CertStorage interface {
	RemoveCertificate(certName string) error
	GetCertificate(certName string) (*dto.Certificate, error)
	GetCertificateAsString(certName string) (certPath string, certContent string, err error)
	GetCertificates() (map[string]*dto.Certificate, error)
	GetCertificatePath(certName string) (certPath string, keyPath string, err error)
}

type CertStorageType string

const (
	Default CertStorageType = "default"
	CertBot CertStorageType = "certbot"
	Lego    CertStorageType = "lego"
)

type CertStorageItem struct {
	StorageType CertStorageType
	CertName    string
	Certificate *dto.Certificate
}

func (i CertStorageItem) Key() string {
	return fmt.Sprintf("%s__%s", i.StorageType, i.CertName)
}

type webServerFactory func(code string, options map[string]string) (webserver.WebServer, error)
type reverterFactory func(wServer webserver.WebServer, logger logger.Logger) (reverter.Reverter, error)

type CertificateManager struct {
	wServerFactory  webServerFactory
	reverterFactory reverterFactory
	certStorages    map[CertStorageType]CertStorage
	acmeClient      client.AcmeClient
	logger          logger.Logger
	config          *config.Config
}

func (c *CertificateManager) Issue(request request.IssueRequest) (*dto.Certificate, error) {
	serverName := request.ServerName
	wServer, err := c.wServerFactory(request.WebServer, c.config.ToMap())

	if err != nil {
		return nil, err
	}

	sReverter, err := c.reverterFactory(wServer, c.logger)

	if err != nil {
		return nil, err
	}

	certDeployer := createCertificateDeployer(c.config, wServer, sReverter, c.logger)
	commonDirQuery, err := commondir.CreateCommonDirStatusQuery(wServer)

	if err != nil {
		return nil, err
	}

	vhost, err := wServer.GetVhostByName(serverName)

	if err != nil {
		return nil, err
	}

	if vhost == nil {
		return nil, fmt.Errorf("host %s not found", serverName)
	}

	docRoot := vhost.DocRoot
	commonDir := commonDirQuery.GetCommonDirStatus(serverName)

	if commonDir.Enabled {
		docRoot = commonDir.Root
	}

	certPath, keyPath, err := c.acmeClient.Issue(docRoot, request)

	if err != nil {
		c.logger.Debug("%v", err)

		return nil, err
	}

	if request.Assign {
		certDeployer.DeployCertificate(serverName, certPath, keyPath, request.PreventReload)
	}

	return utils.GetCertificateFromFile(certPath)
}

func (c *CertificateManager) Assign(request request.AssignRequest) (*dto.Certificate, error) {
	storageType := CertStorageType(request.StorageType)
	storage, err := c.getStorage(CertStorageType(storageType))

	if err != nil {
		return nil, err
	}

	certPath, keyPath, err := storage.GetCertificatePath(request.CertName)

	if err != nil {
		return nil, err
	}

	wServer, err := c.wServerFactory(request.WebServer, c.config.ToMap())

	if err != nil {
		return nil, err
	}

	sReverter, err := c.reverterFactory(wServer, c.logger)

	if err != nil {
		return nil, err
	}

	certDeployer := createCertificateDeployer(c.config, wServer, sReverter, c.logger)
	err = certDeployer.DeployCertificate(request.ServerName, certPath, keyPath, false)

	if err != nil {
		return nil, err
	}

	return utils.GetCertificateFromFile(certPath)
}

func (c *CertificateManager) Upload(request request.UploadRequest) (*dto.Certificate, error) {
	var (
		certPath string
		err      error
	)

	storage, err := c.getStorage(Default)

	if err != nil {
		return nil, err
	}

	defaultStorage, ok := storage.(*DefaultStorage)

	if !ok {
		return nil, errors.New("invalid storage")
	}

	if certPath, err = defaultStorage.AddPemCertificate(request.CertName, request.PemCertificate); err != nil {
		return nil, err
	}

	wServer, err := c.wServerFactory(request.WebServer, c.config.ToMap())

	if err != nil {
		return nil, err
	}

	sReverter, err := c.reverterFactory(wServer, c.logger)

	if err != nil {
		return nil, err
	}

	certDeployer := createCertificateDeployer(c.config, wServer, sReverter, c.logger)
	err = certDeployer.DeployCertificate(request.ServerName, certPath, certPath, false)

	if err != nil {
		return nil, err
	}

	return utils.GetCertificateFromFile(certPath)
}

func (c *CertificateManager) GetStorageCertificate(certName, storageType string) (*dto.Certificate, error) {
	storage, err := c.getStorage(CertStorageType(storageType))

	if err != nil {
		return nil, err
	}

	return storage.GetCertificate(certName)
}

func (c *CertificateManager) GetStorageCertificateAsString(certName string, storageType string) (certPath string, certContent string, err error) {
	storage, err := c.getStorage(CertStorageType(storageType))

	if err != nil {
		return "", "", err
	}

	return storage.GetCertificateAsString(certName)
}

func (c *CertificateManager) AddStorageCertificate(certName, pemData string) (string, error) {
	storage, err := c.getStorage(Default)

	if err != nil {
		return "", err
	}

	defaultStorage, ok := storage.(*DefaultStorage)

	if !ok {
		return "", errors.New("invalid storage")
	}

	return defaultStorage.AddPemCertificate(certName, pemData)
}

func (c *CertificateManager) RemoveStorageCertificate(certName, storageType string) error {
	storage, err := c.getStorage(CertStorageType(storageType))

	if err != nil {
		return err
	}

	return storage.RemoveCertificate(certName)
}

func (c *CertificateManager) GetStorageCertificates() ([]CertStorageItem, error) {
	items := []CertStorageItem{}

	for storageType, storage := range c.certStorages {
		certs, err := storage.GetCertificates()

		if err != nil {
			return nil, err
		}

		for certName, cert := range certs {
			items = append(items, CertStorageItem{StorageType: storageType, CertName: certName, Certificate: cert})
		}
	}

	return items, nil
}

func (c *CertificateManager) getStorage(storageType CertStorageType) (CertStorage, error) {
	storage, ok := c.certStorages[storageType]

	if !ok {
		return nil, fmt.Errorf("storage %s not found", storageType)
	}

	return storage, nil
}

func CreateCertificateManager(
	config *config.Config,
	webServerFactory webServerFactory,
	reverterFactory reverterFactory,
	logger logger.Logger,
) (*CertificateManager, error) {
	certStorages := map[CertStorageType]CertStorage{}
	acmeClient, err := client.CreateAcmeClient(config, logger)

	if err != nil {
		return nil, err
	}

	defaultStorage, err := CreateCertStorage(config, logger)

	if err != nil {
		return nil, err
	}

	certStorages[Default] = defaultStorage
	legoStorage, err := lego.CreateCertStorage(config, logger)

	if err != nil {
		return nil, err
	}

	certStorages[Lego] = legoStorage
	certbotStorage, err := certbot.CreateCertStorage(config, logger)

	if err != nil && !errors.Is(err, certbot.ErrStorageDirNotExists) {
		return nil, err
	}

	if err == nil {
		certStorages[CertBot] = certbotStorage
	}

	certManager := &CertificateManager{
		logger:          logger,
		config:          config,
		acmeClient:      acmeClient,
		certStorages:    certStorages,
		wServerFactory:  webServerFactory,
		reverterFactory: reverterFactory,
	}

	return certManager, nil
}

package handler

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/r2dtools/agentintegration"
	"github.com/r2dtools/sslbot/cmd/tcp/contract"
	"github.com/r2dtools/sslbot/cmd/tcp/router"
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/certificates"
	"github.com/r2dtools/sslbot/internal/certificates/commondir"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/utils"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/r2dtools/sslbot/internal/webserver/reverter"
)

type CertificatesHandler struct {
	certManager *certificates.CertificateManager
	logger      logger.Logger
	config      *config.Config
	mx          *sync.Mutex
}

func (h *CertificatesHandler) Handle(request router.Request) (interface{}, error) {
	var response any
	var err error

	switch action := request.GetAction(); action {
	case "issue":
		response, err = h.issueCertificateToDomain(request.Data)
	case "upload":
		response, err = h.uploadCertificateToDomain(request.Data)
	case "storagecertificates":
		response, err = h.storageCertificates()
	case "storagecertdata":
		response, err = h.storageCertData(request.Data)
	case "storagecertupload":
		response, err = h.uploadCertToStorage(request.Data)
	case "storagecertremove":
		err = h.removeCertFromStorage(request.Data)
	case "storagecertdownload":
		response, err = h.downloadCertFromStorage(request.Data)
	case "domainassign":
		response, err = h.assignCertificateToDomain(request.Data)
	case "commondirstatus":
		response, err = h.commonDirStatus(request.Data)
	case "changecommondirstatus":
		err = h.changeCommonDirStatus(request.Data)
	default:
		response, err = nil, fmt.Errorf("invalid action '%s' for module '%s'", action, request.GetModule())
	}

	return response, err
}

func (h *CertificatesHandler) issueCertificateToDomain(data any) (*agentintegration.Certificate, error) {
	var request agentintegration.CertificateIssueRequestData
	err := mapstructure.Decode(data, &request)

	if err != nil {
		return nil, fmt.Errorf("invalid request data: %v", err)
	}

	cert, err := h.certManager.Issue(contract.ConvertIssueRequest(request))

	if err != nil {
		return nil, err
	}

	return contract.ConvertCertificate(cert), nil
}

func (h *CertificatesHandler) uploadCertificateToDomain(data any) (*agentintegration.Certificate, error) {
	var request agentintegration.CertificateUploadRequestData
	err := mapstructure.Decode(data, &request)

	if err != nil {
		return nil, fmt.Errorf("invalid request data: %v", err)
	}

	if request.ServerName == "" {
		return nil, errors.New("domain name is missed")
	}

	cert, err := h.certManager.Upload(contract.ConvertUploadRequest(request))

	if err != nil {
		return nil, err
	}

	return contract.ConvertCertificate(cert), nil
}

func (h *CertificatesHandler) storageCertificates() (*agentintegration.CertificatesResponseData, error) {
	certItems, err := h.certManager.GetStorageCertificates()

	if err != nil {
		return nil, err
	}

	certsMap := map[string]*agentintegration.Certificate{}

	for _, item := range certItems {
		certsMap[item.Key()] = contract.ConvertCertificate(item.Certificate)
	}

	return &agentintegration.CertificatesResponseData{Certificates: certsMap}, nil
}

func (h *CertificatesHandler) storageCertData(data any) (*agentintegration.Certificate, error) {
	var request agentintegration.CertificateInfoRequestData
	err := mapstructure.Decode(data, &request)

	if err != nil {
		return nil, fmt.Errorf("invalid request data: %v", err)
	}

	cert, err := h.certManager.GetStorageCertificate(request.CertName, request.StorageType)

	if err != nil {
		return nil, err
	}

	return contract.ConvertCertificate(cert), nil
}

func (h *CertificatesHandler) uploadCertToStorage(data any) (*agentintegration.Certificate, error) {
	var request agentintegration.CertificateUploadRequestData
	err := mapstructure.Decode(data, &request)

	if err != nil {
		return nil, fmt.Errorf("invalid request data: %v", err)
	}

	if request.CertName == "" {
		return nil, errors.New("certificate name is missed")
	}

	certPath, err := h.certManager.AddStorageCertificate(request.CertName, request.PemCertificate)

	if err != nil {
		return nil, err
	}

	cert, err := utils.GetCertificateFromFile(certPath)

	if err != nil {
		return nil, err
	}

	return contract.ConvertCertificate(cert), nil
}

func (h *CertificatesHandler) removeCertFromStorage(data any) error {
	var request agentintegration.CertificateRemoveRequestData
	err := mapstructure.Decode(data, &request)

	if err != nil {
		return fmt.Errorf("invalid request data: %v", err)
	}

	return h.certManager.RemoveStorageCertificate(request.CertName, request.StorageType)
}

func (h *CertificatesHandler) downloadCertFromStorage(data any) (*agentintegration.CertificateDownloadResponseData, error) {
	var request agentintegration.CertificateRemoveRequestData
	err := mapstructure.Decode(data, &request)

	if err != nil {
		return nil, fmt.Errorf("invalid request data: %v", err)
	}

	certPath, certContent, err := h.certManager.GetStorageCertificateAsString(request.CertName, request.StorageType)

	if err != nil {
		return nil, err
	}

	var certDownloadResponse agentintegration.CertificateDownloadResponseData
	certDownloadResponse.CertFileName = filepath.Base(certPath)
	certDownloadResponse.CertContent = certContent

	return &certDownloadResponse, nil
}

func (h *CertificatesHandler) assignCertificateToDomain(data any) (*agentintegration.Certificate, error) {
	var request agentintegration.CertificateAssignRequestData
	err := mapstructure.Decode(data, &request)

	if err != nil {
		return nil, fmt.Errorf("invalid request data: %v", err)
	}

	cert, err := h.certManager.Assign(contract.ConvertAssignRequest(request))

	if err != nil {
		return nil, err
	}

	return contract.ConvertCertificate(cert), nil
}

func (h *CertificatesHandler) commonDirStatus(data any) (*agentintegration.CommonDirStatusResponseData, error) {
	var requestData agentintegration.CommonDirChangeStatusRequestData
	err := mapstructure.Decode(data, &requestData)

	if err != nil {
		return nil, fmt.Errorf("invalid request data: %v", err)
	}

	options := h.config.ToMap()
	wServer, err := webserver.GetWebServer(requestData.WebServer, options)

	if err != nil {
		return nil, err
	}

	commonDirQuery, err := commondir.CreateCommonDirStatusQuery(wServer)

	if err != nil {
		return nil, err
	}

	status := commonDirQuery.GetCommonDirStatus(requestData.ServerName)

	return &agentintegration.CommonDirStatusResponseData{Status: status.Enabled}, nil
}

func (h *CertificatesHandler) changeCommonDirStatus(data any) error {
	var requestData agentintegration.CommonDirChangeStatusRequestData
	err := mapstructure.Decode(data, &requestData)

	if err != nil {
		return fmt.Errorf("invalid request data: %v", err)
	}

	options := h.config.ToMap()
	wServer, err := webserver.GetWebServer(requestData.WebServer, options)

	if err != nil {
		return err
	}

	sReverter, err := reverter.CreateReverter(wServer, h.logger)

	if err != nil {
		return err
	}

	commonDirCommand, err := commondir.CreateCommonDirChangeCommand(h.config, wServer, sReverter, h.logger, h.mx)

	if err != nil {
		return err
	}

	if requestData.Status {
		err = commonDirCommand.EnableCommonDir(requestData.ServerName)
	} else {
		err = commonDirCommand.DisableCommonDir(requestData.ServerName)
	}

	return err
}

func CreateCertificatesHandler(config *config.Config, logger logger.Logger, mx *sync.Mutex) (router.HandlerInterface, error) {
	certManager, err := certificates.CreateCertificateManager(
		config,
		webserver.GetWebServer,
		reverter.CreateReverter,
		logger,
		mx,
	)

	if err != nil {
		return nil, err
	}

	return &CertificatesHandler{
		logger:      logger,
		certManager: certManager,
		config:      config,
		mx:          mx,
	}, nil
}

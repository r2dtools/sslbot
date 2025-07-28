package handler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/r2dtools/agentintegration"
	"github.com/r2dtools/sslbot/cmd/tcp/contract"
	"github.com/r2dtools/sslbot/cmd/tcp/router"
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/r2dtools/sslbot/internal/utils"
	"github.com/r2dtools/sslbot/internal/webserver"
	"github.com/shirou/gopsutil/host"
)

type MainHandler struct {
	config *config.Config
	logger logger.Logger
	mx     *sync.Mutex
}

func (h *MainHandler) Handle(request router.Request) (any, error) {
	var response any
	var err error

	switch action := request.GetAction(); action {
	case "getserverdata":
		response, err = h.getServerData()
	case "getVhosts":
		response, err = h.getVhosts()
	case "getVhostCertificate":
		response, err = h.getVhostCertificate(request.Data)
	case "getvhostconfig":
		response, err = h.getVhostConfig(request.Data)
	case "reloadwebserver":
		err = h.reloadWebServer(request.Data)
	default:
		response, err = nil, fmt.Errorf("invalid action '%s' for module '%s'", action, request.GetModule())
	}

	return response, err
}

func (h *MainHandler) getServerData() (agentintegration.ServerData, error) {
	var serverData agentintegration.ServerData
	info, err := host.Info()

	if err != nil {
		return serverData, fmt.Errorf("failed to load server data: %v", err)
	}

	serverData.BootTime = info.BootTime
	serverData.Uptime = info.Uptime
	serverData.KernelArch = info.KernelArch
	serverData.KernelVersion = info.KernelVersion
	serverData.HostName = info.Hostname
	serverData.Platform = info.Platform
	serverData.PlatformFamily = info.PlatformFamily
	serverData.PlatformVersion = info.PlatformVersion
	serverData.Os = info.OS
	serverData.AgentVersion = h.config.Version

	certbotStatus := "false"

	if h.config.CertBotEnabled {
		certbotStatus = "true"
	}

	serverData.Settings = map[string]string{
		"certbotstatus": certbotStatus,
	}

	return serverData, nil
}

func (h *MainHandler) getVhosts() ([]agentintegration.VirtualHost, error) {
	webServerCodes := webserver.GetSupportedWebServers()
	var vhosts []agentintegration.VirtualHost
	options := h.config.ToMap()

	for _, webServerCode := range webServerCodes {
		webserver, err := webserver.GetWebServer(webServerCode, options)

		if err != nil {
			h.logger.Error(err.Error())
			continue
		}

		wVhosts, err := webserver.GetVhosts()

		if err != nil {
			h.logger.Error(err.Error())
			continue
		}

		vhosts = append(vhosts, contract.ConvertVirtualHosts(wVhosts)...)
	}

	return vhosts, nil
}

func (h *MainHandler) getVhostCertificate(data any) (*agentintegration.Certificate, error) {
	mData, ok := data.(map[string]any)

	if !ok {
		return nil, errors.New("invalid request data format")
	}

	vhostNameRaw, ok := mData["vhostName"]

	if !ok {
		return nil, errors.New("invalid request data: vhost name is not specified")
	}

	vhostName, ok := vhostNameRaw.(string)

	if !ok {
		return nil, errors.New("invalid request data: vhost name is invalid")
	}

	cert, err := utils.GetCertificateForDomainFromRequest(vhostName)

	if err != nil {
		message := "could not get vhost '%s' certificate: %v"
		h.logger.Info(message, vhostName, err)

		return nil, fmt.Errorf(message, vhostName, err)
	}

	return contract.ConvertCertificate(cert), nil
}

func (h *MainHandler) getVhostConfig(data any) (agentintegration.VirtualHostConfigResponseData, error) {
	var response agentintegration.VirtualHostConfigResponseData
	var request agentintegration.VirtualHostConfigRequestData

	err := mapstructure.Decode(data, &request)

	if err != nil {
		return response, fmt.Errorf("invalid vhodt config request data: %v", err)
	}

	options := h.config.ToMap()
	wServer, err := webserver.GetWebServer(request.WebServer, options)

	if err != nil {
		return response, err
	}

	vhost, err := wServer.GetVhostByName(request.ServerName)

	if err != nil {
		return response, err
	}

	if vhost == nil {
		return response, fmt.Errorf("vhost %s not found", request.ServerName)
	}

	configFile, err := os.Open(vhost.FilePath)

	if err != nil {
		return response, err
	}

	content, err := io.ReadAll(configFile)

	if err != nil {
		return response, err
	}

	response.Content = string(content)

	return response, nil
}

func (h *MainHandler) reloadWebServer(data any) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	var request agentintegration.ReloadWebServerRequestData

	err := mapstructure.Decode(data, &request)

	if err != nil {
		return fmt.Errorf("invalid vhodt config request data: %v", err)
	}

	wServer, err := webserver.GetWebServer(request.WebServer, h.config.ToMap())

	if err != nil {
		return err
	}

	p, err := wServer.GetProcessManager()

	if err != nil {
		return err
	}

	return p.Reload()
}

func CreateMainHandler(config *config.Config, logger logger.Logger, mx *sync.Mutex) *MainHandler {
	return &MainHandler{
		config: config,
		logger: logger,
		mx:     mx,
	}
}

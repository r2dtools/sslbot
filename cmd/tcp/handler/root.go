package handler

import (
	"errors"
	"fmt"
	"io"
	"os"

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
	Config *config.Config
	Logger logger.Logger
}

func (h *MainHandler) Handle(request router.Request) (any, error) {
	var response interface{}
	var err error

	switch action := request.GetAction(); action {
	case "refresh":
		response, err = h.refresh()
	case "getVhosts":
		response, err = h.getVhosts()
	case "getVhostCertificate":
		response, err = h.getVhostCertificate(request.Data)
	case "getvhostconfig":
		response, err = h.getVhostConfig(request.Data)
	default:
		response, err = nil, fmt.Errorf("invalid action '%s' for module '%s'", action, request.GetModule())
	}

	return response, err
}

func (h *MainHandler) refresh() (*agentintegration.ServerData, error) {
	info, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("could not get system info: %v", err)
	}
	var serverData agentintegration.ServerData
	serverData.BootTime = info.BootTime
	serverData.Uptime = info.Uptime
	serverData.KernelArch = info.KernelArch
	serverData.KernelVersion = info.KernelVersion
	serverData.HostName = info.Hostname
	serverData.Platform = info.Platform
	serverData.PlatformFamily = info.PlatformFamily
	serverData.PlatformVersion = info.PlatformVersion
	serverData.Os = info.OS

	serverData.AgentVersion = h.Config.Version

	return &serverData, nil
}

func (h *MainHandler) getVhosts() ([]agentintegration.VirtualHost, error) {
	webServerCodes := webserver.GetSupportedWebServers()
	var vhosts []agentintegration.VirtualHost
	options := h.Config.ToMap()

	for _, webServerCode := range webServerCodes {
		webserver, err := webserver.GetWebServer(webServerCode, options)

		if err != nil {
			h.Logger.Error(err.Error())
			continue
		}

		wVhosts, err := webserver.GetVhosts()

		if err != nil {
			h.Logger.Error(err.Error())
			continue
		}

		vhosts = append(vhosts, contract.ConvertVirtualHosts(wVhosts)...)
	}

	return vhosts, nil
}

func (h *MainHandler) getVhostCertificate(data any) (*agentintegration.Certificate, error) {
	mData, ok := data.(map[string]interface{})

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
		h.Logger.Info(message, vhostName, err)

		return nil, fmt.Errorf(message, vhostName, err)
	}

	return contract.ConvertCertificate(cert), nil
}

func (h *MainHandler) getVhostConfig(data interface{}) (agentintegration.VirtualHostConfigResponseData, error) {
	var response agentintegration.VirtualHostConfigResponseData
	var request agentintegration.VirtualHostConfigRequestData

	err := mapstructure.Decode(data, &request)

	if err != nil {
		return response, fmt.Errorf("invalid vhodt config request data: %v", err)
	}

	options := h.Config.ToMap()
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

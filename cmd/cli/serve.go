package cli

import (
	"sync"

	"github.com/r2dtools/sslbot/cmd/tcp"
	"github.com/r2dtools/sslbot/cmd/tcp/handler"
	"github.com/r2dtools/sslbot/cmd/tcp/router"
	"github.com/r2dtools/sslbot/config"
	"github.com/r2dtools/sslbot/internal/logger"
	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts TCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.GetConfig()

		if err != nil {
			return err
		}

		logger, err := logger.NewLogger(config)

		if err != nil {
			return err
		}

		mx := &sync.Mutex{}
		mainHandler := handler.CreateMainHandler(config, logger, mx)
		certificatesHandler, err := handler.CreateCertificatesHandler(config, logger, mx)

		if err != nil {
			return err
		}

		router := router.Router{}
		router.RegisterHandler("main", mainHandler)
		router.RegisterHandler("certificates", certificatesHandler)

		server := &tcp.Server{
			Port:   config.Port,
			Router: router,
			Logger: logger,
			Config: config,
		}

		return server.Serve()
	},
}

package cli

import (
	"github.com/r2dtools/sslbot/cmd/tcp"
	"github.com/r2dtools/sslbot/cmd/tcp/handler"
	certificates "github.com/r2dtools/sslbot/cmd/tcp/handler"
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

		certificatesHandler, err := certificates.GetHandler(config, logger)

		if err != nil {
			return err
		}

		router := router.Router{}
		router.RegisterHandler("main", &handler.MainHandler{
			Config: config,
			Logger: logger,
		})
		router.RegisterHandler("certificates", certificatesHandler)

		server := &tcp.Server{
			Port:   config.Port,
			Router: router,
			Logger: logger,
			Config: config,
		}

		if err := server.Serve(); err != nil {
			return err
		}

		return nil
	},
}

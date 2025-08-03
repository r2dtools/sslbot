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
		conf, err := config.GetConfig()

		if err != nil {
			return err
		}

		logger, err := logger.NewLogger(conf)

		if err != nil {
			return err
		}

		mx := &sync.Mutex{}
		mainHandler := handler.CreateMainHandler(conf, logger, mx)
		certificatesHandler, err := handler.CreateCertificatesHandler(conf, logger, mx)

		if err != nil {
			return err
		}

		botRouter := router.Router{}
		botRouter.RegisterHandler("main", mainHandler)
		botRouter.RegisterHandler("certificates", certificatesHandler)

		server := &tcp.Server{
			Port:   conf.Port,
			Router: botRouter,
			Logger: logger,
			Config: conf,
		}

		conf.OnChange(func() {
			logger.Info("reload router ...")
			mainHandler = handler.CreateMainHandler(conf, logger, mx)
			certificatesHandler, err = handler.CreateCertificatesHandler(conf, logger, mx)

			if err != nil {
				logger.Error("router reload failed: %v", err)

				return
			}

			botRouter = router.Router{}
			botRouter.RegisterHandler("main", mainHandler)
			botRouter.RegisterHandler("certificates", certificatesHandler)

			server.Router = botRouter
		})

		return server.Serve()
	},
}

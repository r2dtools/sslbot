package cli

import (
	"fmt"

	"github.com/r2dtools/sslbot/config"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show SSLBot version",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.GetConfig()

		if err != nil {
			return err
		}

		fmt.Println(config.Version)

		return nil
	},
}

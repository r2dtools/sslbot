package cli

import (
	"fmt"

	"github.com/r2dtools/sslbot/config"
	"github.com/spf13/cobra"

	"github.com/google/uuid"
)

var GenerateTokenCmd = &cobra.Command{
	Use:   "generate-token",
	Short: "Generate new token",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.GetConfig()

		if err != nil {
			return err
		}

		if err := config.CreateConfigFileIfNotExists(conf); err != nil {
			return err
		}

		randomUuid, err := uuid.NewRandom()

		if err != nil {
			return err
		}

		token := randomUuid.String()
		err = conf.SetParam(config.TokenOpt, token)

		if err != nil {
			return err
		}

		fmt.Printf("Token: %s\n", token)

		return nil
	},
}

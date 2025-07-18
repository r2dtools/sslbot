package cli

import (
	"fmt"
	"os"

	"github.com/r2dtools/sslbot/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

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
		data, err := os.ReadFile(conf.ConfigFilePath)

		if err != nil {
			return err
		}

		confMap := make(map[string]interface{})
		err = yaml.Unmarshal(data, confMap)

		if err != nil {
			return err
		}

		confMap["token"] = token
		data, err = yaml.Marshal(confMap)

		if err != nil {
			return err
		}

		err = os.WriteFile(conf.ConfigFilePath, data, 0644)

		if err != nil {
			return err
		}

		fmt.Printf("Token: %s\n", token)

		return nil
	},
}

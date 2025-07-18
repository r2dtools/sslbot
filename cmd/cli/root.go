package cli

import (
	"github.com/spf13/cobra"
)

var isJson bool

var webServerCode string
var serverName string

func Create() *cobra.Command {
	cli := &cobra.Command{
		Use:   "sslbot",
		Short: "R2DTools SSLBot",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	cli.PersistentFlags().BoolVarP(&isJson, "json", "j", false, "show result in json format")
	cli.AddCommand(ServeCmd)
	cli.AddCommand(VersionCmd)
	cli.AddCommand(HostsCmd)
	cli.AddCommand(DeployCertificateCmd)
	cli.AddCommand(IssueCertificateCmd)
	cli.AddCommand(GenerateTokenCmd)
	cli.AddCommand(CommonDirCmd)
	cli.AddCommand(ShowTokenCmd)
	cli.PersistentFlags().StringVarP(&webServerCode, "webserver", "w", "", "webserver (nginx|apache)")

	return cli
}

func writeOutput(cmd *cobra.Command, output string) error {
	_, err := cmd.OutOrStdout().Write([]byte(output))

	if err != nil {
		return err
	}

	return nil
}

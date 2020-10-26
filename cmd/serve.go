package cmd

import (
	"github.com/Vilsol/yeet/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the webserver",
	RunE: func(cmd *cobra.Command, args []string) error {
		return server.RunWebserver()
	},
}

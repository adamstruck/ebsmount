package cmd

import (
	"github.com/adamstruck/ebsmount/server"
	"github.com/spf13/cobra"
)

var port string

func init() {
	f := serverCmd.Flags()
	f.StringVarP(&port, "port", "p", "9000", "http port")

	RootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start ebsmount as a service.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return server.Run(port)
	},
}

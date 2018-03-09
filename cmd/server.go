package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/adamstruck/ebsmount/server"
	"github.com/spf13/cobra"
)

var socket string

func init() {
	f := serverCmd.Flags()
	f.StringVarP(&socket, "socket", "s", "./ebsmount.sock", "unix socket")

	RootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start ebsmount as a service.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sch := make(chan os.Signal, 1)
		signal.Notify(sch, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sch
			cancel()
		}()

		return server.Run(ctx, socket)
	},
}

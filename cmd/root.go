package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "ebsmount",
	Short:         "Mount or unmount EBS volume(s) to an EC2 instance.",
	SilenceUsage:  true,
	SilenceErrors: false,
}

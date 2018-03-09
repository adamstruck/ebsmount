package cmd

import (
	"fmt"
	"os"

	"github.com/adamstruck/ebsmount/ebsmount"
	"github.com/spf13/cobra"
)

var volumeID string

func init() {
	f := unmountCmd.Flags()
	f.StringVarP(&volumeID, "volume-id", "v", "", "EBS volume ID to detach and/or delete from instance")

	RootCmd.AddCommand(unmountCmd)
}

var unmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount an EBS volume from an EC2 instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if volumeID == "" {
			fmt.Fprintln(os.Stderr, "required flag 'volume-id' not set", "\n")
			fmt.Fprintln(os.Stderr, cmd.UsageString())
			return fmt.Errorf("invalid flag(s)")
		}

		mounter, err := ebsmount.NewEC2Mounter()
		if err != nil {
			return err
		}

		return mounter.DetachAndDelete(volumeID)
	},
}

package cmd

import (
	"fmt"
	"os"

	"github.com/adamstruck/ebsmount/ebsmount"
	"github.com/adamstruck/ebsmount/server"
	"github.com/spf13/cobra"
)

var unmountReq = &server.UnmountRequest{}

func init() {
	f := unmountCmd.Flags()
	f.StringVarP(&unmountReq.VolumeID, "volume-id", "v", unmountReq.VolumeID, "EBS volume ID to detach and/or delete from instance")
	f.StringVarP(&unmountReq.MountPoint, "mount-point", "m", unmountReq.MountPoint, "directory to unmount")

	RootCmd.AddCommand(unmountCmd)
}

var unmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount an EBS volume from an EC2 instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := unmountReq.Validate()
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "\n")
			fmt.Fprintln(os.Stderr, cmd.UsageString())
			return fmt.Errorf("invalid flag(s)")
		}

		mounter, err := ebsmount.NewEC2Mounter()
		if err != nil {
			return err
		}

		return mounter.DetachAndDelete(unmountReq.VolumeID, unmountReq.MountPoint)
	},
}

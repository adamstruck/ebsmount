package cmd

import (
	"fmt"
	"os"

	"github.com/adamstruck/ebsmount/ebsmount"
	"github.com/spf13/cobra"
)

func defaultMountArgs() *ebsmount.MountRequest {
	return &ebsmount.MountRequest{
		Size:       200,
		VolumeType: "gp2",
		FSType:     "ext4",
	}
}

var mountReq = defaultMountArgs()

func init() {
	f := mountCmd.Flags()
	f.Int64VarP(&mountReq.Size, "size", "s", mountReq.Size, "size in GB of desired EBS volume")
	f.StringVarP(&mountReq.MountPoint, "mount-point", "m", mountReq.MountPoint, "directory on which to mount the EBS volume")
	f.StringVarP(&mountReq.VolumeType, "volume-type", "v", mountReq.VolumeType, "desired volume type; gp2 for General Purpose SSD; io1 for Provisioned IOPS SSD; st1 for Throughput Optimized HDD; sc1 for HDD or Magnetic volumes; standard for infrequent")
	f.StringVarP(&mountReq.FSType, "fs-type", "t", mountReq.FSType, "file system type to create (argument must be accepted by mkfs)")
	f.Int64VarP(&mountReq.Iops, "iops", "i", mountReq.Iops, "Provisioned IOPS. Only valid for volume type io1. Range is 100 to 20000 and <= 50*size of volume")
	f.BoolVarP(&mountReq.Keep, "keep", "k", mountReq.Keep, "don't delete the volume on termination (default is to delete)")

	RootCmd.AddCommand(mountCmd)
}

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount an EBS volume to an EC2 instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := mountReq.Validate()
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "\n")
			fmt.Fprintln(os.Stderr, cmd.UsageString())
			return fmt.Errorf("invalid flag(s)")
		}

		mounter, err := ebsmount.NewEC2Mounter()
		if err != nil {
			return err
		}

		_, err = mounter.CreateAndMount(mountReq)
		if err != nil {
			return err
		}
		return nil
	},
}

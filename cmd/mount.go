package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

var cli = defaultMountArgs()

func init() {
	f := mountCmd.Flags()
	f.Int64VarP(&cli.Size, "size", "s", cli.Size, "size in GB of desired EBS volume")
	f.StringVarP(&cli.MountPoint, "mount-point", "m", cli.MountPoint, "directory on which to mount the EBS volume")
	f.StringVarP(&cli.VolumeType, "volume-type", "v", cli.VolumeType, "desired volume type; gp2 for General Purpose SSD; io1 for Provisioned IOPS SSD; st1 for Throughput Optimized HDD; sc1 for HDD or Magnetic volumes; standard for infrequent")
	f.StringVarP(&cli.FSType, "fs-type", "t", cli.FSType, "file system type to create (argument must be accepted by mkfs)")
	f.Int64VarP(&cli.Iops, "iops", "i", cli.Iops, "Provisioned IOPS. Only valid for volume type io1. Range is 100 to 20000 and <= 50*size of volume")
	f.BoolVarP(&cli.Keep, "keep", "k", cli.Keep, "don't delete the volume on termination (default is to delete)")

	RootCmd.AddCommand(mountCmd)
}

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount an EBS volume to an EC2 instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		validationErrs := []string{}

		if cli.MountPoint == "" {
			validationErrs = append(validationErrs, "required flag 'mount-point' not set")
		} else if !filepath.IsAbs(cli.MountPoint) {
			validationErrs = append(validationErrs, "invalid value passed to 'mount-point' flag; must be an absolute path")
		}

		if cli.VolumeType != "gp2" && cli.VolumeType != "io1" && cli.VolumeType != "st1" && cli.VolumeType != "sc1" && cli.VolumeType != "standard" {
			validationErrs = append(validationErrs, "invalid value passed to 'volume-type' flag; must be one of [ 'gp2', 'io1', 'st1', 'sc1', 'standard' ]")
		}

		if cli.FSType != "ext4" && cli.FSType != "ext3" && cli.FSType != "ext2" {
			validationErrs = append(validationErrs, "invalid value passed to 'fs-type' flag; must be one of [ 'ext4', 'ext3', 'ext2' ]")
		}

		if cli.Iops != 0 && (cli.Iops < 100 || cli.Iops > 20000) {
			validationErrs = append(validationErrs, "invalid value passed to 'iops' flag; range is 100 to 20000 and <= 50*size of volume")
		}

		if cli.Size < 0 {
			validationErrs = append(validationErrs, "invalid value passed to 'size' flag; must be a positive integer")
		}

		if len(validationErrs) > 0 {
			fmt.Fprintln(os.Stderr, strings.Join(validationErrs, "\n"), "\n")
			fmt.Fprintln(os.Stderr, cmd.UsageString())
			return fmt.Errorf("invalid flag(s)")
		}

		mounter, err := ebsmount.NewEC2Mounter()
		if err != nil {
			return err
		}

		_, err = mounter.CreateAndMount(cli)
		if err != nil {
			return err
		}
		return nil
	},
}

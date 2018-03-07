package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/base2genomics/batchit/exsmount"
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:   "ebsmount",
	Short: "Mount an EBS volume to an EC2 instance and run a command.",
	RunE: func(cmd *cobra.Command, args []string) error {
		
		if cli.Exsmount.VolumeType != "gp2" && cli.Exsmount.VolumeType != "io1" && cli.Exsmount.VolumeType != "st1" && cli.Exsmount.VolumeType != "sc1" && cli.Exsmount.VolumeType != "standard" {
			return fmt.Errorf("invalid value passed to 'volume-type' flag; must be one of [ 'gp2', 'io1', 'st1', 'sc1', 'standard' ]")
		}
		
		if cli.Exsmount.FSType != "ext4" {
			return fmt.Errorf("invalid value passed to 'fs-type' flag; must be one of [ 'ext4' ]")
		}

		if cli.Exsmount.Iops != 0 && (cli.Exsmount.Iops < 100 || cli.Exsmount.Iops > 20000) {
			return fmt.Errorf("invalid value passed to 'iops' flag; range is 100 to 20000 and <= 50*size of volume")
		}

		if cli.Exsmount.Size < 0 {
			return fmt.Errorf("invalid value passed to 'size' flag; must be a positive integer")
		}

		if !filepath.IsAbs(cli.Exsmount.MountPoint) {
			return fmt.Errorf("invalid value passed to 'mount-point' flag; must be an absolute path")
		}

		return MountAndRun(cli)
	},
}

type Args struct {
	Command    string
	Exsmount   *exsmount.Args
}

func DefaultArgs() *Args {
	return &Args{
		Exsmount: &exsmount.Args{
			Size:       200,
			VolumeType: "gp2",
			FSType:     "ext4",
			N: 1, 
		},
	}
}

var cli = DefaultArgs()

func init() {
	f := RootCmd.Flags()
	f.StringVarP(&cli.Command, "command", "c", cli.Command, "Command to run after volume mounting completes")
	f.Int64VarP(&cli.Exsmount.Size, "size", "s", cli.Exsmount.Size, "size in GB of desired EBS volume")
	f.StringVarP(&cli.Exsmount.MountPoint, "mount-point", "m", cli.Exsmount.MountPoint, "directory on which to mount the EBS volume")
	f.StringVarP(&cli.Exsmount.VolumeType, "volume-type", "v", cli.Exsmount.VolumeType, "desired volume type; gp2 for General Purpose SSD; io1 for Provisioned IOPS SSD; st1 for Throughput Optimized HDD; sc1 for HDD or Magnetic volumes; standard for infrequent")
	f.StringVarP(&cli.Exsmount.FSType, "fs-type", "t", cli.Exsmount.FSType, "file system type to create (argument must be accepted by mkfs)")
	f.Int64VarP(&cli.Exsmount.Iops, "iops", "i", cli.Exsmount.Iops, "Provisioned IOPS. Only valid for volume type io1. Range is 100 to 20000 and <= 50*size of volume")
	f.BoolVarP(&cli.Exsmount.Keep, "keep", "k", cli.Exsmount.Keep, "don't delete the volume on termination (default is to delete)")

	RootCmd.MarkFlagRequired("command")
	RootCmd.MarkFlagRequired("mount-point")
}

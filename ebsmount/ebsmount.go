package ebsmount

import (
	"fmt"
	"os/exec"

	"github.com/adamstruck/ebsmount/cmd"
	"github.com/kballard/go-shellquote"
	"github.com/base2genomics/batchit/exsmount"
)

func MountAndRun(args *cmd.Args) error {
	command, err := shellquote.Split(args.Command)
	if err != nil {
		return fmt.Errorf("failed to parse command: %v", err)
	}
	
	cli := args.Exsmount
	// This method prints the volume id to stdout...
	devices, err := exsmount.CreateAttach(cli)
	if err != nil {
		return fmt.Errorf("CreateAttach call failed: %v", err)
	}

	if devices, err := MountLocal(devices, cli.MountPoint); err != nil {
		return fmt.Errorf("MountLocal call failed: %v", err)
	} else if cli.VolumeType == "st1" || cli.VolumeType == "sc1" {
		// https://aws.amazon.com/blogs/aws/amazon-ebs-update-new-cold-storage-and-throughput-options/
		for _, d := range devices {
			cmd := exec.Command("blockdev", "--setra", "2048", d)
			cmd.Stderr, cmd.Stdout = os.Stderr, os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: error setting read-ahead\n", err)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "mounted %d EBS drives to %s\n", len(devices), cli.MountPoint)

	cmd = exec.Command(command...)
	cmd.Stderr, cmd.Stdout = os.Stderr, os.Stderr
	return cmd.Run()
}

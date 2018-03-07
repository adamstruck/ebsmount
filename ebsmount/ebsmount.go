package ebsmount

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/base2genomics/batchit/ddv"
	"github.com/base2genomics/batchit/exsmount"
	"github.com/kballard/go-shellquote"
)

type Args struct {
	Command  string
	Exsmount *exsmount.Args
}

func DefaultArgs() *Args {
	return &Args{
		Exsmount: &exsmount.Args{
			Size:       200,
			VolumeType: "gp2",
			FSType:     "ext4",
			N:          1,
		},
	}
}

func MountAndRun(args *Args) (err error) {
	parsedCmd, err := shellquote.Split(args.Command)
	if err != nil {
		return fmt.Errorf("failed to parse command: %v", err)
	}

	// Create and mount the EBS volume
	// This method prints the volume id to stdout...
	OGStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	devices, err := exsmount.CreateAttach(args.Exsmount)
	if err != nil {
		return fmt.Errorf("CreateAttach call failed: %v", err)
	}
	w.Close()
	vid, _ := ioutil.ReadAll(r)
	os.Stdout = OGStdout

	if devices, err := exsmount.MountLocal(devices, args.Exsmount.MountPoint); err != nil {
		return fmt.Errorf("MountLocal call failed: %v", err)
	} else if args.Exsmount.VolumeType == "st1" || args.Exsmount.VolumeType == "sc1" {
		// https://aws.amazon.com/blogs/aws/amazon-ebs-update-new-cold-storage-and-throughput-options/
		for _, d := range devices {
			cmd := exec.Command("blockdev", "--setra", "2048", d)
			cmd.Stderr, cmd.Stdout = os.Stderr, os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: error setting read-ahead: %v\n", err)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "mounted %d EBS drives to %s\n", len(devices), args.Exsmount.MountPoint)

	// Cleanup the volume on exit
	defer func() {
		derr := ddv.DetachAndDelete(string(vid))
		if derr != nil {
			if err != nil {
				err = fmt.Errorf("command error: %v; detach and delete volume error: %v", err, derr)
			} else {
				err = fmt.Errorf("detach and delete volume error: %v", derr)
			}
		}
	}()

	// Run the command
	cmdEntry := parsedCmd[0]
	cmdArgs := parsedCmd[1:]
	cmd := exec.Command(cmdEntry, cmdArgs...)
	cmd.Stderr, cmd.Stdout = os.Stderr, os.Stderr
	return cmd.Run()
}

package ebsmount

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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
	devices, volumes, err := exsmount.CreateAttach(args.Exsmount)
	if err != nil {
		return fmt.Errorf("CreateAttach call failed: %v", err)
	}

	// Cleanup the volume on exit
	defer func() {
		errs := []string{}
		for _, vid := range volumes {
			derr := ddv.DetachAndDelete(string(vid))
			if derr != nil {
				errs = append(errs, derr.Error())
			}
		}
		if len(errs) > 0 {
			if err != nil {
				err = fmt.Errorf("original error: %v; detach and delete volume error: %v", err, strings.Join(errs, "; "))
			} else {
				err = fmt.Errorf("detach and delete volume error: %v", strings.Join(errs, "; "))
			}
		}
	}()

	// Mount the newly created volume
	if devices, err := exsmount.MountLocal(devices, args.Exsmount.MountPoint); err != nil {
		return fmt.Errorf("MountLocal call failed: %v", err)
	} else if args.Exsmount.VolumeType == "st1" || args.Exsmount.VolumeType == "sc1" {
		// https://aws.amazon.com/blogs/aws/amazon-ebs-update-new-cold-storage-and-throughput-options/
		for _, d := range devices {
			cmd := exec.Command("blockdev", "--setra", "2048", d)
			cmd.Stderr, cmd.Stdout = os.Stderr, os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("warning: error setting read-ahead: %v\n", err)
			}
		}
	}
	log.Printf("mounted %d EBS drive to %s\n", len(devices), args.Exsmount.MountPoint)

	// Run the command
	log.Println("Running command:", strings.Join(parsedCmd, " "))
	cmd := exec.Command(parsedCmd[0], parsedCmd[1:]...)
	cmd.Stderr, cmd.Stdout = os.Stderr, os.Stderr
	return cmd.Run()
}

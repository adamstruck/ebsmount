package ebsmount

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type MountRequest struct {
	Size       int64
	MountPoint string
	VolumeType string
	FSType     string
	Iops       int64
	Keep       bool
}

type MountResponse struct {
	Device   string
	VolumeID string
}

type EC2Mounter struct {
	Session *session.Session
	EC2     *ec2.EC2
	IID     ec2metadata.EC2InstanceIdentityDocument
}

func NewEC2Mounter() (*EC2Mounter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "error creating aws session")
	}

	metasvc := ec2metadata.New(sess)
	iid, err := metasvc.GetInstanceIdentityDocument()
	if err != nil {
		return nil, errors.Wrap(err, "error getting instance identity document")
	}

	ec2Svc := ec2.New(sess, &aws.Config{Region: aws.String(iid.Region), MaxRetries: aws.Int(3)})

	return &EC2Mounter{
		Session: sess,
		EC2:     ec2Svc,
		IID:     iid,
	}, nil
}

func (mounter *EC2Mounter) create(size int64, vtype string, iops int64) (*ec2.Volume, error) {
	cvi := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(mounter.IID.AvailabilityZone),
		Size:             aws.Int64(size), //GB
		VolumeType:       aws.String(vtype),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("volume"),
				Tags:         []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("batchit-%s", mounter.IID.InstanceID))}},
			},
		},
	}

	if vtype == "io1" {
		if iops == 0 {
			iops = 45 * size
		}
		if iops < 100 || iops > 20000 {
			return nil, fmt.Errorf("iops must be between 100 and 20000")
		}
		if iops > 50*size {
			iops = 45 * size
			if iops > 200000 {
				iops = 20000
			}
			log.Printf("setting IOPs value to %s; value must be <= 50 times size", iops)
		}
		cvi.Iops = aws.Int64(iops)
	}

	rsp, err := mounter.EC2.CreateVolume(cvi)
	if err != nil {
		return nil, errors.Wrap(err, "CreateVolume returned an error")
	}

	err = mounter.waitForVolumeStatus(*rsp.VolumeId, "available")
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func (mounter *EC2Mounter) attach(volumeID string) (*MountResponse, error) {
	var attached bool

	defer func() {
		if !attached {
			log.Println("unsuccessful EBS volume attachment, deleting volume")
			err := mounter.DetachAndDelete(volumeID)
			if err != nil {
				log.Println("error deleting volume:", err)
			}
		}
	}()

	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html
	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/volume_limits.html
	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html
	prefix := "/dev/sd"
	letters := strings.Split("fghijklmnopqrstuvwxyz", "")
	device := ""
	// start at a random position
	off := rand.Int63n(int64(len(letters)))
	// retry up to 10 times
	for k := int64(0); k < 10; k++ {
		if off+k > int64(len(letters)-1) {
			off = rand.Int63n(int64(len(letters)))
		}

		device = prefix + letters[off]
		if _, err := os.Stat(device); !os.IsNotExist(err) {
			continue
		}

		_, err := mounter.EC2.AttachVolume(&ec2.AttachVolumeInput{
			InstanceId: aws.String(mounter.IID.InstanceID),
			VolumeId:   aws.String(volumeID),
			Device:     aws.String(device),
		})
		if err != nil {
			// race condition attaching devices from multiple containers to the same host /dev address.
			// so retry with randomish wait time.
			log.Printf("retrying EBS attach because of difficulty getting volume. error was: %+T. %s", err, err)
			if strings.Contains(err.Error(), "is already in use") {
				time.Sleep(time.Duration(1000+rand.Int63n(1000)) * time.Millisecond)
				continue
			}
			return nil, fmt.Errorf("failed to attach volume %s: %v", volumeID, err)
		}

		err = mounter.waitForVolumeStatus(volumeID, "in-use")
		if err != nil {
			return nil, err
		}

		if !mounter.waitForDevice(device) {
			return nil, fmt.Errorf("error waiting for device %s to attach", device)
		}

		attached = true
		break
	}

	if !attached {
		return nil, fmt.Errorf("failed to find and attach device")
	}

	return &MountResponse{device, volumeID}, nil
}

func (mounter *EC2Mounter) waitForVolumeStatus(volumeId string, status string) error {
	var xstatus string
	for i := 0; i < 30; i++ {
		drsp, err := mounter.EC2.DescribeVolumes(
			&ec2.DescribeVolumesInput{
				VolumeIds: []*string{aws.String(volumeId)},
			})
		if err != nil {
			return errors.Wrapf(err, "DescribeVolumes returned an error")
		}
		if len(drsp.Volumes) == 0 {
			return fmt.Errorf("volume %s not found", volumeId)
		}
		xstatus = *drsp.Volumes[0].State
		if xstatus == status {
			return nil
		}
		time.Sleep(time.Duration(5000+rand.Int63n(5000)) * time.Millisecond)
	}
	return fmt.Errorf("volume %s never transitioned to status %s. last was: %s", volumeId, status, xstatus)
}

func (mounter *EC2Mounter) waitForDevice(device string) bool {
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(device); err != nil {
			time.Sleep(2 * time.Second)
		} else {
			return true
		}
	}
	return false
}

func (mounter *EC2Mounter) deleteOnTermination(volumeId string, device string) error {
	moi := &ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(mounter.IID.InstanceID),
		BlockDeviceMappings: []*ec2.InstanceBlockDeviceMappingSpecification{
			{
				DeviceName: aws.String(device),
				Ebs: &ec2.EbsInstanceBlockDeviceSpecification{
					DeleteOnTermination: aws.Bool(true),
					VolumeId:            aws.String(volumeId),
				},
			}},
	}
	_, err := mounter.EC2.ModifyInstanceAttribute(moi)
	if err != nil {
		return errors.Wrap(err, "error setting delete on termination")
	}
	return nil
}

func (mounter *EC2Mounter) createAttach(cli *MountRequest) (*MountResponse, error) {
	log.Println("creating EBS volume")
	createResp, err := mounter.create(cli.Size, cli.VolumeType, cli.Iops)
	if err != nil {
		return nil, errors.Wrap(err, "error creating volume")
	}
	log.Println("created EBS volume", *createResp.VolumeId)

	log.Println("attaching EBS volume")
	attachResp, err := mounter.attach(*createResp.VolumeId)
	if err != nil {
		return nil, errors.Wrap(err, "error attaching volume")
	}
	log.Println("attached EBS volume to", attachResp.Device)

	if !cli.Keep {
		log.Println("configuring EBS volume to delete on instance termination")
		err = mounter.deleteOnTermination(*createResp.VolumeId, attachResp.Device)
		if err != nil {
			return nil, errors.Wrap(err, "error setting delete on termination")
		}
	}

	return attachResp, nil
}

func (mounter *EC2Mounter) mountLocal(dev string, mountPoint string) error {
	if _, err := os.Stat(dev); err != nil {
		return errors.Wrap(err, "device does not appear to be attached")
	}

	log.Printf("making fs for %s", dev)
	cmd := exec.Command("mkfs", "-t", "ext4", dev)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mkfs failed: %s; %v", string(out), err)
	}

	log.Printf("mounting %s to %s", dev, mountPoint)
	if _, err = os.Stat(mountPoint); os.IsNotExist(err) {
		err = os.MkdirAll(mountPoint, os.FileMode(0777))
		if err != nil {
			return err
		}
	}

	cmd = exec.Command("mount", "-o", "noatime", dev, mountPoint)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %s; %v", string(out), err)
	}

	return nil
}

func (mounter *EC2Mounter) CreateAndMount(args *MountRequest) (*MountResponse, error) {
	// Create and mount the EBS volume
	vol, err := mounter.createAttach(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create and attach an EBS volume")
	}

	// Mount the newly created volume
	err = mounter.mountLocal(vol.Device, args.MountPoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to mount EBS volume")
	}
	if args.VolumeType == "st1" || args.VolumeType == "sc1" {
		// https://aws.amazon.com/blogs/aws/amazon-ebs-update-new-cold-storage-and-throughput-options/
		cmd := exec.Command("blockdev", "--setra", "2048", vol.Device)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("warning: error setting read-ahead: %s, %v\n", out, err)
		}
	}

	log.Printf("mounted EBS volume %s on device %s to %s\n", vol.VolumeID, vol.Device, args.MountPoint)
	return vol, nil
}

func (mounter *EC2Mounter) DetachAndDelete(volumeID string) error {
	log.Printf("detaching volume %s from instance %s", volumeID, mounter.IID.InstanceID)
	for i := 0; i < 10; i++ {
		v, err := mounter.EC2.DetachVolume(&ec2.DetachVolumeInput{
			VolumeId: aws.String(volumeID),
			Force:    aws.Bool(true),
		})
		if err == nil {
			werr := mounter.waitForVolumeStatus(volumeID, "available")
			if werr != nil {
				return werr
			}
			break
		}
		if strings.Contains(err.Error(), "is in the 'available' state") {
			break
		}
		if v != nil && *v.State == "available" {
			break
		}
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	log.Printf("deleting volume %s", volumeID)
	_, err := mounter.EC2.DeleteVolume(&ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeID),
	})
	if err != nil {
		return errors.Wrap(err, "delete volume request returned an error")
	}

	return nil
}

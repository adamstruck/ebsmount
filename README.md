# EBSMount

ebsmount is a set of utilities for managing EBS volumes on a running EC2 instance. It was inspired by the similar functionality in https://github.com/base2genomics/batchit.

### Install

`go get github.com/adamstruck/ebsmount`

# Usage 

## Mount a EBS volume

Create, attach and mount an EBS volume.

```
$ ebsmount mount -h
Mount an EBS volume to an EC2 instance.

Usage:
  ebsmount mount [flags]

Flags:
  -t, --fs-type string       file system type to create (argument must be accepted by mkfs) (default "ext4")
  -h, --help                 help for mount
  -i, --iops int             Provisioned IOPS. Only valid for volume type io1. Range is 100 to 20000 and <= 50*size of volume
  -k, --keep                 don't delete the volume on termination (default is to delete)
  -m, --mount-point string   directory on which to mount the EBS volume
  -s, --size int             size in GB of desired EBS volume (default 200)
  -v, --volume-type string   desired volume type; gp2 for General Purpose SSD; io1 for Provisioned IOPS SSD; st1 for Throughput Optimized HDD; sc1 for HDD or Magnetic volumes; standard for infrequent (default "gp2")
```

## Unmount a volume

Unmount, detach and delete an EBS volume. 

```
$ ebsmount unmount -h
Unmount an EBS volume from an EC2 instance.

Usage:
  ebsmount unmount [flags]

Flags:
  -h, --help                 help for unmount
  -m, --mount-point string   directory to unmount
  -v, --volume-id string     EBS volume ID to detach and/or delete from instance
```

## Server

Run as a service on a local unix socket.

```
$ ebsmount server -h
Start ebsmount as a service.

Usage:
  ebsmount server [flags]

Flags:
  -h, --help            help for server
  -s, --socket string   unix socket (default "./ebsmount.sock")
```

## API


### `POST /mount`


#### Parameters
| Name       | Type      | Description                          |
|------------|-----------|--------------------------------------|
| Size       | Int32     |  Size in GB of desired EBS volume
| MountPoint | String    |  Directory on which to mount the EBS volume
| VolumeType | String    | Desired volume type; gp2 for General Purpose SSD; io1 for Provisioned IOPS SSD; st1 for Throughput Optimized HDD; sc1 for HDD or Magnetic volumes; standard for infrequent
| FSType     | String    | File system type to create (argument must be accepted by mkfs)
| Iops       | Int32     | Provisioned IOPS. Only valid for volume type io1. Range is 100 to 20000 and <= 50*size of volume


CURL example:

```
curl -d '{"Size": 100, "MountPoint": "/home/ec2-user/mnt", "VolumeType": "gp2", "FSType": "ext4"}' --unix-socket /var/run/ebsmount.sock http://localhost/mount
```

Success-Response (example):

```
HTTP/1.1 200 OK

{
    "Device":"/dev/sdq",
    "VolumeID":"vol-0b725f1904fed492d"
}
```


### `POST /unmount`


#### Parameters
| Name       | Type   | Description                          |
|------------|--------|--------------------------------------|
| VolumeID   | String | EBS volume ID
| MountPoint | String | Directory to unmount


CURL example:

```
curl -d '{"VolumeID": "vol-03378474e88e5fd04", "MountPoint": "/home/ec2-user/mnt"}' --unix-socket /var/run/ebsmount.sock http://localhost/unmount
```

Success-Response (example):

```
HTTP/1.1 200 OK

{}
```

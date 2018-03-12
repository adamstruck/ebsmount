#!/bin/sh

set -e
set -o pipefail

# check required env vars are set
[ -z "$DISK" ] && echo "Need to set DISK"
[ -z "$TASKID" ] && echo "Need to set TASKID"
([ -z "$TASKID" ] || [ -z "$TASKID" ]) && exit 1 

set -o xtrace

echo "mounting $DISK GB EBS volume to /mnt/$TASKID..."
curl -d "{\"Size\": $DISK, \"MountPoint\": \"/mnt/$TASKID\", \"VolumeType\": \"gp2\", \"FSType\": \"ext4\"}" \
     --unix-socket /var/run/ebsmount.sock \
     http://localhost/mount \
     > vid

echo $(cat vid)
VID=$(cat vid | jq -r .VolumeID)
[ -z "$VID" ] && echo "VID was not set... exiting" && exit 1

echo "starting funnel worker..."
docker run -i \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v /mnt:/mnt \
       docker.io/ohsucompbio/funnel:kf-dev worker \
       run \
       --Worker.WorkDir /mnt \
       --Database dynamodb \
       --DynamoDB.Region us-east-1 \
       --DynamoDB.TableBasename funnel \
       --Worker.PollingRate 30s \
       --Worker.LogUpdateRate 10m \
       --Worker.LogTailSize 10000 \
       --taskID $TASKID

echo "unmounting and deleting EBS volume $VID..."
curl -d "{\"VolumeID\": \"$VID\", \"MountPoint\": \"/mnt/$TASKID\"}" \
     --unix-socket /var/run/ebsmount.sock \
     http://localhost/unmount

FROM centos:7.4.1708
RUN yum install e4fsprogs
ADD ebsmount /usr/local/bin/ebsmount
ENTRYPOINT ["ebsmount"]

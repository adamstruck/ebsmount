FROM centos:7.4.1708
RUN yum install -y e4fsprogs curl docker
RUN systemctl enable docker
ADD ebsmount /usr/local/bin/ebsmount

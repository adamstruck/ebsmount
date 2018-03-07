FROM docker:stable
ADD ebsmount /usr/local/bin/ebsmount
ENTRYPOINT ["ebsmount"]

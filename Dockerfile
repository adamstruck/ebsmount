FROM alpine
RUN apk add --update --no-cache curl jq docker
ADD ebsmount /usr/local/bin/ebsmount
# add a custom script that will make requests to a ebsmount server
# ADD ./mountAndStartWorker.sh /opt/mountAndStartWorker.sh
CMD ["sh"]

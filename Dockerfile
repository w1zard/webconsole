FROM alpine:3.3

MAINTAINER Eric Shi <postmaster@apibox.club>

ENV LANG en_US.UTF-8

RUN apk add --no-cache --update-cache go && mkdir -p /data/apibox 
ENV GOPATH /data/apibox

ADD . /data/apibox

RUN cd /data/apibox/src/apibox.club/apibox/ && go install

EXPOSE 8080

CMD /data/apibox/bin/apibox start
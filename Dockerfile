FROM alpine:3.3

MAINTAINER Eric Shi <postmaster@apibox.club>

ENV LANG en_US.UTF-8

RUN apk --update add tar git wget
RUN mkdir -p /data/tools && mkdir -p /data/apibox 
ENV GOPATH /data/apibox

RUN cd /data/tools && wget https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz
RUN cd /data/tools && tar -zxvf go1.6.linux-amd64.tar.gz -C /usr/local
ENV PATH /usr/local/go/bin:$PATH
ADD . /data/apibox

RUN cd /data/apibox/src/apibox.club/apibox/ && go install

EXPOSE 8080

CMD /data/apibox/bin/apibox start
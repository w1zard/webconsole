FROM alpine:3.3

MAINTAINER Eric Shi <postmaster@apibox.club>

ENV LANG en_US.UTF-8

RUN apk --update add tar git curl
RUN mkdir -p /data/tools && mkdir -p /data/apibox 
RUN cd /data/tools && curl -L 'http://www.golangtc.com/static/go/1.6/go1.6.linux-amd64.tar.gz' | tar -zx -C /usr/local

ENV PATH /usr/local/go/bin:$PATH

ADD . /data/apibox
ENV GOPATH /data/apibox
RUN cd /data/apibox/src/apibox.club/apibox/ && go install

EXPOSE 8080

CMD /data/apibox/bin/apibox start
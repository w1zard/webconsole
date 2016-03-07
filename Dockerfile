FROM ubuntu:15.10

MAINTAINER Eric Shi <postmaster@apibox.club>

ENV LANGUAGE en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8

RUN apt-get -yq update && apt-get -yq upgrade
RUN apt-get -yq install git curl


RUN mkdir -p /data/tools
RUN cd /data/tools && curl -L 'http://www.golangtc.com/static/go/1.6/go1.6.linux-amd64.tar.gz' | tar -zx -C /usr/local
ENV PATH /usr/local/go/bin:$PATH

RUN mkdir -p /data/apibox
ADD . /data/apibox
ENV GOPATH /data/apibox
RUN cd /data/apibox/src/apibox.club/apibox/ && go install


EXPOSE 8080

CMD /data/apibox/bin/apibox start 
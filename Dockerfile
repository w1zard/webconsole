FROM ubuntu:16.04

MAINTAINER Eric Shi <postmaster@apibox.club>

ENV LANGUAGE en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8

RUN apt -y update && apt -y upgrade && apt -y install curl
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install openssh-server pwgen
RUN mkdir -p /var/run/sshd && sed -i "s/UsePrivilegeSeparation.*/UsePrivilegeSeparation no/g" /etc/ssh/sshd_config && sed -i "s/UsePAM.*/UsePAM no/g" /etc/ssh/sshd_config && sed -i "s/PermitRootLogin.*/PermitRootLogin yes/g" /etc/ssh/sshd_config

RUN mkdir -p /data/tools && mkdir -p /data/apibox
RUN cd /data/tools && curl -L 'https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz' | tar -zx -C /usr/local
ENV PATH /usr/local/go/bin:$PATH

ADD . /data/apibox
ENV GOPATH /data/apibox
RUN cd /data/apibox/src/apibox.club/apibox/ && go install && sed -i "/exit 0/i /data/apibox/bin/apibox start" /etc/rc.local

ADD set_root_pw.sh /set_root_pw.sh
ADD run.sh /run.sh
RUN chmod +x /*.sh

ENV AUTHORIZED_KEYS **None**
ENV ROOT_PASS **RANDOM**

EXPOSE 22
EXPOSE 8080

CMD ["/run.sh"]
CMD /data/apibox/bin/apibox start

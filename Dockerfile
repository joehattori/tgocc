FROM ubuntu:18.04

RUN mkdir /tgocc && \
    apt-get update && \
    apt-get install -y make man vim less wget gcc

RUN rm /etc/dpkg/dpkg.cfg.d/excludes

RUN wget https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz

RUN tar -C /usr/local -xzf go1.14.4.linux-amd64.tar.gz

RUN rm go1.14.4.linux-amd64.tar.gz

ENV PATH $PATH:/usr/local/go/bin

WORKDIR /tgocc

ADD . .

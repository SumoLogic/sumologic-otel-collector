FROM ubuntu:20.04

ARG DEBIAN_VERSION="bullseye"
ARG UBUNTU_VERSION="20.04"
ARG GO_VERSION=1.21.11

RUN apt-get update
RUN apt-get install -yq tzdata
ENV TZ="UTC"

RUN apt-get install -y git wget xz-utils build-essential curl texinfo file

RUN curl -Lo go.tar.gz https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
RUN tar -zxvf go.tar.gz -C /usr/local
ENV PATH="/usr/local/go/bin:${PATH}"

COPY config.mak /config.mak
COPY Makefile /Makefile

RUN make toolchain-linux_arm64
RUN make toolchain-linux_amd64

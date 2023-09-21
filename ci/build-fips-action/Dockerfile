FROM amazonlinux:2

ARG TARGETARCH

ENV TARGETARCH=$TARGETARCH

RUN yum groupinstall -y "Development Tools" && yum install -y curl git

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

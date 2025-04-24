FROM amazonlinux:2

ARG TARGETARCH

ENV TARGETARCH=$TARGETARCH

RUN yum groupinstall -y "Development Tools" && yum install -y curl git

RUN curl -Lo go.tar.gz https://go.dev/dl/go1.24.0.linux-$TARGETARCH.tar.gz
RUN tar -zxvf go.tar.gz -C /usr/local
ENV PATH="/usr/local/go/bin:${PATH}"

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

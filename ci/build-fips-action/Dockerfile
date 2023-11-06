# build-cross builds a cross compiler targeting $TARGETARCH-linux-musl
FROM alpine:latest as build-cross-musl

RUN apk add alpine-sdk

WORKDIR /root

ENV MUSL_CC=https://more.musl.cc/11.2.1/x86_64-linux-musl/aarch64-linux-musl-cross.tgz
ENV MUSL_CC_SHA="dfdbdd592d24e5f013c282b693753d57d1307e8135f734fe25b83e0790ccc86acf0a52e0923de33390173f5fd9368474b9f61c2fbfc41d65d5867c9cea9efc35"
RUN wget $MUSL_CC -O musl-cross.tgz \
		&& echo "$MUSL_CC_SHA musl-cross.tgz" | sha512sum -c \
		&& tar xzf musl-cross.tgz -C /opt && rm musl-cross.tgz;

# golang alpine image ready to build sumologic-otel-collector
# includes native and cross CC toolchains ready for CGO.
FROM golang:1.20.10-alpine3.18 as build

RUN apk add alpine-sdk

COPY --from=build-cross-musl /opt/aarch64-linux-musl-cross /opt/cross/aarch64-linux-musl

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

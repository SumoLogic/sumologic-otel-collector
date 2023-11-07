# build-cross builds a cross compiler targeting $TARGETARCH-linux-musl
FROM alpine:latest as build-base
ARG PARALLEL=3

RUN apk add xz alpine-sdk

WORKDIR /opt

ENV MUSL_X_MAKE_REPO=https://github.com/richfelker/musl-cross-make.git
ENV MUSL_X_MAKE_REF=fe915821b652a7fa37b34a596f47d8e20bc72338
RUN git clone $MUSL_X_MAKE_REPO \
		&& cd musl-cross-make \
		&& git checkout $MUSL_X_MAKE_REF;

WORKDIR /opt/musl-cross-make

ADD config.mak /opt/musl-cross-make/config.mak

RUN make -j$PARALLEL extract_all;

FROM build-base as build-musl-aarch64
ARG PARALLEL=3

RUN make -j$PARALLEL TARGET=aarch64-linux-musl install \
		&& mv output /opt/aarch64-linux-musl

FROM build-base as build-musl-amd64
ARG PARALLEL=3

RUN make -j$PARALLEL TARGET=x86_64-linux-musl install \
		&& mv output /opt/x86_64-linux-musl


# golang alpine image ready to build sumologic-otel-collector
# includes native and cross CC toolchains ready for CGO.
FROM golang:1.20.10-alpine3.18 as build

RUN apk add alpine-sdk

ARG TARGETARCH=aarch64

COPY --from=build-musl-amd64	/opt/x86_64-linux-musl	/opt/x86_64-linux-musl
COPY --from=build-musl-aarch64  /opt/aarch64-linux-musl /opt/aarch64-linux-musl

WORKDIR /

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

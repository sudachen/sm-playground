FROM ubuntu:18.04 as linux
ENV DEBIAN_FRONTEND noninteractive
ENV SHELL /bin/bash
ARG TZ=US/Eastern
ENV TZ $TZ

USER root
RUN bash -c "for i in {1..9}; do mkdir -p /usr/share/man/man\$i; done" \
 && echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/90noninteractive \
 && echo 'DPkg::Options "--force-confnew";' >> /etc/apt/apt.conf.d/90noninteractive \
 && apt-get update --fix-missing \
 && apt-get install -qy --no-install-recommends \
    ca-certificates \
    tzdata \
    locales \
    git \
    bash \
    sudo \
    unzip \
    make \
    curl \
    procps \
    net-tools \
    apt-transport-https \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* \
 && locale-gen en_US.UTF-8 \
 && update-locale LANG=en_US.UTF-8 \
 && echo "$TZ" > /etc/timezone

ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

FROM linux as golang

ENV GOLANG_VERSION 1.16.5
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN set -ex \
 && apt-get update --fix-missing \
 && apt-get install -qy --no-install-recommends \
    gcc \
	libc6-dev \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* \
 && curl -L https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz | tar zx -C /usr/local \
 && go version \
 && mkdir -p "$GOPATH/src" "$GOPATH/bin" \
 && chmod -R 777 "$GOPATH"

FROM golang AS build_base
WORKDIR /go/src/github.com/spacemeshos/go-spacemesh

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

COPY r0/g0.mod go.mod
COPY r0/go.sum .
RUN go mod download

FROM build_base AS server_builder
WORKDIR /go/src/github.com/spacemeshos
COPY . .
RUN rm -rf build

WORKDIR /go/src/github.com/spacemeshos/r0
RUN make build-miner

FROM linux AS spacemesh
LABEL maintainer="Alexey Sudachen <alexey@spacemesh.io>"

RUN set -ex \
 && apt-get update --fix-missing \
 && apt-get install -qy --no-install-recommends \
    valgrind \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

COPY --from=server_builder /go/src/github.com/spacemeshos/r0/build/go-spacemesh /bin/
COPY --from=server_builder /go/src/github.com/spacemeshos/r0/build/libgpu-setup.so /bin/
RUN mkdir /massif
ENTRYPOINT ["/bin/go-spacemesh"]


FROM golang AS builder
WORKDIR /go/src/github.com/spacemeshos/go-spacemesh

ARG REV=.
ARG GO_SPACEMESH=$REV/go-spacemesh
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

COPY $GO_SPACEMESH/go.mod go.mod
COPY $GO_SPACEMESH/go.sum go.sum
RUN	go mod download
RUN GO111MODULE=off go get golang.org/x/lint/golint

WORKDIR /go/src/github.com/spacemeshos
COPY . .
WORKDIR /go/src/github.com/spacemeshos/$REV
RUN echo $REV && pwd && rm -rf build


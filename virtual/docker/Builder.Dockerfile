FROM golang AS builder
WORKDIR /go/src/github.com/spacemeshos/go-spacemesh

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

COPY r0/g0.mod go.mod
COPY r0/go.sum .
RUN go mod download

WORKDIR /go/src/github.com/spacemeshos
COPY . .
RUN rm -rf build


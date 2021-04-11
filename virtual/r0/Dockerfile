# Inspired by https://container-solutions.com/faster-builds-in-docker-with-go-1-11/
# Base build image
FROM golang:1.14.4-alpine3.12 AS build_base
RUN apk add bash make git curl unzip rsync libc6-compat gcc musl-dev lsof
WORKDIR /go/src/github.com/spacemeshos/go-spacemesh

# Force the go compiler to use modules
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

#We want to populate the module cache based on the go.{mod,sum} files.
COPY go-spacemesh/go.mod .
COPY r0/go.sum .
#RUN go mod tidy
RUN go mod download
RUN go get github.com/stretchr/testify@v1.7.0
RUN go get github.com/golang/snappy@v0.0.1

#WORKDIR /go/src/github.com/spacemeshos/go-spacemesh
#COPY setup_env.sh .
#COPY scripts/* scripts/
#RUN ./setup_env.sh

# This image builds the go-spacemesh server
FROM build_base AS server_builder
WORKDIR /go/src/github.com/spacemeshos
COPY . .
RUN rm -rf build

WORKDIR /go/src/github.com/spacemeshos/r0
RUN make build-miner

#In this last stage, we start from a fresh Alpine image, to reduce the image size and not ship the Go compiler in our production artifacts.
FROM alpine AS spacemesh
COPY --from=server_builder /go/src/github.com/spacemeshos/r0/build/go-spacemesh /bin/go-spacemesh
ENTRYPOINT ["/bin/go-spacemesh"]

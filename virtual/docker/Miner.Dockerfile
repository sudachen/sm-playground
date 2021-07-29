FROM builder as build
RUN make build-miner

FROM linux AS spacemesh
LABEL maintainer="Alexey Sudachen <alexey@spacemesh.io>"

RUN set -ex \
 && apt-get update --fix-missing \
 && apt-get install -qy --no-install-recommends \
    valgrind \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

ARG REV=.
COPY --from=build /go/src/github.com/spacemeshos/$REV/build/go-spacemesh /bin/
COPY --from=build /go/src/github.com/spacemeshos/$REV/build/libgpu-setup.so /bin/
RUN mkdir /massif
ENTRYPOINT ["/bin/go-spacemesh"]

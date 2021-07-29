FROM linux as golang

ENV GOLANG_MAJOR_VERSION 1
ENV GOLANG_MINOR_VERSION 15
ENV GOLANG_PATCH_VERSION 14
ENV GOLANG_VERSION $GOLANG_MAJOR_VERSION.$GOLANG_MINOR_VERSION.$GOLANG_PATCH_VERSION
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



FROM builder AS test
ENV CGO_LDFLAGS -Wl,-rpath,\$$ORIGIN/build
ENTRYPOINT ["/usr/bin/make"]

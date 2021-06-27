export VERSION = localnet
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=_ -X main.branch=_"
BINARY := go-spacemesh$(EXE)

include Makefile.Inc

build-miner: clean get-gpu-setup $(BIN_DIR)$(BINARY)

build/Test.Dockerfile: $(foreach X,Linux Golang Builder Test,docker/$(X).Dockerfile)
	mkdir -p $(dir $@)
	cat $^ > $@

docker-noapp-test: build/Test.Dockerfile
	docker build -t go-spacemesh-test -f build/Dockerfile.Test .
	docker run go-spacemesh-test no-app-test

docker-onlyapp-test: build/Test.Dockerfile
	docker build -t go-spacemesh-test -f build/Test.Dockerfile
	docker run go-spacemesh-test only-app-test

docker-test-all: docker-noapp-test docker-onlyapp-test

build/Dockerfile.Miner: $(foreach X,Linux Golang Builder Miner,docker/$(X).Dockerfile)
	mkdir -p $(dir $@)
	cat $^ > $@

docker-build-miner: clean build/Dockerfile.Miner
	docker build -t go-spacemesh:${VERSION} -f build/Dockerfile.Miner .
	docker tag go-spacemesh:${VERSION} local/go-spacemesh:latest

docker-build-poet:
	cd poet && docker build -t poet:${VERSION} .
	docker tag poet:${VERSION} local/poet:latest

clean:
	-rm -f $(BIN_DIR)/$(BINARY)

no-app-test:
	ulimit -n 9999; go test -v -timeout 0 -p 1 -tags exclude_app_test ./...

only-app-test:
	ulimit -n 9999; go test -timeout 0 -p 1 -v -tags !exclude_app_test ./cmd/node

export VERSION = localnet-$(REVISION)
GO_SPACEMESH ?= rev_$(REVISION)/go-spacemesh
BUILD_ARGS = --build-arg REV=rev_$(REVISION) --build-arg GO_SPACEMESH=$(GO_SPACEMESH)
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=_ -X main.branch=_"
BINARY := go-spacemesh$(EXE)
build-miner: install clean get-gpu-setup go-build
.PHONY: docker-noapp-test docker-onlyapp-test docker-test-all \
        docker-build-miner docker-build-poet \
        clean no-app-test only-app-test install go-build

include ../rev_0/Makefile.Inc

install:
	go mod download
	go run ../rev_0/check-go-version.go --major 1 --minor 15
	GO111MODULE=off go get golang.org/x/lint/golint

go-build: $(BIN_DIR)$(BINARY)
$(BIN_DIR)$(BINARY):
	mkdir -p $(dir $@)
	go build ${LDFLAGS} -o $(BIN_DIR)$(BINARY) $(GOTAGS) .

build/Test.Dockerfile: $(foreach X,Linux Golang Builder Test,docker/$(X).Dockerfile)
	mkdir -p $(dir $@)
	cat $^ > $@

docker-noapp-test: build/Test.Dockerfile
	docker build $(BUILD_ARGS) -t go-spacemesh-test -f build/Dockerfile.Test ..
	docker run go-spacemesh-test no-app-test

docker-onlyapp-test: build/Test.Dockerfile
	docker build $(BUILD_ARGS) -t go-spacemesh-test -f build/Test.Dockerfile ..
	docker run go-spacemesh-test only-app-test

docker-test-all: docker-noapp-test docker-onlyapp-test

build/Dockerfile.Miner: $(foreach X,Linux Golang Builder Miner,../docker/$(X).Dockerfile)
	mkdir -p $(dir $@)
	cat $^ > $@

docker-build-miner: clean build/Dockerfile.Miner
	docker build $(BUILD_ARGS) -t go-spacemesh:${VERSION} -f build/Dockerfile.Miner ..
	docker tag go-spacemesh:${VERSION} local/go-spacemesh:${REVISION}

docker-build-poet:
	cd poet && docker build -t poet:${VERSION} .
	docker tag poet:${VERSION} local/poet:${REVISION}

clean:
	-rm -f $(BIN_DIR)/$(BINARY)

no-app-test:
	ulimit -n 9999; go test -v -timeout 0 -p 1 -tags exclude_app_test ./...

only-app-test:
	ulimit -n 9999; go test -timeout 0 -p 1 -v -tags !exclude_app_test ./cmd/node

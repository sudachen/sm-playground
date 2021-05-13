
export VERSION = localnet
BINARY := go-spacemesh
BIN_DIR := build
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=_ -X main.branch=_"

docker-build-miner:
	-rm -f $(BIN_DIR)/$(BINARY)
	cd .. && docker build -t go-spacemesh:${VERSION} -f r0/Dockerfile .
	docker tag go-spacemesh:${VERSION} local/go-spacemesh:latest

docker-build-poet:
	cd ../poet && docker build -t poet:${VERSION} .
	docker tag poet:${VERSION} local/poet:latest

build-miner: $(BIN_DIR)/$(BINARY)

$(BIN_DIR)/$(BINARY):
	go build ${LDFLAGS} -o $(BIN_DIR)/$(BINARY)

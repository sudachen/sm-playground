LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=_ -X main.branch=_"
BIN_DIR := build
PROJ_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
export CGO_LDFLAGS := -L$(PROJ_DIR)build -Wl,-rpath,\$$ORIGIN $(CGO_LDFLAGS)

$(info "OS=$(OS)")
ifeq ($(OS),Windows_NT)
  platform := windows
  export PATH := $(PATH):$(PROJ_DIR)/build
  EXE := .exe
else
  OS := $(shell uname)
  $(info "uname OS=$(OS)")
  ifeq ($(OS),Darwin)
	  platform := macos
  else
      platform := linux
  endif
endif

GPU_SETUP_REV = 0.1.20
GPU_SETUP_ZIP = libgpu-setup-$(platform)-$(GPU_SETUP_REV).zip
GPU_SETUP_URL_ZIP = https://github.com/spacemeshos/gpu-post/releases/download/v$(GPU_SETUP_REV)/$(GPU_SETUP_ZIP)
ifeq ($(platform), windows)
  GPU_SETUP_LIBS = gpu-setup.dll
else
  ifeq ($(platform), macos)
  	GPU_SETUP_LIBS = libgpu-setup.dylib libMoltenVK.dylib libvulkan.1.dylib MoltenVK_icd.json
  else
  	GPU_SETUP_LIBS = libgpu-setup.so
  endif
endif

BINDIR_GPU_SETUP_LIBS = $(foreach X,$(GPU_SETUP_LIBS),$(BIN_DIR)/$(X))
$(BINDIR_GPU_SETUP_LIBS):
	mkdir -p $(dir $@)
	cd $(dir $@) && unzip $(GPU_SETUP_ZIP) $(notdir $@)
$(BIN_DIR)/$(GPU_SETUP_ZIP):
	mkdir -p $(dir $@)
	curl -L $(GPU_SETUP_URL_ZIP) -o $(BIN_DIR)/$(GPU_SETUP_ZIP)
$(BIN_DIR)/$(BINARY):
	mkdir -p $(dir $@)
	go build ${LDFLAGS} -o $(BIN_DIR)/$(BINARY)

get-gpu-setup: $(BIN_DIR)/$(GPU_SETUP_ZIP) $(BINDIR_GPU_SETUP_LIBS)

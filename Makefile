################################################################################
# Variables
################################################################################
export GO111MODULE ?= on
export GOPROXY ?= https://proxy.golang.org
export GOSUMDB ?= sum.golang.org
# By default, disable CGO_ENABLED. See the details on https://golang.org/cmd/cgo
CGO         ?= 0
BINARIES ?= filter poolmanager
FILTER_BINARY ?= filter
POOLMANAGER_BINARY ?= poolmanager


################################################################################
# Git info
################################################################################
GIT_COMMIT  = $(shell git rev-list -1 HEAD)
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty)

################################################################################
# Release version
################################################################################
LASTEST_VERSION_TAG ?=

ifdef REL_VERSION
	POOL_VERSION := $(REL_VERSION)
	DOCKER_TAG := $(REL_VERSION)
	DOCKER_LATEST_TAG := latest
else
	POOL_VERSION := edge
	DOCKER_TAG := edge
	DOCKER_LATEST_TAG := edge
endif

FILTER_IMAGE_NAME := cs-agentpool-filter
POOLMANAGER_IMAGE_NAME := cs-agentpool-manager

################################################################################
# Architectue
################################################################################
LOCAL_ARCH := $(shell uname -m)
ifeq ($(LOCAL_ARCH),x86_64)
	TARGET_ARCH_LOCAL=amd64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),armv8)
	TARGET_ARCH_LOCAL=arm64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 4),armv)
	TARGET_ARCH_LOCAL=arm
else
	TARGET_ARCH_LOCAL=amd64
endif
export GOARCH ?= $(TARGET_ARCH_LOCAL)

################################################################################
# OS
################################################################################
LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
   TARGET_OS_LOCAL = linux
else ifeq ($(LOCAL_OS),Darwin)
   TARGET_OS_LOCAL = darwin
else
   TARGET_OS_LOCAL ?= windows
endif
export GOOS ?= $(TARGET_OS_LOCAL)

################################################################################
# Binaries extension
################################################################################
ifeq ($(GOOS),windows)
BINARY_EXT_LOCAL:=.exe
GOLANGCI_LINT:=golangci-lint.exe
else
BINARY_EXT_LOCAL:=
GOLANGCI_LINT:=golangci-lint
endif

export BINARY_EXT ?= $(BINARY_EXT_LOCAL)

################################################################################
# GO build flags
################################################################################
BASE_PACKAGE_NAME := github.com/AndreasM009/cloudshipper-agent-pool

DEFAULT_LDFLAGS := -X $(BASE_PACKAGE_NAME)/pkg/version.commit=$(GIT_VERSION) -X $(BASE_PACKAGE_NAME)/pkg/version.version=$(AGENT_VERSION)
ifeq ($(DEBUG),)
  BUILDTYPE_DIR:=release
  LDFLAGS:="$(DEFAULT_LDFLAGS) -s -w"
else ifeq ($(DEBUG),0)
  BUILDTYPE_DIR:=release
  LDFLAGS:="$(DEFAULT_LDFLAGS) -s -w"
else
  BUILDTYPE_DIR:=debug
  GCFLAGS:=-gcflags="all=-N -l"
  LDFLAGS:="$(DEFAULT_LDFLAGS)"
  $(info Build with debugger information)
endif

################################################################################
# output directory
################################################################################
OUT_DIR := ./dist
POOL_OUT_DIR := $(OUT_DIR)/$(GOOS)_$(GOARCH)/$(BUILDTYPE_DIR)
POOL_LINUX_OUT_DIR := $(OUT_DIR)/linux_$(GOARCH)/$(BUILDTYPE_DIR)

################################################################################
# Target: build-all                                                               
################################################################################
.PHONY: build-all
POOL_BINS:=$(foreach ITEM,$(BINARIES),$(POOL_OUT_DIR)/$(ITEM)$(BINARY_EXT))
build-all: $(POOL_BINS)

# Generate builds for agent binaries for the target
# Params:
# $(1): the binary name for the target
# $(2): the binary main directory
# $(3): the target os
# $(4): the target arch
# $(5): the output directory
define genBinariesForTarget
.PHONY: $(5)/$(1)
$(5)/$(1):
	CGO_ENABLED=$(CGO) GOOS=$(3) GOARCH=$(4) go build $(GCFLAGS) -ldflags=$(LDFLAGS) \
	-o $(5)/$(1) \
	$(2)/main.go;
endef

# Generate binary targets
$(foreach ITEM,$(BINARIES),$(eval $(call genBinariesForTarget,$(ITEM)$(BINARY_EXT),./cmd/$(ITEM),$(GOOS),$(GOARCH),$(POOL_OUT_DIR))))

################################################################################
# Target: build-filter                                                     
################################################################################
.PHONY: build-filter
FILTER_BIN_EXT:=$(FILTER_BINARY)$(BINARY_EXT)
build-filter:
	CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GCFLAGS) -ldflags=$(LDFLAGS) \
	-o $(POOL_OUT_DIR)/$(FILTER_BIN_EXT) ./cmd/$(FILTER_BINARY)/main.go

################################################################################
# Target: build-poolmanager                                                          
################################################################################
.PHONY: build-poolmanager
POOLMANAGER_BIN_EXT:=$(POOLMANAGER_BINARY)$(BINARY_EXT)
build-poolmanager:
	CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GCFLAGS) -ldflags=$(LDFLAGS) \
	-o $(POOL_OUT_DIR)/$(POOLMANAGER_BIN_EXT) ./cmd/$(POOLMANAGER_BINARY)/main.go

################################################################################
# Target: lint                                                                
################################################################################
.PHONY: lint	
lint:
	$(GOLANGCI_LINT) run --fix

################################################################################
# Target: test
################################################################################
.PHONY: test
test:
	go test ./pkg/...

################################################################################
# Target: docker-build-...
################################################################################
.PHONY: docker-build-filter docker-build-poolmanager docker-build-all

docker-build-filter:
	docker build -t $(FILTER_IMAGE_NAME):$(DOCKER_TAG) -f ./Docker/filter/Dockerfile .

docker-build-poolmanager:
	docker build -t $(POOLMANAGER_IMAGE_NAME):$(DOCKER_TAG) -f ./Docker/poolmanager/Dockerfile .

docker-build-all: docker-build-filter docker-build-poolmanager

################################################################################
# Target: docker-publish
################################################################################
check-docker-publish-args:
ifeq ($(s),)
	$(error docker server must be set: s=<dockerserver>)
endif
ifeq ($(u),)
	$(error docker login must be set: u=<dockerusername>)
endif
ifeq ($(p),)
	$(error docker password must be set: p=<dockerpassword>)
endif

.PHONY: docker-publish-filter
docker-publish-filter: check-docker-publish-args
	docker login -p $(p) -u $(u)
	docker build -t $(s)/$(FILTER_IMAGE_NAME):$(DOCKER_TAG) -f ./Docker/filter/Dockerfile .
	docker tag $(s)/$(FILTER_IMAGE_NAME):$(DOCKER_TAG) $(s)/$(FILTER_IMAGE_NAME):$(DOCKER_LATEST_TAG)
	docker push $(s)/$(FILTER_IMAGE_NAME):$(DOCKER_TAG)
	docker push $(s)/$(FILTER_IMAGE_NAME):$(DOCKER_LATEST_TAG)

.PHONY: docker-publish-poolmanager
docker-publish-poolmanager: check-docker-publish-args
	docker login -p $(p) -u $(u)
	docker build -t $(s)/$(POOLMANAGER_IMAGE_NAME):$(DOCKER_TAG) -f ./Docker/poolmanager/Dockerfile .
	docker tag $(s)/$(POOLMANAGER_IMAGE_NAME):$(DOCKER_TAG) $(s)/$(POOLMANAGER_IMAGE_NAME):$(DOCKER_LATEST_TAG)
	docker push $(s)/$(POOLMANAGER_IMAGE_NAME):$(DOCKER_TAG)
	docker push $(s)/$(POOLMANAGER_IMAGE_NAME):$(DOCKER_LATEST_TAG)

.PHONY: docker-publish-all
docker-publish-all: docker-publish-poolmanager docker-publish-filter

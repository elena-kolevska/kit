# ------------------------------------------------------------
# Copyright 2021 The Dapr Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ------------------------------------------------------------

################################################################################
# Variables                                                                    #
################################################################################

export GO111MODULE ?= on
export GOPROXY ?= https://proxy.golang.org
export GOSUMDB ?= sum.golang.org

GIT_COMMIT  = $(shell git rev-list -1 HEAD)
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty)
# By default, disable CGO_ENABLED. See the details on https://golang.org/cmd/cgo
CGO         ?= 0

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

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
   TARGET_OS_LOCAL = linux
else ifeq ($(LOCAL_OS),Darwin)
   TARGET_OS_LOCAL = darwin
else
   TARGET_OS_LOCAL ?= windows
endif
export GOOS ?= $(TARGET_OS_LOCAL)

ifeq ($(GOOS),windows)
BINARY_EXT_LOCAL:=.exe
GOLANGCI_LINT:=golangci-lint.exe
# Workaround for https://github.com/golang/go/issues/40795
BUILDMODE:=-buildmode=exe
else
BINARY_EXT_LOCAL:=
GOLANGCI_LINT:=golangci-lint
endif

PROTOC_GEN_GO_VERSION = v1.28.1
PROTOC_GEN_GO_NAME+= $(PROTOC_GEN_GO_VERSION)

PROTOC_GEN_GO_GRPC_VERSION = 1.2.0  

################################################################################
# Target: test                                                                 #
################################################################################
.PHONY: test
test:
	go test ./... $(COVERAGE_OPTS) $(BUILDMODE)

################################################################################
# Target: lint                                                                 #
################################################################################
.PHONY: lint
lint:
	# Due to https://github.com/golangci/golangci-lint/issues/580, we need to add --fix for windows
	$(GOLANGCI_LINT) run --timeout=20m

################################################################################
# Target: go.mod                                                               #
################################################################################
.PHONY: go.mod
go.mod:
	go mod tidy

################################################################################
# Target: check-diff                                                           #
################################################################################
.PHONY: check-diff
check-diff:
	git diff --exit-code ./go.mod # check no changes

################################################################################
# Target: init-proto                                                            #
################################################################################
.PHONY: init-proto
init-proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)

################################################################################
# Target: gen-proto                                                            #
################################################################################
GRPC_PROTOS:=customerrors
PROTO_PREFIX:=./pkg

# Generate archive files for each binary
# $(1): the binary name to be archived
define genProtoc
.PHONY: gen-proto-$(1)
gen-proto-$(1):
	$(PROTOC) protoc --go_out=./pkg  --go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false,module=$(PROTO_PREFIX) ./proto/$(1)/v1/*.proto
endef

$(foreach ITEM,$(GRPC_PROTOS),$(eval $(call genProtoc,$(ITEM))))

GEN_PROTOS:=$(foreach ITEM,$(GRPC_PROTOS),gen-proto-$(ITEM))

.PHONY: gen-proto
gen-proto: check-proto-version $(GEN_PROTOS) modtidy

################################################################################
# Target: check-proto-version                                                         #
################################################################################
.PHONY: check-proto-version
check-proto-version: ## Checking the version of proto related tools
	echo "checking...done";

################################################################################
# Target: check-proto-diff                                                           #
################################################################################
.PHONY: check-proto-diff
check-proto-diff:
	git diff --exit-code ./pkg/proto/customerrors/v1/customerrors.pb.go # check no changes

################################################################################
# Target: modtidy                                                              #
################################################################################
.PHONY: modtidy
modtidy:
	go mod tidy
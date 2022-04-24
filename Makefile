# This how we want to name the binary output
#
MAINVERSION=$(shell cat version)

GOPATH ?= $(shell go env GOPATH)
# Ensure GOPATH is set before running build process.
ifeq "$(GOPATH)" ""
  $(error Please set the environment variable GOPATH before running `make`)
endif
PATH := ${GOPATH}/bin:$(PATH)
GCFLAGS=-gcflags "all=-trimpath=${GOPATH}"

GITTAG := $(shell git describe --tags --always)
GITSHA := $(shell git rev-parse --short HEAD)
GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILDTIME=`date +%FT%T%z`

LDFLAGS=-ldflags "-X main.MainVersion=${MAINVERSION} -X main.GitSha=${GITSHA} -X main.GitTag=${GITTAG} -X main.GitBranch=${GITBRANCH} -X main.BuildTime=${BUILDTIME} -s -w"

# colors compatible setting
CRED:=$(shell tput setaf 1 2>/dev/null)
CGREEN:=$(shell tput setaf 2 2>/dev/null)
CYELLOW:=$(shell tput setaf 3 2>/dev/null)
CEND:=$(shell tput sgr0 2>/dev/null)

.PHONY: all
all: | fmt build

.PHONY: go_version_check
GO_VERSION_MIN=1.17
# Parse out the x.y or x.y.z version and output a single value x*10000+y*100+z (e.g., 1.9 is 10900)
# that allows the three components to be checked in a single comparison.
VER_TO_INT:=awk '{split(substr($$0, match ($$0, /[0-9\.]+/)), a, "."); print a[1]*10000+a[2]*100+a[3]}'
go_version_check:
	@echo "$(CGREEN)=> Go version check ...$(CEND)"
	@if test $(shell go version | $(VER_TO_INT) ) -lt \
  	$(shell echo "$(GO_VERSION_MIN)" | $(VER_TO_INT)); \
  	then printf "go version $(GO_VERSION_MIN)+ required, found: "; go version; exit 1; \
		else echo "go version check pass";	fi

# Code format
.PHONY: fmt
fmt: go_version_check
	@echo "$(CGREEN)=> Run gofmt on all source files ...$(CEND)"
	@echo "gofmt -l -s -w ..."
	@ret=0 && for d in $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		gofmt -l -s -w $$d/*.go || ret=$$? ; \
	done ; exit $$ret

# Run golang test cases
.PHONY: test
test:
	@echo "$(CGREEN)=> Run all test cases ...$(CEND)"
	@go test $(LDFLAGS) -timeout 10m -race ./...
	@echo "$(CGREEN)=> Test Success!$(CEND)"

# Compile
.PHONY: compile	
compile:
	@echo "$(CGREEN)=> Compile protobuf ...$(CEND)"
	@bash build/compile.sh


# Builds the project
.PHONY: build
clear:
	@rm -rf examples/omega*
	@rm -rf packages/omega*

# Build all
.PHONY: build
build: build_o build_c build_h build_w build_l

.PHONY: build_o
build_o: fmt
# build omega
	@mkdir -p examples/omega/bin
	@mkdir -p examples/omega/etc
	@cp etc/omega.conf examples/omega/etc/omega.conf
	@echo "$(CGREEN)=> Building binary(omega)...$(CEND)"
	go build -race ${LDFLAGS} ${GCFLAGS} -o examples/omega/bin/omega cmd/omega/main.go
	@echo "$(CGREEN)=> Build Success!$(CEND)"

.PHONY: build_w
build_w: fmt
# build omega-watchdog
	@mkdir -p examples/omega/bin
	@mkdir -p examples/omega/etc
	@cp etc/omega.conf examples/omega/etc/omega.conf
	@echo "$(CGREEN)=> Building binary(omega-watchdog)...$(CEND)"
	go build -race ${LDFLAGS} ${GCFLAGS} -o examples/omega/bin/omega-watchdog cmd/omega-watchdog/main.go
	@echo "$(CGREEN)=> Build Success!$(CEND)"

.PHONY: build_c
build_c: fmt
# build omega-collector
	@mkdir -p examples/omega-collector/bin
	@mkdir -p examples/omega-collector/etc
	@cp etc/omega-collector.conf examples/omega-collector/etc/omega-collector.conf
	@echo "$(CGREEN)=> Building binary(omega-collector)...$(CEND)"
	go build -race ${LDFLAGS} ${GCFLAGS} -o examples/omega-collector/bin/omega-collector cmd/omega-collector/main.go
	@echo "$(CGREEN)=> Build Success!$(CEND)"

.PHONY: build_h
build_h: fmt
# build omega-hub
	@mkdir -p examples/omega-hub/bin
	@mkdir -p examples/omega-hub/etc
	@cp etc/omega-hub.conf examples/omega-hub/etc/omega-hub.conf
	@echo "$(CGREEN)=> Building binary(omega-hub)...$(CEND)"
	go build -race ${LDFLAGS} ${GCFLAGS} -o examples/omega-hub/bin/omega-hub cmd/omega-hub/main.go
	@echo "$(CGREEN)=> Build Success!$(CEND)"

.PHONY: build_l
build_l: fmt
# build omega-ctl
	@mkdir -p examples/omega-ctl/bin
	@mkdir -p examples/omega-ctl/image
	@echo "$(CGREEN)=> Building binary(omega-ctl)...$(CEND)"
	go build -race ${LDFLAGS} ${GCFLAGS} -o examples/omega-ctl/bin/omega-ctl cmd/omega-ctl/main.go
	@echo "$(CGREEN)=> Build Success!$(CEND)"

# Package all
.PHONY: package
package: package_o

.PHONY: package_o
package_o: fmt
	@mkdir -p packages/omega
	@mkdir -p packages/omega/bin
	@mkdir -p packages/omega/etc
	@cp build/omega-install.sh packages/
	@echo "$(CGREEN)=> Packaging binary(omega-watchdog)...$(CEND)"
	go build -race ${LDFLAGS} ${GCFLAGS} -o packages/omega/bin/omega cmd/omega/main.go
	go build -race ${LDFLAGS} ${GCFLAGS} -o packages/omega/bin/omega-watchdog cmd/omega-watchdog/main.go
	@echo "$(CGREEN)=> Package Success!$(CEND)"
	@cd packages; tar -zcvf omega.tar.gz omega > /dev/null
	@rm -rf packages/omega
	@mkdir -p examples/omega-ctl/image
	@cp build/omega-install.sh examples/omega-ctl/image
	@cp build/resource.txt examples/omega-ctl/image
	@cp packages/omega.tar.gz examples/omega-ctl/image

.PHONY: mod
mod:export GO111MODULE=on
mod:
	@echo "$(CGREEN)=> go mod tidy...$(CEND)"
	@go mod tidy

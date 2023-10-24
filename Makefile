# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

GO    := GO111MODULE=on go
PROMU := $(shell $(GO) env GOPATH)/bin/promu
GOLINTER                ?= $(shell $(GO) env GOPATH)/bin/gometalinter
pkgs   = $(shell $(GO) list ./...)
# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le
DOCKER_REPO  ?= prometheuscommunity

include Makefile.common

DOCKER_IMAGE_NAME       ?= elasticsearch-exporter
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))


all: format build test

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -name '*.go' -print) | grep '^'

test:
	@echo ">> running tests"
	@$(GO) test -short $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

build: promu
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

crossbuild: promu
	@echo ">> cross-building binaries"
	@$(PROMU) crossbuild

tarball: promu
	@echo ">> building release tarball"
	@$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

tarballs: promu
	@echo ">> building release tarballs"
	@$(PROMU) crossbuild tarballs
	@echo ">> calculating release checksums"
	@$(PROMU) checksum $(BIN_DIR)/.tarballs

docker:
	@echo ">> building docker image"
	@docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

promu:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
	        GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
	        $(GO) get -u github.com/prometheus/promu

gometalinter: $(GOLINTER)
	@echo ">> linting code"
	@$(GOLINTER) --install > /dev/null
	@$(GOLINTER) --config=./.gometalinter.json ./...

$(GOPATH)/bin/gometalinter lint:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get -u github.com/alecthomas/gometalinter

.PHONY: all style format build test vet tarball docker promu $(GOPATH)/bin/gometalinter lint
=======
>>>>>>> upstream/master

SHELL = /bin/bash

VERSION  	= $(shell git describe --always --tags --dirty)
VPP_VERSION	= $(shell dpkg-query -f '\${Version}' -W vpp)

VPP_IMG 	?= ligato/vpp-base:latest
BINAPI_DIR	?= ./examples/binapi
GENBINAPI_CMDS = $(shell go generate -n ./examples/binapi 2>&1 | tr "\n" ";")

GO = GO111MODULE=on go

all: test build examples

install:
	@echo "=> installing binapi-generator ${VERSION}"
	$(GO) install ./cmd/binapi-generator

build:
	@echo "=> building binapi-generator ${VERSION}"
	cd cmd/binapi-generator && $(GO) build -v

examples:
	@echo "=> building examples"
	cd examples/simple-client && $(GO) build -v
	cd examples/stats-api && $(GO) build -v
	cd examples/perf-bench && $(GO) build -v
	cd examples/union-example && $(GO) build -v

test:
	@echo "=> running tests"
	$(GO) test -v ./cmd/...
	$(GO) test -v ./ ./adapter ./core ./api ./codec

test-cover:
	@echo "=> running tests with coverage"
	$(GO) test -cover ./cmd/...
	$(GO) test -cover ./ ./adapter ./core ./api ./codec

extras:
	@echo "=> building extras"
	cd extras/libmemif/examples/gopacket && $(GO) build -v
	cd extras/libmemif/examples/icmp-responder && $(GO) build -v
	cd extras/libmemif/examples/jumbo-frames && $(GO) build -v
	cd extras/libmemif/examples/raw-data && $(GO) build -v

clean:
	@echo "=> cleaning"
	rm -f cmd/binapi-generator/binapi-generator
	rm -f examples/perf-bench/perf-bench
	rm -f examples/simple-client/simple-client
	rm -f examples/stats-api/stats-api
	rm -f examples/union-example/union-example
	rm -f extras/libmemif/examples/gopacket/gopacket
	rm -f extras/libmemif/examples/icmp-responder/icmp-responder
	rm -f extras/libmemif/examples/jumbo-frames/jumbo-frames
	rm -f extras/libmemif/examples/raw-data/raw-data

gen-binapi-docker: install
	@echo "=> generating binapi VPP $(VPP_VERSION)"
	docker run -t --rm \
		-v "$(shell which gofmt):/usr/local/bin/gofmt:ro" \
		-v "$(shell which binapi-generator):/usr/local/bin/binapi-generator:ro" \
		-v "$(shell pwd):/govpp" -w /govpp \
		-u "$(shell id -u):$(shell id -g)" \
		"${VPP_IMG}" \
	  sh -xc "cd $(BINAPI_DIR) && $(GENBINAPI_CMDS)"

generate-binapi: install
	@echo "=> generating binapi VPP $(VPP_VERSION)"
	$(GO) generate -x "$(BINAPI_DIR)"

generate:
	@echo "=> generating code"
	$(GO) generate -x ./...

update-vppapi:
	@echo "=> updating API JSON files using installed VPP ${VPP_VERSION}"
	@cd ${BINAPI_DIR} && find . -type f -name '*.api.json' -exec cp /usr/share/vpp/api/'{}' '{}' \;
	@echo ${VPP_VERSION} > ${BINAPI_DIR}/VPP_VERSION

lint:
	@echo "=> running linter"
	@golint ./... | grep -v vendor | grep -v /binapi/ || true

.PHONY: all \
	install build examples test \
	extras clean lint \
	generate generate-binapi gen-binapi-docker

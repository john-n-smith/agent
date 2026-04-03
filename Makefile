SHELL := /bin/bash
.DEFAULT_GOAL := help
VERSION ?=
DIST_DIR := dist

.PHONY: help build dist require-release-version clean-dist fmt test

help:
	@printf "Available targets:\n"
	@printf "  make build                      Build the harbour binary for macOS ARM64\n"
	@printf "  make dist VERSION=vX.Y.Z        Build Darwin release archives and checksums\n"
	@printf "  make clean-dist                 Remove release artefacts\n"
	@printf "  make fmt                        Format the Go source\n"
	@printf "  make test                       Run the Go tests\n"

build:
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=dev" -o bin/harbour ./cmd/harbour

dist: require-release-version clean-dist $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64.tar.gz $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64.tar.gz $(DIST_DIR)/sha256sums.txt

require-release-version:
	@test -n "$(VERSION)" || (echo "VERSION is required for make dist, e.g. make dist VERSION=v0.1.0" >&2; exit 1)

$(DIST_DIR)/harbour-$(VERSION)-darwin-amd64.tar.gz:
	mkdir -p $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64/harbour ./cmd/harbour
	tar -C $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64 -czf $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64.tar.gz harbour
	rm -rf $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64

$(DIST_DIR)/harbour-$(VERSION)-darwin-arm64.tar.gz:
	mkdir -p $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64/harbour ./cmd/harbour
	tar -C $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64 -czf $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64.tar.gz harbour
	rm -rf $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64

$(DIST_DIR)/sha256sums.txt: $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64.tar.gz $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64.tar.gz
	cd $(DIST_DIR) && shasum -a 256 harbour-$(VERSION)-darwin-amd64.tar.gz harbour-$(VERSION)-darwin-arm64.tar.gz > sha256sums.txt

clean-dist:
	rm -rf $(DIST_DIR)

fmt:
	gofmt -w ./cmd/harbour/*.go

test:
	go test ./...

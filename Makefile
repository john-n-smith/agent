SHELL := /bin/bash
.DEFAULT_GOAL := help
VERSION ?=
DIST_DIR := dist
LDFLAGS := -X main.version=$(VERSION)
DARWIN_AMD64_ARCHIVE := $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64.tar.gz
DARWIN_ARM64_ARCHIVE := $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64.tar.gz

.PHONY: help build release require-release-version clean-dist fmt test

help:
	@printf "Available targets:\n"
	@printf "  make build                      Build the harbour binary for the current platform\n"
	@printf "  make release VERSION=vX.Y.Z     Build Darwin release archives and checksums\n"
	@printf "  make clean-dist                 Remove release artefacts\n"
	@printf "  make fmt                        Format the Go source\n"
	@printf "  make test                       Run the Go tests\n"

build:
	mkdir -p bin
	go build -ldflags "-X main.version=dev" -o bin/harbour ./cmd/harbour

release: require-release-version clean-dist $(DARWIN_AMD64_ARCHIVE) $(DARWIN_ARM64_ARCHIVE) $(DIST_DIR)/sha256sums.txt

require-release-version:
	@test -n "$(VERSION)" || (echo "VERSION is required for make release, e.g. make release VERSION=v0.1.0" >&2; exit 1)

$(DARWIN_AMD64_ARCHIVE):
	mkdir -p $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64/harbour ./cmd/harbour
	tar -C $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64 -czf $(DARWIN_AMD64_ARCHIVE) harbour
	rm -rf $(DIST_DIR)/harbour-$(VERSION)-darwin-amd64

$(DARWIN_ARM64_ARCHIVE):
	mkdir -p $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64/harbour ./cmd/harbour
	tar -C $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64 -czf $(DARWIN_ARM64_ARCHIVE) harbour
	rm -rf $(DIST_DIR)/harbour-$(VERSION)-darwin-arm64

$(DIST_DIR)/sha256sums.txt: $(DARWIN_AMD64_ARCHIVE) $(DARWIN_ARM64_ARCHIVE)
	cd $(DIST_DIR) && shasum -a 256 $(notdir $(DARWIN_AMD64_ARCHIVE)) $(notdir $(DARWIN_ARM64_ARCHIVE)) > sha256sums.txt

clean-dist:
	rm -rf $(DIST_DIR)

fmt:
	gofmt -w ./cmd/harbour/*.go

test:
	go test ./...

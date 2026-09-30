.DEFAULT_GOAL := build

.PHONY: \
	build-ui \
	ui-dev \
	fmt \
	vet \
	test \
	build \
	build-go \
	build-linux-amd64 \
	build-linux-arm64 \
	build-windows-amd64 \
	build-all \
	clean \
	clean-ui \
	clean-all


build-ui:
	@echo "Building UI..."
	npm --prefix web/ui install
	npm --prefix web/ui run build

ui-dev:
	@echo "Running UI dev..."
	npm --prefix web/ui install
	npm --prefix web/ui run dev



fmt:
	@echo "Formatting Go code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

test:
	@echo "Running tests..."
	go test ./...



VERSION ?= dev
OUTPUT_DIR := bin

ifeq ($(OS),Windows_NT)
	BINARY_EXT := .exe
else
	BINARY_EXT :=
endif



build:
	@echo "Building for current platform..."
	go build \
		-tags "sqlite_foreign_keys" \
		-o $(OUTPUT_DIR)/mizu$(BINARY_EXT) \
		./cmd/...


# Build backend only for current platform
build-go:
	@echo "Building backend for current platform..."
	go build \
		-tags "sqlite_foreign_keys" \
		-o $(OUTPUT_DIR)/mizu$(BINARY_EXT) \
		./cmd/...



build-linux-amd64:
	@echo "Building Linux amd64..."
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=amd64 \
	go build \
		-tags "sqlite_foreign_keys" \
		-o $(OUTPUT_DIR)/mizu-$(VERSION)-linux-amd64 \
		./cmd/...


build-linux-arm64:
	@echo "Building Linux arm64..."
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=arm64 \
	CC=aarch64-linux-gnu-gcc \
	go build \
		-tags "sqlite_foreign_keys" \
		-o $(OUTPUT_DIR)/mizu-$(VERSION)-linux-arm64 \
		./cmd/...


build-windows-amd64:
	@echo "Building Windows amd64..."
	CGO_ENABLED=1 \
	GOOS=windows \
	GOARCH=amd64 \
	CC=x86_64-w64-mingw32-gcc \
	go build \
		-tags "sqlite_foreign_keys" \
		-o $(OUTPUT_DIR)/mizu-$(VERSION)-windows-amd64.exe \
		./cmd/...

build-all: \
	build-ui \
	build-linux-amd64 \
	build-linux-arm64 \
	build-windows-amd64

	@echo "All builds completed."



clean:
	@echo "Cleaning binaries..."
	rm -rf $(OUTPUT_DIR)

clean-ui:
	@echo "Cleaning UI build files..."
	rm -rf web/ui/.svelte-kit/
	rm -rf web/ui/build/

clean-all: clean clean-ui

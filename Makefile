.DEFAULT_GOAL := build

.PHONY: ui ui-dev fmt vet build build-windows-on-linux clean

ui:
	@echo "Building UI..."
	npm --prefix web/ui install
	npm --prefix web/ui run build

ui-dev:
	@echo "Running UI dev..."
	npm --prefix web/ui install
	npm --prefix web/ui run dev

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run `go vet` after formatting
vet: fmt
	@echo "Running go vet..."
	go vet ./...


OUTPUT = bin/mizu

ifeq ($(OS),Windows_NT)
	OUTPUT := $(OUTPUT).exe
endif


# Build for current OS
build: ui vet
	@echo "Building for current platform..."
	go build -tags "sqlite_foreign_keys" -o $(OUTPUT) ./cmd/...

# Build backend for current OS
build-go:
	@echo "Building for current platform..."
	go build -tags "sqlite_foreign_keys" -o $(OUTPUT) ./cmd/...

# Cross-Compile for Windows
build-xwin: ui vet
	@echo "Cross-Compiling for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
	go build -o $(OUTPUT).exe ./cmd/...

# Cross-Compile backend for Windows
build-xwin-go:
	@echo "Cross-Compiling for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
	go build -o $(OUTPUT).exe ./cmd/...

# Clean built binaries
clean:
	@echo "Cleaning binaries..."
	rm -rf bin/

# Clean built ui
clean-ui:
	@echo "Cleaning ui build files..."
	rm -rf web/ui/.svelte-kit/
	rm -rf web/ui/build/

clean-all: clean clean-ui
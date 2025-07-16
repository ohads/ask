.PHONY: build build-all clean release

# Build for current platform
build:
	go build -o ask main.go

# Build for all platforms
build-all: clean
	GOOS=darwin GOARCH=amd64 go build -o ask-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o ask-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -o ask-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o ask-linux-arm64 main.go

# Clean build artifacts
clean:
	rm -f ask ask-darwin-amd64 ask-darwin-arm64 ask-linux-amd64 ask-linux-arm64

# Calculate SHA256 hashes for Homebrew formula
sha256:
	@echo "SHA256 hashes for Homebrew formula:"
	@echo "Darwin AMD64:"
	@shasum -a 256 ask-darwin-amd64
	@echo "Darwin ARM64:"
	@shasum -a 256 ask-darwin-arm64
	@echo "Linux AMD64:"
	@shasum -a 256 ask-linux-amd64
	@echo "Linux ARM64:"
	@shasum -a 256 ask-linux-arm64

# Install locally for testing
install: build
	cp ask /usr/local/bin/

# Uninstall
uninstall:
	rm -f /usr/local/bin/ask 
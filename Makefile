APP_NAME := nats-client
VERSION := 0.0.1
BUILD_DIR := build

# GOOS and GOARCH for different platforms
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	darwin/amd64 \
	darwin/arm64

# Flags for building
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: all clean build

all: clean build

# Clean up build directory
clean:
	rm -rf $(BUILD_DIR)

# Build for all platforms
build: $(PLATFORMS)

$(PLATFORMS):
	@GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
	mkdir -p $(BUILD_DIR)/$@ && \
	go build $(LDFLAGS) -o $(BUILD_DIR)/$@/$(APP_NAME) .

# Cross-compiling for specific OS/architecture
linux: GOOS=linux
linux: GOARCH=amd64
linux: build

windows: GOOS=windows
windows: GOARCH=amd64
windows: build

darwin: GOOS=darwin
darwin: GOARCH=amd64
darwin: build

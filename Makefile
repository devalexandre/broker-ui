APP_NAME := nats.client.com
VERSION := 0.0.1
BUILD_DIR := build



.PHONY: all clean build $(PLATFORMS)

all: clean build

# Clean up build directory
clean:
	rm -rf $(BUILD_DIR)

# Build for all platforms
build: windows linux darwin


# Build for all platforms using fyne-cross
windows:
	fyne-cross windows -arch=amd64,386  -app-id $(APP_NAME) -app-version $(VERSION)
linux:
	fyne-cross linux -arch=amd64,arm64 -app-id $(APP_NAME) -app-version $(VERSION)
darwin:
	fyne-cross darwin -arch=amd64,arm64 -app-id $(APP_NAME) -app-version $(VERSION)

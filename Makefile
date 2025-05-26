.PHONY: build run

BINARY_NAME=atcli
BUILD_DIR=build

build:
	@if [ -z "$(word 2,$(MAKECMDGOALS))" ] || [ -z "$(word 3,$(MAKECMDGOALS))" ]; then \
		echo "Usage: make build <arch> <os>"; \
		echo "Example: make build arm64 linux"; \
		exit 1; \
	fi; \
	ARCH=$(word 2,$(MAKECMDGOALS)); \
	OS=$(word 3,$(MAKECMDGOALS)); \
	OUTPUT="$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH"; \
	mkdir -p $(BUILD_DIR); \
	echo "Building for $$OS/$$ARCH..."; \
	GOOS=$$OS GOARCH=$$ARCH go build -o $$OUTPUT ./src; \
	if [ $$? -eq 0 ]; then \
		echo "Binary created: $$OUTPUT"; \
	else \
		echo "Build failed"; \
		exit 1; \
	fi; \
	$(eval MAKECMDGOALS := build)

# Swallow extra args so make doesn't treat them as targets
%:
	@:

run:
	go run .

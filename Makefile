.PHONY: build test lint fmt clean install
.DEFAULT_GOAL := build

# Installation directory
INSTALL_DIR ?= /usr/local/bin

build:
	go build -o bin/mindful cmd/mindful/main.go

test:
	go test ./tests/...

test-unit:
	go test ./tests/unit/...

test-integration:
	go test ./tests/integration/...

test-contract:
	go test ./tests/contract/...

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

clean:
	rm -rf bin/

install: build
	@echo "Installing mindful to $(INSTALL_DIR)..."
	sudo cp bin/mindful $(INSTALL_DIR)/mindful
	sudo chmod +x $(INSTALL_DIR)/mindful
	@echo "mindful installed successfully to $(INSTALL_DIR)/mindful"
	@echo "You can now run 'mindful' from anywhere"

install-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
.PHONY: dev build test lint vet clean swagger run help

APP_BIN := bin/youtube-watcher
MAIN_PATH := ./cmd/api/

## dev: start hot-reload development server (air)
dev:
	air

## run: build and run the server
run: build
	$(APP_BIN)

## build: compile binary
build:
	go build -o $(APP_BIN) $(MAIN_PATH)

## test: run all tests
test:
	go test ./... -v

## lint: run golangci-lint (if installed)
lint:
	golangci-lint run ./...

## vet: run go vet
vet:
	go vet ./...

## swagger: regenerate Swagger docs
swagger:
	swag init -g $(MAIN_PATH)main.go -o docs --parseInternal

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -f $(APP_BIN)
	rm -rf tmp/

## deps: install all dependencies
deps:
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest

## help: show this help
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## //'

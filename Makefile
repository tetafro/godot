VERSION := $(shell git describe --tags)

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -ldflags="-X 'main.version=$(VERSION)'" -o godot ./cmd/godot

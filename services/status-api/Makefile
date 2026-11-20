VERSION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
IMAGE   ?= ghcr.io/echeekene/status-api

.PHONY: test lint build run image

test:
	go test -race -count=1 ./...

lint:
	gofmt -l . | (! grep .)
	go vet ./...

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o bin/status-api ./cmd/status-api
	CGO_ENABLED=0 go build -o bin/statusctl ./cmd/statusctl

run:
	go run ./cmd/status-api

image:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE):$(VERSION) .

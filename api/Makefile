.PHONY: setup build build-http build-sync build-cli build-bot test lint migrate-up migrate-down migrate-create sync rerender bot

GO_LDFLAGS ?= -s -w
GO_BUILD_FLAGS ?= -trimpath -ldflags "$(GO_LDFLAGS)"

setup:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.12.2
	go mod download

build-http:
	go build $(GO_BUILD_FLAGS) -o bin/http ./cmd/http

build-sync:
	go build $(GO_BUILD_FLAGS) -o bin/sync ./cmd/sync

build-cli:
	go build $(GO_BUILD_FLAGS) -o bin/cli ./cmd/cli

build-bot:
	go build $(GO_BUILD_FLAGS) -o bin/bot ./cmd/bot

build: build-http build-sync build-cli build-bot

test:
	go test ./...

lint:
	golangci-lint run --config .golangci.yml ./...

migrate-up: build-cli
	./bin/cli migrations up

migrate-down: build-cli
	./bin/cli migrations down

migrate-create: build-cli
	./bin/cli migrations create --name=$(name)

sync: build-sync
	./bin/sync

bot: build-bot
	./bin/bot

rerender: build-cli
	./bin/cli rerender

export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
export DOCKER_HOST=unix://$(HOME)/.colima/docker.sock
ifneq (,$(wildcard ./.env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

.PHONY: install
install:
	mise install

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: format
format:
	golangci-lint fmt --enable gofumpt,goimports ./...

.PHONY: generate
generate:
	go generate ./...

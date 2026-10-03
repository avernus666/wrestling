SHELL := /bin/sh

.PHONY: dev test test-race vet build frontend docker-up docker-down

dev:
	./scripts/start.sh

test:
	cd server-go && go test ./...
	npm test -- --runInBand

test-race:
	cd server-go && go test -race ./...

vet:
	cd server-go && go vet ./...

build:
	cd server-go && go build -trimpath ./cmd/server
	npm run build

frontend:
	npm run build

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-clean:
	docker compose down -v

quality:
	cd server-go && gofmt -w . && go test -race ./... && go vet ./...

test-all: quality
	npm test -- --watchAll=false
	npm run build
	npm run test:e2e

.PHONY: setup dev build test css

setup:
	go mod download
	npm install
	npm run css:build

dev:
	go run ./cmd/server

build:
	npm run css:build
	go build -o pos-phoenix ./cmd/server

test:
	go test ./...

css:
	npm run css:watch

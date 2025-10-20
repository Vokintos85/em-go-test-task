SHELL := /bin/bash

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable
GOENV := GOTOOLCHAIN=local

.PHONY: build run test tidy migrate docker-up docker-down

build:
	$(GOENV) go build ./...

run:
	$(GOENV) go run ./cmd/server

test:
	$(GOENV) go test ./...

tidy:
	$(GOENV) go mod tidy

migrate:
	psql "$(DATABASE_URL)" -f migrations/0001_init.sql

docker-up:
	docker compose up --build

docker-down:
	docker compose down

SHELL := /bin/bash

export PROJECT = mat-api

.DEFAULT_GOAL := help

build: ## build admin executable file
	go build -o bin/main ./cmd/mat-admin/


up: ## Builds, (re)creates, starts, and attaches to Docker containers for a service.
	docker-compose up

down: ## Stops Docker containers and removes containers, networks, volumes, and images created by up.
	docker-compose down

stop-all:
	docker stop $(docker ps -aq)

remove-all:
	docker rm $(docker ps -aq)

migrate: ## Migrate attempts to bring the schema for db up to date with the migrations defined.
	go run ./cmd/mat-admin/ migrate

seed: migrate ## Seed runs the set of seed-data queries against db. The queries are ran in a transaction and rolled back if any fail.
	go run ./cmd/mat-admin/ seed

keys: ## Generate private key file to private.pem file
	go run ./cmd/mat-admin/ keygen private.pem

tidy:
	go mod tidy
	go mod vendor

lint: ## go code linter with revive
	revive ./...

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -t -d -v ./...

deps-cleancache:
	go clean -modcache

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

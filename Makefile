export PROJECT = mat-api

.DEFAULT_GOAL := help

keys: ## Generate private key file to private.pem file
	go run ./cmd/mat-admin/ keygen private.pem

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

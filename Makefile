SHELL := /bin/bash
.DEFAULT_GOAL := help

WEB_DIR := web
API_DIR := services/api
PNPM := pnpm

.PHONY: help setup dev web api web-install api-install clean \
        dev-web dev-api \
        test test-v test-api test-api-handlers

help: ## Show commands
	@awk 'BEGIN{FS=":.*##"} /^[a-zA-Z_-]+:.*##/{printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: web-install api-install ## Install deps for web + api

dev: ## Run web + api together (Ctrl+C to stop)
	@$(MAKE) -j2 dev-web dev-api

# --- Web ---
web-install: ## pnpm install (web)
	@cd $(WEB_DIR) && $(PNPM) install

dev-web: ## pnpm run dev (web)
	@cd $(WEB_DIR) && $(PNPM) run dev

web: dev-web ## Alias

# --- API ---
api-install: ## go mod download (api)
	@cd $(API_DIR) && go mod download

dev-api: ## ./run-dev.sh (api)
	@cd $(API_DIR) && ./run-dev.sh

api: dev-api ## Alias

# --- Tests ---
test: test-api-handlers ## Run unit tests (default: handlers)

test-v: ## Run handlers unit tests (verbose, no cache)
	cd $(API_DIR) && go test -v -count=1 ./internal/handlers/...

test-api: ## Run all Go tests under services/api
	@cd $(API_DIR) && go test ./...

test-api-handlers: ## Run Go unit tests for handlers only
	@cd $(API_DIR) && go test ./internal/handlers/...

clean: ## Remove node_modules (web) and Go build cache
	@rm -rf $(WEB_DIR)/node_modules
	@cd $(API_DIR) && go clean -cache -testcache -modcache
	@echo "✅ cleaned"
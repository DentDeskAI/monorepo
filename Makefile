SHELL := /bin/sh

.DEFAULT_GOAL := help

ENV_FILE := .env
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export
endif

API_PORT ?= 8082
WEB_PORT ?= 3000
POSTGRES_USER ?= dentdesk

DC := docker compose

.PHONY: \
	help \
	up up-build down restart ps logs \
	api-logs worker-logs web-logs db-logs \
	db db-shell \
	rebuild clean \
	run-api run-worker \
	fmt test vet tidy \
	web-install web-dev web-build \
	docs

help: ## Show available commands
	@printf "\nDentDesk developer commands\n\n"
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\nURLs\n"
	@printf "  WEB:     http://localhost:%s\n" "$(WEB_PORT)"
	@printf "  API:     http://localhost:%s\n" "$(API_PORT)"
	@printf "  Swagger: http://localhost:%s/docs/index.html\n\n" "$(API_PORT)"

up: ## Start all services (detached)
	@$(DC) up -d
	@$(MAKE) --no-print-directory ps

up-build: ## Start all services and force image build
	@$(DC) up -d --build
	@$(MAKE) --no-print-directory ps

down: ## Stop and remove services
	@$(DC) down

restart: ## Restart all services
	@$(DC) restart
	@$(MAKE) --no-print-directory ps

ps: ## Show compose service status
	@$(DC) ps

logs: ## Tail logs for all services
	@$(DC) logs -f --tail=200

api-logs: ## Tail API logs
	@$(DC) logs -f --tail=200 api

worker-logs: ## Tail worker logs
	@$(DC) logs -f --tail=200 worker

web-logs: ## Tail web logs
	@$(DC) logs -f --tail=200 web

db-logs: ## Tail postgres logs
	@$(DC) logs -f --tail=200 postgres

db: db-shell ## Alias for db-shell

db-shell: ## Open psql inside postgres container
	@$(DC) exec postgres psql -U "$(POSTGRES_USER)" -d "$${POSTGRES_DB:-dentdesk}"

rebuild: ## Rebuild all images without cache
	@$(DC) build --no-cache

clean: ## Stop services and remove volumes
	@$(DC) down -v

run-api: ## Run API locally (without Docker)
	@go run ./cmd/api

run-worker: ## Run worker locally (without Docker)
	@go run ./cmd/worker

fmt: ## Format Go code
	@go fmt ./...

test: ## Run Go tests
	@go test ./...

vet: ## Run go vet checks
	@go vet ./...

tidy: ## Tidy Go modules
	@go mod tidy

web-install: ## Install frontend deps
	@cd web && npm install

web-dev: ## Run frontend in dev mode
	@cd web && npm install && npm run dev

web-build: ## Build frontend production bundle
	@cd web && npm install && npm run build

docs: ## Regenerate Swagger docs (requires swag CLI)
	@command -v swag >/dev/null 2>&1 || { echo "swag is not installed. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; exit 1; }
	@swag init -g cmd/api/main.go -o docs

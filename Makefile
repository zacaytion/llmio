.DEFAULT_GOAL := help

.PHONY: help up down logs logs-postgres logs-redis logs-pgadmin logs-mailpit clean-volumes \
        build build-server build-migrate run-server run-migrate server migrate install tidy \
        test coverage-view lint lint-fix fmt \
        clean clean-go-build clean-go-test clean-go-mod clean-go-fuzz clean-go-all

##@ Containers

up: ## Start all services
	podman compose up -d

down: ## Stop all services
	podman compose down

logs: ## Tail logs for all services
	podman compose logs -f

logs-postgres: ## Tail PostgreSQL logs
	podman compose logs -f postgres

logs-redis: ## Tail Redis logs
	podman compose logs -f redis

logs-pgadmin: ## Tail PgAdmin logs
	podman compose logs -f pgadmin

logs-mailpit: ## Tail Mailpit logs
	podman compose logs -f mailpit

clean-volumes: down ## Stop services and delete volumes
	podman volume rm llmio_postgres_data llmio_pgadmin_data 2>/dev/null || true

##@ Build

build-server: ## Build server binary
	go build -o bin/server ./cmd/server

build-migrate: ## Build migrate binary
	go build -o bin/migrate ./cmd/migrate

build: build-server build-migrate ## Build all binaries

##@ Run

run-server: ## Run server via go run (dev mode)
	go run ./cmd/server $(ARGS)

run-migrate: ## Run migrations via go run (dev mode)
	go run ./cmd/migrate $(ARGS)

server: bin/server ## Run server binary
	./bin/server $(ARGS)

migrate: bin/migrate ## Run migrate binary
	./bin/migrate $(ARGS)

bin/server: $(shell find . -name '*.go' -not -path './vendor/*')
	go build -o bin/server ./cmd/server

bin/migrate: $(shell find . -name '*.go' -not -path './vendor/*')
	go build -o bin/migrate ./cmd/migrate

##@ Dependencies

install: ## Download Go dependencies
	go mod download

tidy: ## Tidy Go modules
	go mod tidy

##@ Testing

.var/coverage:
	mkdir -p .var/coverage

test: .var/coverage ## Run tests with coverage
	go test -coverprofile=.var/coverage/coverage.out ./...

.var/coverage/coverage.out: test

coverage-view: .var/coverage/coverage.out ## View coverage report in browser
	go tool cover -html=.var/coverage/coverage.out

##@ Quality

.var/log:
	mkdir -p .var/log

lint: .var/log ## Run linter
	golangci-lint run ./... 2>&1 | tee .var/log/golangci-lint.log

lint-fix: ## Run linter with auto-fix
	golangci-lint run ./... --fix

fmt: ## Format code
	gofmt -w .
	goimports -w -local github.com/zacaytion/llmio .

##@ Cleanup

clean-go-build: ## Clean Go build cache
	go clean -cache

clean-go-test: ## Clean Go test cache
	go clean -testcache

clean-go-mod: ## Clean Go module cache (requires sudo on some systems)
	go clean -modcache

clean-go-fuzz: ## Clean Go fuzz cache
	go clean -fuzzcache

clean-go-all: ## Clean all Go caches
	go clean -cache -testcache -fuzzcache

clean: down ## Clean everything (volumes, caches, binaries, artifacts)
	@echo "Removing Docker volumes..."
	podman volume rm llmio_postgres_data llmio_pgadmin_data 2>/dev/null || true
	@echo "Removing binaries..."
	rm -rf bin/server bin/migrate
	@echo "Removing test/lint artifacts..."
	rm -rf .var/coverage .var/log
	@echo "Cleaning Go build and test caches..."
	go clean -cache -testcache
	@echo "Done."

##@ Help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

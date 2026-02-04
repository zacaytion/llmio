.DEFAULT_GOAL := help
SHELL := /bin/bash
.SHELLFLAGS := -o pipefail -c

# Volume names (prefixed with project directory name by Podman Compose)
VOLUMES := llmio_postgres_data llmio_pgadmin_data

# Go source files for dependency tracking
GO_SRC := $(shell find . -name '*.go' -not -path './vendor/*')

# Docker test image configuration
GIT_SHA    := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD | sed 's/\//-/g')
IMAGE_NAME := ghcr.io/zacaytion/llmio-pg-tap

.PHONY: all help up down logs clean-volumes \
        build build-server build-migrate run-server run-migrate server migrate install tidy \
        test test-unit test-pgtap test-integration test-all coverage-view psql lint lint-fix lint-files lint-md lint-makefile lint-migrations lint-all fmt \
        clean clean-go-build clean-go-test clean-go-mod clean-go-fuzz clean-go-all \
        docker-test-build docker-test-tag docker-test-push docker-test-image

# Default target (required by checkmake)
all: build

##@ Containers

.env:
	@echo "Error: .env file not found. Run: cp .env.example .env" && exit 1

up: .env ## Start all services
	podman compose up -d

down: ## Stop all services
	podman compose down

logs: ## Tail logs for all services
	podman compose logs -f

# Pattern rule for service-specific logs (usage: make logs-postgres, logs-redis, etc.)
logs-%: ## Tail logs for specific service (postgres, redis, pgadmin, mailpit)
	podman compose logs -f $*

clean-volumes: down ## Stop services and delete volumes
	@for vol in $(VOLUMES); do \
		if podman volume exists $$vol 2>/dev/null; then \
			echo "Removing volume: $$vol"; \
			podman volume rm $$vol || echo "Warning: Failed to remove $$vol" >&2; \
		else \
			echo "Volume $$vol does not exist (already clean)"; \
		fi; \
	done

##@ Build

# File targets with dependency tracking (auto-rebuild when Go sources change)
bin/server: $(GO_SRC)
	go build -o $@ ./cmd/server

bin/migrate: $(GO_SRC)
	go build -o $@ ./cmd/migrate

# Phony targets delegate to file targets
build-server: bin/server ## Build server binary (auto-rebuilds if sources changed)
build-migrate: bin/migrate ## Build migrate binary (auto-rebuilds if sources changed)
build: build-server build-migrate ## Build all binaries

##@ Run

run-server: ## Run server directly via go run (no build step)
	go run ./cmd/server $(ARGS)

run-migrate: ## Run migrations directly via go run (no build step)
	go run ./cmd/migrate $(ARGS)

server: bin/server ## Run server binary (auto-rebuilds if needed)
	./bin/server $(ARGS)

migrate: bin/migrate ## Run migrate binary (auto-rebuilds if needed)
	./bin/migrate $(ARGS)

##@ Dependencies

install: ## Download Go dependencies
	go mod download

tidy: ## Tidy Go modules
	go mod tidy

node_modules: package.json pnpm-lock.yaml
	@pnpm install


##@ Testing

.var/coverage:
	@mkdir -p .var/coverage || (echo "ERROR: Cannot create .var/coverage directory" >&2; exit 1)

test-unit: .var/coverage ## Run unit tests (no containers, fast)
	go test -coverprofile=.var/coverage/coverage.out ./...

test-pgtap: ## Run pgTap schema validation tests (requires container)
	go test -v -tags=pgtap ./internal/db/...

test-integration: ## Run API/DB integration tests (requires container)
	go test -v -tags=integration ./...

test-all: test-pgtap test-integration ## Run all tests (pgTap first, then integration)

test: test-unit ## Alias for test-unit (default test target)

.var/coverage/coverage.out: test-unit

coverage-view: .var/coverage/coverage.out ## View coverage report in browser
	@test -s .var/coverage/coverage.out || (echo "ERROR: No coverage data. Run 'make test-unit' first." >&2; exit 1)
	go tool cover -html=.var/coverage/coverage.out

##@ Database

# Database connection parameters (from compose.yml defaults)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= loomio_development
DB_USER ?= postgres
DB_PASSWORD ?= postgres

psql: ## Connect to PostgreSQL via psql (usage: make psql or make psql DB_NAME=loomio_test)
	PGPASSWORD=$(DB_PASSWORD) psql-18 -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

##@ Quality

.var/log:
	@mkdir -p .var/log || (echo "ERROR: Cannot create .var/log directory" >&2; exit 1)

lint: .var/log ## Run Go linter (exit code preserved via pipefail)
	golangci-lint run --config .config/golangci.yml ./... 2>&1 | tee .var/log/golangci-lint.log

lint-fix: ## Run Go linter with auto-fix
	golangci-lint run --config .config/golangci.yml ./... --fix

lint-files: node_modules ## Lint file/directory naming conventions
	@node_modules/.bin/ls-lint -config .config/ls-lint.yml

lint-md: node_modules ## Lint markdown files
	@node_modules/.bin/markdownlint-cli2

lint-makefile: ## Lint Makefile
	go tool checkmake --config .config/checkmake.ini Makefile

lint-migrations: ## Lint SQL migrations for safety
	mise exec -- squawk --config .config/squawk.toml db/migrations/*.sql

lint-all: lint lint-files lint-md lint-makefile lint-migrations ## Run all linters

fmt: ## Format code
	gofmt -w . && goimports -w -local github.com/zacaytion/llmio .

##@ Cleanup

clean-go-build: ## Clean Go build cache
	go clean -cache

clean-go-test: ## Clean Go test cache
	go clean -testcache

clean-go-mod: ## Clean Go module cache (requires sudo on some systems)
	go clean -modcache

clean-go-fuzz: ## Clean Go fuzz cache
	go clean -fuzzcache

clean-go-all: ## Clean Go build, test, and fuzz caches (excludes module cache)
	go clean -cache -testcache -fuzzcache

clean: clean-volumes ## Clean everything (volumes, caches, binaries, artifacts)
	@echo "Removing binaries..."
	rm -rf bin/server bin/migrate
	@echo "Removing test/lint artifacts..."
	rm -rf .var/coverage .var/log
	@echo "Cleaning Go build and test caches..."
	go clean -cache -testcache
	@echo "Done."

##@ Docker Test Image

docker-test-build: ## Build llmio-pg-tap image locally
	docker build --load -t $(IMAGE_NAME):latest -f db/Dockerfile.pgtap db/

docker-test-tag: docker-test-build ## Tag image with branch and SHA
	docker tag $(IMAGE_NAME):latest $(IMAGE_NAME):br-$(GIT_BRANCH)
	docker tag $(IMAGE_NAME):latest $(IMAGE_NAME):$(GIT_SHA)

docker-test-push: docker-test-tag ## Push image to ghcr.io (requires auth)
	docker push $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):br-$(GIT_BRANCH)
	docker push $(IMAGE_NAME):$(GIT_SHA)

docker-test-image: docker-test-push ## Build, tag, and push image

##@ Help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9%-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

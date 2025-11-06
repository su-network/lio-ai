.PHONY: help build run run-bg stop logs deps test test-coverage fmt vet lint security clean db-reset all frontend-install frontend-dev frontend-build dev stop-all

# Root-level Makefile to manage the Go app in joles/

GO_DIR ?= joles
FRONTEND_DIR ?= frontend
BIN_NAME ?= lio-ai
BIN_PATH := $(GO_DIR)/$(BIN_NAME)

GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

VERSION ?= 0.1.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

PID_FILE := $(GO_DIR)/server.pid
LOG_FILE := logs/server.log
ROOT_DIR := $(CURDIR)

help: ## Display this help screen
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go gateway binary
	cd $(GO_DIR) && $(GOBUILD) -o $(BIN_NAME) ./cmd/server

run: build ## Run the gateway in the foreground
	set -a; [ -f $(ROOT_DIR)/.env ] && . $(ROOT_DIR)/.env; set +a; \
	$(BIN_PATH)

run-bg: build ## Run the gateway in the background (logs/server.log; PID in joles/server.pid)
	@mkdir -p logs
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "Server already running with PID $$(cat $(PID_FILE))"; \
	else \
		set -a; [ -f $(ROOT_DIR)/.env ] && . $(ROOT_DIR)/.env; set +a; \
		nohup $(BIN_PATH) >> $(ROOT_DIR)/$(LOG_FILE) 2>&1 & echo $$! > $(ROOT_DIR)/$(PID_FILE); \
		echo "Started server PID $$(cat $(PID_FILE)) (logging to $(LOG_FILE))"; \
	fi

stop: ## Stop the background gateway if running
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "Stopped PID $$PID"; \
		else \
			echo "No running process with PID $$PID"; \
		fi; \
		rm -f $(PID_FILE); \
	else \
		echo "No PID file found"; \
	fi

logs: ## Tail the server log
	@mkdir -p logs
	tail -f $(LOG_FILE)

deps: ## Download and verify Go dependencies
	cd $(GO_DIR) && $(GOMOD) tidy && $(GOMOD) verify

test: ## Run Go tests
	cd $(GO_DIR) && $(GOTEST) -v ./...

test-coverage: ## Run tests with coverage report
	cd $(GO_DIR) && $(GOTEST) -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=$(GO_DIR)/coverage.out -o $(GO_DIR)/coverage.html
	@echo "Coverage report: $(GO_DIR)/coverage.html"

fmt: ## Format Go code
	cd $(GO_DIR) && $(GOCMD) fmt ./...

vet: ## Run go vet
	cd $(GO_DIR) && $(GOCMD) vet ./...

lint: ## Run golangci-lint (if installed)
	cd $(GO_DIR) && golangci-lint run ./...

security: ## Run gosec (if installed)
	cd $(GO_DIR) && gosec ./...

clean: ## Clean build artifacts and temp files
	cd $(GO_DIR) && $(GOCLEAN)
	rm -f $(BIN_PATH) $(GO_DIR)/coverage.out $(GO_DIR)/coverage.html $(PID_FILE)

db-reset: ## Remove local SQLite DB (data/lio.db)
	rm -f data/lio.db
	@echo "Removed data/lio.db"

# Frontend commands
frontend-install: ## Install frontend dependencies
	cd $(FRONTEND_DIR) && npm install

frontend-dev: ## Run frontend development server in background
	@echo "Starting frontend dev server..."
	@cd $(FRONTEND_DIR) && npm run dev > /dev/null 2>&1 & echo $$! > $(ROOT_DIR)/frontend.pid
	@sleep 2
	@echo "Frontend started on http://localhost:3000 (PID: $$(cat $(ROOT_DIR)/frontend.pid))"

frontend-build: ## Build frontend for production
	cd $(FRONTEND_DIR) && npm run build

# Combined commands
dev: build frontend-dev ## Run both backend and frontend in development mode
	@echo "====================================="
	@echo "Starting Lio AI Development Environment"
	@echo "====================================="
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "✓ Backend already running with PID $$(cat $(PID_FILE))"; \
	else \
		set -a; [ -f $(ROOT_DIR)/.env ] && . $(ROOT_DIR)/.env; set +a; \
		mkdir -p logs; \
		nohup $(BIN_PATH) >> $(ROOT_DIR)/$(LOG_FILE) 2>&1 & echo $$! > $(ROOT_DIR)/$(PID_FILE); \
		echo "✓ Backend started (PID: $$(cat $(PID_FILE)))"; \
	fi
	@sleep 2
	@echo "====================================="
	@echo "✓ Backend API: http://localhost:8080"
	@echo "✓ Frontend:    http://localhost:3000"
	@echo "====================================="
	@echo "Run 'make stop-all' to stop all services"

stop-all: ## Stop both backend and frontend
	@echo "Stopping all services..."
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "✓ Stopped backend (PID $$PID)"; \
		fi; \
		rm -f $(PID_FILE); \
	fi
	@if [ -f frontend.pid ]; then \
		PID=$$(cat frontend.pid); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "✓ Stopped frontend (PID $$PID)"; \
		fi; \
		rm -f frontend.pid; \
	fi
	@-pkill -f "vite" 2>/dev/null
	@echo "All services stopped"

all: clean deps fmt vet build test ## Full cycle: clean, deps, format, vet, build, test

.DEFAULT_GOAL := help

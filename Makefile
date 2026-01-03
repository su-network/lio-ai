.PHONY: help build run run-bg stop logs deps test test-coverage fmt vet lint security clean db-reset all frontend-install frontend-dev frontend-build ai-install ai-dev ai-stop ai-logs dev start stop-all restart status test-security

# Root-level Makefile to manage all Lio AI services (Go Gateway + Python AI + Frontend)

GO_DIR ?= joles
AI_DIR ?= ai
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
AI_PID_FILE := $(AI_DIR)/ai_service.pid
FRONTEND_PID_FILE := frontend.pid
LOG_FILE := logs/server.log
AI_LOG_FILE := $(AI_DIR)/ai_service.log
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

logs: ## Tail the server log
	@mkdir -p logs
	tail -f $(LOG_FILE)

deps: ## Download and verify all dependencies (Go, Python, Frontend)
	@echo "Checking and installing all dependencies..."
	@echo "1. Go dependencies..."
	cd $(GO_DIR) && $(GOMOD) download
	@echo "2. Python AI dependencies..."
	@if ! python3 -c "import uvicorn" 2>/dev/null; then \
		echo "Installing Python dependencies..."; \
		cd $(AI_DIR) && pip3 install -r requirements.txt; \
	else \
		echo "✓ Python dependencies already installed"; \
	fi
	@echo "3. Frontend dependencies..."
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd $(FRONTEND_DIR) && npm install; \
	else \
		echo "✓ Frontend dependencies already installed"; \
	fi
	@echo "✓ All dependencies ready"

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
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd $(FRONTEND_DIR) && npm install; \
	fi
	@cd $(FRONTEND_DIR) && npm run dev > /dev/null 2>&1 & echo $$! > $(ROOT_DIR)/frontend.pid
	@sleep 2
	@echo "Frontend started on http://localhost:3000 (PID: $$(cat $(ROOT_DIR)/frontend.pid))"

frontend-build: ## Build frontend for production
	cd $(FRONTEND_DIR) && npm run build

# Python AI Service commands
ai-install: ## Install Python AI service dependencies
	@echo "Installing Python AI service dependencies..."
	cd $(AI_DIR) && pip3 install -r requirements.txt

ai-dev: ## Run Python AI service in background
	@echo "Starting Python AI service..."
	@if [ -f $(AI_PID_FILE) ] && kill -0 $$(cat $(AI_PID_FILE)) 2>/dev/null; then \
		echo "✓ AI service already running with PID $$(cat $(AI_PID_FILE))"; \
	else \
		if ! python3 -c "import uvicorn" 2>/dev/null; then \
			echo "Installing Python dependencies..."; \
			cd $(AI_DIR) && pip3 install -r requirements.txt; \
		fi; \
		cd $(AI_DIR) && nohup python3 -m uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload > $(ROOT_DIR)/$(AI_LOG_FILE) 2>&1 & PID=$$!; echo $$PID > $(ROOT_DIR)/$(AI_PID_FILE); \
		sleep 3; \
		if kill -0 $$PID 2>/dev/null; then \
			echo "✓ AI service started (PID: $$PID)"; \
		else \
			echo "✗ Failed to start AI service. Check $(AI_LOG_FILE) for details"; \
			cat $(ROOT_DIR)/$(AI_LOG_FILE) | tail -20; \
			exit 1; \
		fi; \
	fi

ai-stop: ## Stop Python AI service
	@echo "Stopping Python AI service..."
	@if [ -f $(AI_PID_FILE) ]; then \
		PID=$$(cat $(AI_PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "✓ Stopped AI service (PID $$PID)"; \
		fi; \
		rm -f $(AI_PID_FILE); \
	fi
	@-pkill -f "python -m app.main" 2>/dev/null || true

ai-logs: ## Tail Python AI service logs
	@if [ -f $(AI_LOG_FILE) ]; then \
		tail -f $(AI_LOG_FILE); \
	else \
		echo "No AI service log file found at $(AI_LOG_FILE)"; \
	fi

# Combined commands
start: build ai-dev ## Start all services (Go Gateway + Python AI)
	@echo "====================================="
	@echo "Starting Lio AI Platform"
	@echo "====================================="
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "✓ Go Gateway already running with PID $$(cat $(PID_FILE))"; \
	else \
		set -a; [ -f $(ROOT_DIR)/.env ] && . $(ROOT_DIR)/.env; set +a; \
		mkdir -p logs; \
		nohup $(BIN_PATH) >> $(ROOT_DIR)/$(LOG_FILE) 2>&1 & echo $$! > $(ROOT_DIR)/$(PID_FILE); \
		sleep 2; \
		echo "✓ Go Gateway started (PID: $$(cat $(PID_FILE)))"; \
	fi
	@echo "====================================="
	@echo "✓ Go Gateway:  http://localhost:8080"
	@echo "✓ Python AI:   http://localhost:8000"
	@echo "====================================="
	@echo "Run 'make stop' to stop all services"
	@echo "Run 'make logs' to view Go Gateway logs"
	@echo "Run 'make ai-logs' to view Python AI logs"

dev: deps build ai-dev frontend-dev ## Run all services in development mode (Go + Python AI + Frontend)
	@echo "====================================="
	@echo "Starting Lio AI Development Environment"
	@echo "====================================="
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "✓ Go Gateway already running with PID $$(cat $(PID_FILE))"; \
	else \
		set -a; [ -f $(GO_DIR)/.env ] && . $(GO_DIR)/.env; set +a; \
		mkdir -p logs; \
		nohup $(BIN_PATH) >> $(ROOT_DIR)/$(LOG_FILE) 2>&1 & echo $$! > $(ROOT_DIR)/$(PID_FILE); \
		sleep 2; \
		echo "✓ Go Gateway started (PID: $$(cat $(PID_FILE)))"; \
	fi
	@sleep 1
	@echo "====================================="
	@echo "✓ Go Gateway:  http://localhost:8080"
	@echo "✓ Python AI:   http://localhost:8000"
	@echo "✓ Frontend:    http://localhost:5173"
	@echo "====================================="
	@echo "Run 'make stop-all' to stop all services"

stop: ai-stop ## Stop Go Gateway and Python AI service
	@echo "Stopping Go Gateway..."
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "✓ Stopped Go Gateway (PID $$PID)"; \
		fi; \
		rm -f $(PID_FILE); \
	else \
		echo "Go Gateway not running"; \
	fi

stop-all: stop ## Stop all services (Go Gateway + Python AI + Frontend)
	@echo "Stopping frontend..."
	@if [ -f $(FRONTEND_PID_FILE) ]; then \
		PID=$$(cat $(FRONTEND_PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "✓ Stopped frontend (PID $$PID)"; \
		fi; \
		rm -f $(FRONTEND_PID_FILE); \
	fi
	@-pkill -f "vite" 2>/dev/null && echo "✓ Stopped vite dev server" || true
	@echo "Cleaning up any remaining processes..."
	@-killall lio-ai 2>/dev/null && echo "✓ Killed all lio-ai processes" || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null && echo "✓ Freed port 8080" || true
	@-lsof -ti:8000 | xargs kill -9 2>/dev/null && echo "✓ Freed port 8000" || true
	@-lsof -ti:5173 | xargs kill -9 2>/dev/null && echo "✓ Freed port 5173" || true
	@echo "====================================="
	@echo "All services stopped"
	@echo "====================================="

restart: stop start ## Restart Go Gateway and Python AI service
	@echo "Services restarted successfully"

test-security: ## Run security tests
	@echo "Running security tests..."
	cd $(GO_DIR) && $(GOTEST) ./tests/... -v

status: ## Show status of all services
	@echo "====================================="
	@echo "Lio AI Platform Status"
	@echo "====================================="
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "✓ Go Gateway:  RUNNING (PID: $$(cat $(PID_FILE)))"; \
		echo "  URL: http://localhost:8080"; \
	else \
		echo "✗ Go Gateway:  STOPPED"; \
	fi
	@echo ""
	@if [ -f $(AI_PID_FILE) ] && kill -0 $$(cat $(AI_PID_FILE)) 2>/dev/null; then \
		echo "✓ Python AI:   RUNNING (PID: $$(cat $(AI_PID_FILE)))"; \
		echo "  URL: http://localhost:8000"; \
	else \
		echo "✗ Python AI:   STOPPED"; \
	fi
	@echo ""
	@if [ -f $(FRONTEND_PID_FILE) ] && kill -0 $$(cat $(FRONTEND_PID_FILE)) 2>/dev/null; then \
		echo "✓ Frontend:    RUNNING (PID: $$(cat $(FRONTEND_PID_FILE)))"; \
		echo "  URL: http://localhost:5173"; \
	else \
		echo "✗ Frontend:    STOPPED"; \
	fi
	@echo "====================================="

all: clean deps fmt vet build test ## Full cycle: clean, deps, format, vet, build, test

.DEFAULT_GOAL := help

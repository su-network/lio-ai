#!/bin/bash

# Live Development Environment for Lio AI Platform
# This script starts all services with live reload capabilities

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
log() {
    echo -e "${GREEN}[DEV]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Function to kill all background processes on exit
cleanup() {
    log "Stopping all development services..."
    
    # Kill by process patterns
    pkill -f "go run.*main.go" 2>/dev/null || true
    pkill -f "uvicorn.*app.main:app" 2>/dev/null || true
    pkill -f "vite" 2>/dev/null || true
    pkill -f "air" 2>/dev/null || true
    
    # Kill by port numbers
    fuser -k 8000/tcp 2>/dev/null || true
    fuser -k 8080/tcp 2>/dev/null || true
    fuser -k 3000/tcp 2>/dev/null || true
    
    # Kill specific processes by PID files if they exist
    if [ -f ../ai_service.pid ]; then
        kill $(cat ../ai_service.pid) 2>/dev/null || true
        rm -f ../ai_service.pid
    fi
    
    if [ -f ../gateway_service.pid ]; then
        kill $(cat ../gateway_service.pid) 2>/dev/null || true
        rm -f ../gateway_service.pid
    fi
    
    if [ -f ../frontend_service.pid ]; then
        kill $(cat ../frontend_service.pid) 2>/dev/null || true
        rm -f ../frontend_service.pid
    fi
    
    sleep 2
}

# Function to cleanup and exit
cleanup_and_exit() {
    cleanup
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup_and_exit SIGINT SIGTERM

# Check if required tools are installed
check_dependencies() {
    log "Checking dependencies..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        error "Go is not installed"
        exit 1
    fi
    
    # Check Python/Poetry
    if ! command -v poetry &> /dev/null; then
        warn "Poetry not found, checking for Python..."
        if ! command -v python3 &> /dev/null; then
            error "Python3 is not installed"
            exit 1
        fi
    fi
    
    # Check Node.js/npm
    if ! command -v npm &> /dev/null; then
        error "Node.js/npm is not installed"
        exit 1
    fi
    
    # Check for Air (Go live reload)
    if ! command -v air &> /dev/null; then
        warn "Air not found, installing..."
        go install github.com/air-verse/air@latest
        if ! command -v air &> /dev/null; then
            warn "Air installation failed, will use 'go run' instead"
        fi
    fi
    
    log "✓ Dependencies check completed"
}

# Install frontend dependencies
setup_frontend() {
    log "Setting up frontend..."
    cd frontend
    if [ ! -d "node_modules" ]; then
        info "Installing frontend dependencies..."
        npm install
    fi
    cd ..
    log "✓ Frontend setup completed"
}

# Install backend dependencies
setup_backend() {
    log "Setting up Python AI backend..."
    cd ai
    if command -v poetry &> /dev/null; then
        poetry install
    else
        pip install -r requirements.txt 2>/dev/null || warn "Could not install Python dependencies"
    fi
    cd ..
    log "✓ Python AI backend setup completed"
}

# Setup Go gateway
setup_gateway() {
    log "Setting up Go gateway..."
    cd joles
    go mod download
    cd ..
    log "✓ Go gateway setup completed"
}

# Start Python AI service with live reload
start_python_ai() {
    log "Starting Python AI service with live reload..."
    cd ai
    
    # Create .air.toml for Python if using air, otherwise use uvicorn with reload
    if command -v poetry &> /dev/null; then
        poetry run uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload --log-level info &
    else
        python3 -m uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload --log-level info &
    fi
    
    PYTHON_PID=$!
    echo $PYTHON_PID > ../ai_service.pid
    cd ..
    info "Python AI service started (PID: $PYTHON_PID) - http://localhost:8000"
}

# Start Go gateway with live reload
start_go_gateway() {
    log "Starting Go gateway with live reload..."
    cd joles
    
    # Create .air.toml for Go live reload
    cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF

    if command -v air &> /dev/null; then
        air &
    else
        # Fallback to go run with basic file watching
        go run ./cmd/server/main.go &
    fi
    
    GATEWAY_PID=$!
    echo $GATEWAY_PID > ../server.pid
    cd ..
    info "Go Gateway started (PID: $GATEWAY_PID) - http://localhost:8080"
}

# Start frontend with live reload (Vite has this by default)
start_frontend() {
    log "Starting frontend with live reload..."
    cd frontend
    npm run dev &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > ../frontend.pid
    cd ..
    info "Frontend started (PID: $FRONTEND_PID) - http://localhost:3000"
}

# Wait for services to start
wait_for_services() {
    log "Waiting for services to start..."
    sleep 3
    
    # Check Python AI service
    for i in {1..10}; do
        if curl -s http://localhost:8000/health > /dev/null 2>&1; then
            log "✓ Python AI service is ready"
            break
        fi
        if [ $i -eq 10 ]; then
            error "Python AI service failed to start"
        fi
        sleep 1
    done
    
    # Check Go Gateway
    for i in {1..10}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            log "✓ Go Gateway is ready"
            break
        fi
        if [ $i -eq 10 ]; then
            error "Go Gateway failed to start"
        fi
        sleep 1
    done
    
    # Check Frontend (just check if port is open)
    for i in {1..10}; do
        if curl -s http://localhost:3000 > /dev/null 2>&1; then
            log "✓ Frontend is ready"
            break
        fi
        if [ $i -eq 10 ]; then
            error "Frontend failed to start"
        fi
        sleep 1
    done
}

# Show development information
show_dev_info() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║               LIO AI - LIVE DEVELOPMENT MODE                   ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}🌐 Frontend (Vite):${NC}     http://localhost:3000"
    echo -e "${BLUE}🚀 Go Gateway:${NC}         http://localhost:8080"
    echo -e "${BLUE}🐍 Python AI:${NC}          http://localhost:8000"
    echo ""
    echo -e "${YELLOW}📁 Watching for changes:${NC}"
    echo -e "   • Frontend: ${GREEN}src/**/*.{vue,ts,js}${NC}"
    echo -e "   • Gateway:  ${GREEN}internal/**/*.go, cmd/**/*.go${NC}"
    echo -e "   • AI:       ${GREEN}app/**/*.py${NC}"
    echo ""
    echo -e "${BLUE}🔧 Development Features:${NC}"
    echo -e "   • Hot Module Replacement (Frontend)"
    echo -e "   • Automatic Go binary rebuild"
    echo -e "   • Python auto-reload on file changes"
    echo -e "   • Live API synchronization"
    echo ""
    echo -e "${GREEN}Press Ctrl+C to stop all services${NC}"
    echo ""
}

# Monitor logs in a simple way
monitor_logs() {
    log "Monitoring services... (Press Ctrl+C to stop)"
    
    while true; do
        sleep 5
        
        # Check if services are still running
        if ! curl -s http://localhost:8000/health > /dev/null 2>&1; then
            warn "Python AI service may have stopped"
        fi
        
        if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
            warn "Go Gateway may have stopped"
        fi
        
        if ! curl -s http://localhost:3000 > /dev/null 2>&1; then
            warn "Frontend may have stopped"
        fi
    done
}

# Main execution
main() {
    log "Starting Lio AI Live Development Environment..."
    
    # Stop any existing services
    cleanup
    
    # Setup
    check_dependencies
    setup_frontend
    setup_backend
    setup_gateway
    
    # Start services
    start_python_ai
    start_go_gateway
    start_frontend
    
    # Wait and verify
    wait_for_services
    
    # Show info
    show_dev_info
    
    # Monitor
    monitor_logs
}

# Run main function
main "$@"
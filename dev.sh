#!/bin/bash

# Lio AI Development Environment Launcher
# This script ensures environment is configured and starts all services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         LIO AI - DEVELOPMENT ENVIRONMENT LAUNCHER              ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Ensure .env exists
if [ ! -f "joles/.env" ]; then
    echo -e "${YELLOW}⚙️  Creating .env file...${NC}"
    cat > joles/.env << 'EOF'
APP_NAME=lio-ai
ENVIRONMENT=development
PORT=8080
DATABASE_URL=./lio-ai.db
JWT_SECRET_KEY=dev-secret-key-minimum-32-characters-long-change-in-prod
CORS_ORIGINS=http://localhost:5173,http://localhost:3000
LITELLM_BASE_URL=http://localhost:8000
LOG_LEVEL=info
EOF
    echo -e "${GREEN}✅ Environment configured${NC}"
    echo ""
fi

# Check if Python dependencies are installed
echo -e "${BLUE}🔍 Checking Python dependencies...${NC}"
if ! python3 -c "import uvicorn" 2>/dev/null; then
    echo -e "${YELLOW}📦 Installing Python AI dependencies...${NC}"
    cd ai && pip3 install -r requirements.txt && cd ..
    echo -e "${GREEN}✅ Python dependencies installed${NC}"
else
    echo -e "${GREEN}✅ Python dependencies OK${NC}"
fi

# Check if Frontend dependencies are installed
echo -e "${BLUE}🔍 Checking Frontend dependencies...${NC}"
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}📦 Installing Frontend dependencies...${NC}"
    cd frontend && npm install && cd ..
    echo -e "${GREEN}✅ Frontend dependencies installed${NC}"
else
    echo -e "${GREEN}✅ Frontend dependencies OK${NC}"
fi

# Check if Go dependencies are downloaded
echo -e "${BLUE}🔍 Checking Go dependencies...${NC}"
cd joles && go mod download && cd ..
echo -e "${GREEN}✅ Go dependencies OK${NC}"

echo ""
echo -e "${GREEN}🚀 Starting all services...${NC}"
echo ""
echo -e "${BLUE}Services will start on:${NC}"
echo -e "   • ${GREEN}Frontend:${NC}    http://localhost:5173"
echo -e "   • ${GREEN}Go Gateway:${NC}  http://localhost:8080"
echo -e "   • ${GREEN}Python AI:${NC}   http://localhost:8000"
echo ""
echo -e "${YELLOW}📝 Logs will be available at:${NC}"
echo -e "   • Frontend: frontend/vite.log"
echo -e "   • Gateway:  logs/server.log"
echo -e "   • AI:       ai/ai_service.log"
echo ""
echo -e "${BLUE}Press Ctrl+C to stop all services${NC}"
echo ""

# Start services using make
exec make dev
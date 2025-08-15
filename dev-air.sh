#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Fluxio Development with Air (Auto-reload)${NC}"
echo "================================================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go no está instalado. Por favor instálalo primero.${NC}"
    exit 1
fi

# Check if Air is installed
if ! command -v air &> /dev/null; then
    echo -e "${YELLOW}⚠️  Air no está instalado. Instalando...${NC}"
    go install github.com/air-verse/air@latest
fi

echo -e "${GREEN}✅ Go y Air están disponibles${NC}"

# Install dependencies
echo -e "${BLUE}📦 Instalando dependencias...${NC}"
go mod tidy

# Generate Swagger docs
echo -e "${BLUE}📚 Generando documentación Swagger...${NC}"
swag init -g cmd/server/main.go

# Set environment variables
export DATABASE_URL="postgres://postgres:123456@127.0.0.1:5432/fluxio?sslmode=disable"
export JWT_SECRET="dev-super-secret-jwt-key-change-in-production"

echo -e "${GREEN}🚀 Iniciando desarrollo con Air (auto-reload)...${NC}"
echo -e "${GREEN}🌐 Swagger UI: http://localhost:8080/swagger/index.html${NC}"
echo -e "${GREEN}🔌 API: http://localhost:8080${NC}"
echo -e "${GREEN}🗄️  Base de datos: $DATABASE_URL${NC}"
echo ""
echo -e "${YELLOW}✨ Air detectará cambios automáticamente y reiniciará el servidor${NC}"
echo -e "${YELLOW}Presiona Ctrl+C para detener${NC}"

# Run with Air for auto-reload
air

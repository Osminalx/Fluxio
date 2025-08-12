#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Fluxio Hybrid Development Script${NC}"
echo "=========================================="

# Check if Podman is installed
if ! command -v podman &> /dev/null; then
    echo -e "${RED}❌ Podman no está instalado. Por favor instálalo primero.${NC}"
    exit 1
fi

# Check if podman-compose is installed
if ! command -v podman-compose &> /dev/null; then
    echo -e "${YELLOW}⚠️  podman-compose no está instalado. Instalando...${NC}"
    pip3 install podman-compose
fi

echo -e "${GREEN}✅ Podman y podman-compose están disponibles${NC}"

# Start PostgreSQL in Podman
echo -e "${BLUE}🗄️  Iniciando PostgreSQL en Podman...${NC}"
podman-compose -f docker-compose.db.yml up -d

# Wait for PostgreSQL to be ready
echo -e "${BLUE}⏳ Esperando que PostgreSQL esté listo...${NC}"
sleep 15

# Check PostgreSQL status
echo -e "${BLUE}📊 Verificando estado de PostgreSQL...${NC}"
podman-compose -f docker-compose.db.yml ps

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go no está instalado. Por favor instálalo primero.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go está disponible${NC}"

# Install dependencies
echo -e "${BLUE}📦 Instalando dependencias...${NC}"
go mod tidy

# Generate Swagger docs
echo -e "${BLUE}📚 Generando documentación Swagger...${NC}"
swag init -g cmd/server/main.go

# Build the application
echo -e "${BLUE}🔨 Construyendo la aplicación...${NC}"
go build -o bin/fluxio ./cmd/server

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Aplicación construida exitosamente${NC}"
else
    echo -e "${RED}❌ Error al construir la aplicación${NC}"
    exit 1
fi

# Set environment variables
export DATABASE_URL="postgres://postgres:123456@localhost:5432/fluxio?sslmode=disable"
export JWT_SECRET="dev-secret-key-change-in-production"

echo -e "${GREEN}🚀 Iniciando aplicación en modo desarrollo híbrido...${NC}"
echo -e "${GREEN}🌐 Swagger UI: http://localhost:8080/swagger/index.html${NC}"
echo -e "${GREEN}🔌 API: http://localhost:8080${NC}"
echo -e "${GREEN}🗄️  Base de datos: $DATABASE_URL${NC}"
echo -e "${GREEN}🐳 PostgreSQL ejecutándose en Podman${NC}"
echo ""
echo -e "${YELLOW}Presiona Ctrl+C para detener${NC}"
echo -e "${YELLOW}Para detener PostgreSQL: podman-compose -f docker-compose.db.yml down${NC}"

# Run the application
./bin/fluxio

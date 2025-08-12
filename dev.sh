#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛠️  Fluxio Development Script${NC}"
echo "================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go no está instalado. Por favor instálalo primero.${NC}"
    exit 1
fi

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  PostgreSQL no está ejecutándose en localhost:5432${NC}"
    echo -e "${YELLOW}💡 Inicia PostgreSQL o usa el script de despliegue con Podman${NC}"
    echo ""
    echo -e "${BLUE}Para usar Podman:${NC}"
    echo "  ./deploy.sh"
    echo ""
    echo -e "${BLUE}Para desarrollo local:${NC}"
    echo "  1. Instala PostgreSQL"
    echo "  2. Crea la base de datos 'fluxio'"
    echo "  3. Ejecuta: ./dev.sh"
    exit 1
fi

echo -e "${GREEN}✅ Go y PostgreSQL están disponibles${NC}"

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

echo -e "${GREEN}🚀 Iniciando aplicación en modo desarrollo...${NC}"
echo -e "${GREEN}🌐 Swagger UI: http://localhost:8080/swagger/index.html${NC}"
echo -e "${GREEN}🔌 API: http://localhost:8080${NC}"
echo -e "${GREEN}🗄️  Base de datos: $DATABASE_URL${NC}"
echo ""
echo -e "${YELLOW}Presiona Ctrl+C para detener${NC}"

# Run the application
./bin/fluxio

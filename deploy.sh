#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Fluxio Deployment Script${NC}"
echo "================================"

# Check if Podman is installed
if ! command -v podman &> /dev/null; then
    echo -e "${RED}❌ Podman no está instalado. Por favor instálalo primero.${NC}"
    echo "En Fedora/RHEL: sudo dnf install podman"
    echo "En Ubuntu: sudo apt install podman"
    exit 1
fi

# Check if podman-compose is installed
if ! command -v podman-compose &> /dev/null; then
    echo -e "${YELLOW}⚠️  podman-compose no está instalado. Instalando...${NC}"
    pip3 install podman-compose
fi

echo -e "${GREEN}✅ Podman y podman-compose están disponibles${NC}"

# Stop and remove existing containers
echo -e "${BLUE}🔄 Deteniendo contenedores existentes...${NC}"
podman-compose down

# Remove existing images (optional, uncomment if you want fresh builds)
# echo -e "${BLUE}🗑️  Eliminando imágenes existentes...${NC}"
# podman rmi fluxio-app fluxio-postgres

# Build and start services
echo -e "${BLUE}🔨 Construyendo y iniciando servicios...${NC}"
podman-compose up --build -d

# Wait for services to be ready
echo -e "${BLUE}⏳ Esperando que los servicios estén listos...${NC}"
sleep 30

# Check service status
echo -e "${BLUE}📊 Verificando estado de los servicios...${NC}"
podman-compose ps

# Test the application
echo -e "${BLUE}🧪 Probando la aplicación...${NC}"
if curl -f http://localhost:8080/hello > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Aplicación funcionando correctamente!${NC}"
    echo -e "${GREEN}🌐 Swagger UI: http://localhost:8080/swagger/index.html${NC}"
    echo -e "${GREEN}🔌 API: http://localhost:8080${NC}"
    echo -e "${GREEN}🗄️  PostgreSQL: localhost:5432${NC}"
else
    echo -e "${RED}❌ Error: La aplicación no responde${NC}"
    echo -e "${YELLOW}📋 Logs de la aplicación:${NC}"
    podman-compose logs app
fi

echo -e "${BLUE}🎉 Despliegue completado!${NC}"
echo ""
echo -e "${YELLOW}Comandos útiles:${NC}"
echo "  Ver logs: podman-compose logs -f"
echo "  Detener: podman-compose down"
echo "  Reiniciar: podman-compose restart"
echo "  Estado: podman-compose ps"

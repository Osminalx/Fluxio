#!/bin/bash

echo "🚀 Fluxio Development with Air (Auto-reload) - Windows Optimized"
echo "====================================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go no está instalado. Por favor instálalo primero."
    exit 1
fi

# Check if Air is installed
if ! command -v air &> /dev/null; then
    echo "⚠️  Air no está instalado. Instalando..."
    go install github.com/air-verse/air@latest
fi

echo "✅ Go y Air están disponibles"

# Install dependencies
echo "📦 Instalando dependencias..."
go mod tidy

# Generate Swagger docs
echo "📚 Generando documentación Swagger..."
swag init -g cmd/server/main.go

# Set environment variables
export DATABASE_URL="postgres://postgres:123456@127.0.0.1:5432/fluxio?sslmode=disable"
export JWT_SECRET="dev-super-secret-jwt-key-change-in-production"

# Create tmp directory if it doesn't exist
if [ ! -d "tmp" ]; then
    mkdir -p tmp
    echo "📁 Directorio tmp creado"
fi

echo "🚀 Iniciando desarrollo con Air (auto-reload)..."
echo "🌐 Swagger UI: http://localhost:8080/swagger/index.html"
echo "🔌 API: http://localhost:8080"
echo "🗄️  Base de datos: $DATABASE_URL"
echo "⚙️  Usando configuración optimizada para Windows"
echo ""
echo "✨ Air detectará cambios automáticamente y reiniciará el servidor"
echo "Presiona Ctrl+C para detener"

# Run with Air using Windows-specific configuration
air -c .air.windows.toml

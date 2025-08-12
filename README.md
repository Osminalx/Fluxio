# 🚀 Fluxio API

API de autenticación y gestión de usuarios construida con Go, GORM, JWT y PostgreSQL.

## ✨ Características

- 🔐 Autenticación JWT
- 🗄️ Base de datos PostgreSQL con GORM
- 📚 Documentación automática con Swagger
- 🐳 Contenedores Docker/Podman
- 🆔 IDs UUID para mayor seguridad
- 🔒 Middleware de autenticación
- 🧪 Health checks integrados

## 🛠️ Tecnologías

- **Backend**: Go 1.24+
- **Base de datos**: PostgreSQL 15
- **ORM**: GORM
- **Autenticación**: JWT
- **Documentación**: Swagger/OpenAPI
- **Contenedores**: Docker/Podman

## 🚀 Despliegue Rápido con Podman

### Prerrequisitos

1. **Podman** instalado
2. **podman-compose** instalado

```bash
# En Fedora/RHEL
sudo dnf install podman

# En Ubuntu/Debian
sudo apt install podman

# Instalar podman-compose
pip3 install podman-compose
```

### Despliegue Automático

```bash
# Dar permisos de ejecución
chmod +x deploy.sh

# Ejecutar despliegue
./deploy.sh
```

### Despliegue Manual

```bash
# Construir y ejecutar
podman-compose up --build -d

# Ver logs
podman-compose logs -f

# Verificar estado
podman-compose ps

# Detener servicios
podman-compose down
```

## 🛠️ Desarrollo Local

### Prerrequisitos

1. **Go 1.24+** instalado
2. **PostgreSQL** ejecutándose en localhost:5432
3. **Base de datos 'fluxio'** creada

### Ejecutar en Desarrollo

```bash
# Dar permisos de ejecución
chmod +x dev.sh

# Ejecutar en modo desarrollo
./dev.sh
```

### Desarrollo Manual

```bash
# Instalar dependencias
go mod tidy

# Generar documentación Swagger
swag init -g cmd/server/main.go

# Ejecutar
go run cmd/server/main.go
```

## 📚 Endpoints de la API

### Públicos
- `GET /hello` - Endpoint de prueba
- `POST /auth/register` - Registrar usuario
- `POST /auth/login` - Iniciar sesión

### Protegidos (requieren JWT)
- `GET /protected` - Endpoint protegido

## 🔐 Autenticación

### 1. Registrar Usuario
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@ejemplo.com",
    "password": "contraseña123",
    "name": "Juan Pérez"
  }'
```

### 2. Iniciar Sesión
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@ejemplo.com",
    "password": "contraseña123"
  }'
```

### 3. Usar Endpoint Protegido
```bash
curl -X GET http://localhost:8080/protected \
  -H "Authorization: Bearer TU_TOKEN_JWT_AQUI"
```

## 📖 Documentación Swagger

Una vez que la aplicación esté ejecutándose, accede a:

**🌐 Swagger UI**: http://localhost:8080/swagger/index.html

## 🗄️ Base de Datos

### Estructura de Tablas

```sql
-- Tabla de usuarios
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### Variables de Entorno

```bash
DATABASE_URL=postgres://postgres:123456@localhost:5432/fluxio?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

## 🐳 Docker

### Construir Imagen

```bash
podman build -t fluxio:latest .
```

### Ejecutar Contenedor

```bash
podman run -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:123456@host:5432/fluxio" \
  fluxio:latest
```

## 📁 Estructura del Proyecto

```
fluxio/
├── cmd/
│   └── server/
│       └── main.go          # Punto de entrada
├── internal/
│   ├── api/                 # Handlers HTTP
│   ├── auth/                # Middleware de autenticación
│   ├── db/                  # Conexión a base de datos
│   ├── models/              # Modelos GORM
│   └── services/            # Lógica de negocio
├── docs/                    # Documentación Swagger
├── Dockerfile               # Imagen Docker
├── docker-compose.yml       # Orquestación
├── deploy.sh                # Script de despliegue
├── dev.sh                   # Script de desarrollo
└── README.md               # Este archivo
```

## 🔧 Comandos Útiles

```bash
# Ver logs en tiempo real
podman-compose logs -f

# Reiniciar servicios
podman-compose restart

# Ver estado de servicios
podman-compose ps

# Ejecutar comandos en contenedor
podman-compose exec app sh
podman-compose exec postgres psql -U postgres -d fluxio

# Limpiar todo
podman-compose down -v
podman system prune -a
```

## 🚨 Solución de Problemas

### Error de Conexión a Base de Datos
```bash
# Verificar que PostgreSQL esté ejecutándose
podman-compose logs postgres

# Verificar conectividad
podman-compose exec app wget -O- http://postgres:5432
```

### Error de Permisos
```bash
# En sistemas SELinux
sudo setsebool -P container_manage_cgroup 1
```

### Puerto Ocupado
```bash
# Ver qué está usando el puerto 8080
sudo netstat -tulpn | grep :8080

# Cambiar puerto en docker-compose.yml
ports:
  - "8081:8080"
```

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 🆘 Soporte

Si tienes problemas o preguntas:

1. Revisa los logs: `podman-compose logs -f`
2. Verifica el estado: `podman-compose ps`
3. Revisa la documentación Swagger
4. Abre un issue en GitHub

---

**¡Disfruta desarrollando con Fluxio! 🎉**

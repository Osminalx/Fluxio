# 🚀 Fluxio API

Personal Finances API built with Go, GORM, JWT, and PostgreSQL.

## ✨ Features

- 🔐 JWT Authentication
- 🗄️ PostgreSQL database with GORM ORM
- 📚 Automatic API documentation with Swagger
- 🐳 Docker/Podman containerization
- 🆔 UUID IDs for enhanced security
- 🔒 Authentication middleware
- 🧪 Built-in health checks
- 🚀 Hybrid development setup (PostgreSQL in containers, Go locally)

## 🛠️ Technologies

- **Backend**: Go 1.24+
- **Database**: PostgreSQL 15
- **ORM**: GORM
- **Authentication**: JWT
- **Documentation**: Swagger/OpenAPI
- **Containers**: Docker/Podman

## 🚀 Quick Start with Podman

### Prerequisites

1. **Podman** installed
2. **podman-compose** installed

```bash
# On Fedora/RHEL
sudo dnf install podman

# On Ubuntu/Debian
sudo apt install podman

# Install podman-compose
pip3 install podman-compose
```

### Automatic Deployment

```bash
# Make scripts executable
chmod +x deploy.sh

# Run deployment
./deploy.sh
```

### Manual Deployment

```bash
# Build and run
podman-compose up --build -d

# View logs
podman-compose logs -f

# Check status
podman-compose ps

# Stop services
podman-compose down
```

## 🛠️ Local Development

### Prerequisites

1. **Go 1.24+** installed
2. **PostgreSQL** running on localhost:5432
3. **Database 'fluxio'** created

### Run in Development Mode

```bash
# Make scripts executable
chmod +x dev.sh

# Run in development mode
./dev.sh
```

### Hybrid Development (Recommended)

```bash
# Make script executable
chmod +x dev-hybrid.sh

# Run hybrid mode (PostgreSQL in Podman, Go locally)
./dev-hybrid.sh
```

### Manual Development

```bash
# Install dependencies
go mod tidy

# Generate Swagger documentation
swag init -g cmd/server/main.go

# Run
go run cmd/server/main.go
```

## 📚 API Endpoints

### Public
- `GET /hello` - Test endpoint
- `POST /auth/register` - User registration
- `POST /auth/login` - User login

### Protected (require JWT)
- `GET /protected` - Protected endpoint

## 🔐 Authentication

### 1. Register User
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 3. Use Protected Endpoint
```bash
curl -X GET http://localhost:8080/protected \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

## 📖 Swagger Documentation

Once the application is running, access:

**🌐 Swagger UI**: http://localhost:8080/swagger/index.html

## 🗄️ Database

### Table Structure

```sql
-- Users table
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

### Environment Variables

```bash
DATABASE_URL=postgres://postgres:123456@localhost:5432/fluxio?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

## 🐳 Docker

### Build Image

```bash
podman build -t fluxio:latest .
```

### Run Container

```bash
podman run -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:123456@host:5432/fluxio" \
  fluxio:latest
```

## 📁 Project Structure

```
fluxio/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── api/                 # HTTP handlers
│   ├── auth/                # Authentication middleware
│   ├── db/                  # Database connection
│   ├── models/              # GORM models
│   └── services/            # Business logic
├── docs/                    # Swagger documentation
├── Dockerfile               # Docker image
├── docker-compose.yml       # Full orchestration
├── docker-compose.db.yml    # Database only
├── deploy.sh                # Deployment script
├── dev.sh                   # Local development script
├── dev-hybrid.sh            # Hybrid development script
└── README.md               # This file
```

## 🔧 Useful Commands

```bash
# View real-time logs
podman-compose logs -f

# Restart services
podman-compose restart

# Check service status
podman-compose ps

# Execute commands in container
podman-compose exec app sh
podman-compose exec postgres psql -U postgres -d fluxio

# Clean everything
podman-compose down -v
podman system prune -a
```

## 🚨 Troubleshooting

### Database Connection Error
```bash
# Check if PostgreSQL is running
podman-compose logs postgres

# Check connectivity
podman-compose exec app wget -O- http://postgres:5432
```

### Permission Issues
```bash
# On SELinux systems
sudo setsebool -P container_manage_cgroup 1
```

### Port Already in Use
```bash
# Check what's using port 8080
sudo netstat -tulpn | grep :8080

# Change port in docker-compose.yml
ports:
  - "8081:8080"
```

## 🤝 Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

If you have problems or questions:

1. Check the logs: `podman-compose logs -f`
2. Verify status: `podman-compose ps`
3. Check Swagger documentation
4. Open an issue on GitHub

---

**Enjoy developing with Fluxio! 🎉**

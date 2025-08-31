# VLESS Config Generator

A professional, containerized service for generating VLESS proxy configurations with QR codes. Built with Go, featuring structured logging, Docker support, and automated CI/CD.

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build Status](https://github.com/yourusername/vless-generator/workflows/Build%20and%20Push%20Docker%20Image/badge.svg)

## 🚀 Features

- **Multiple Configuration Types**: Support for NekoBox and VLESS clients
- **QR Code Generation**: Direct import via QR scanning
- **Professional Logging**: Structured logging with Logrus (JSON/Text formats)
- **Docker Ready**: Multi-architecture container support
- **Health Monitoring**: Built-in health check endpoints
- **Chrome Fingerprinting**: Enhanced security with uTLS
- **Configurable**: Extensive command-line and environment configuration
- **Production Ready**: Systemd service, graceful shutdown, error handling

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Docker Deployment](#docker-deployment)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)
- [Development](#development)
- [GitHub Actions CI/CD](#github-actions-cicd)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Quick Start

### Using Docker (Recommended)

```bash
# Pull and run the latest image
docker run -d \
  --name vless-generator \
  -p 8080:8080 \
  yourusername/vless-generator:latest \
  -server your-server.com \
  -server-port 443 \
  -ws-path /websocket
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/yourusername/vless-generator.git
cd vless-generator

# Copy and customize environment file
cp .env.example .env
nano .env

# Start the service
docker-compose up -d

# View logs
docker-compose logs -f
```

### Direct Installation

```bash
# Clone and build
git clone https://github.com/yourusername/vless-generator.git
cd vless-generator
go build .

# Run with custom configuration
./vless-generator -server your-server.com -log-level debug

# Or install as systemd service
sudo ./install.sh
```

## 📦 Installation

### Prerequisites

- **Go 1.21+** (for building from source)
- **Docker** (for containerized deployment)
- **Linux** (for systemd service installation)

### Method 1: Docker (Recommended)

```bash
# Using docker-compose
docker-compose up -d

# Or direct docker run
docker run -d \
  --name vless-generator \
  -p 8080:8080 \
  -e LOG_LEVEL=info \
  yourusername/vless-generator:latest
```

### Method 2: Pre-built Binary

```bash
# Download from releases
wget https://github.com/yourusername/vless-generator/releases/latest/download/vless-generator-linux-amd64
chmod +x vless-generator-linux-amd64
./vless-generator-linux-amd64 -server your-server.com
```

### Method 3: Build from Source

```bash
git clone https://github.com/yourusername/vless-generator.git
cd vless-generator
go mod download
go build -o vless-generator .
./vless-generator -server your-server.com
```

### Method 4: System Service Installation

```bash
git clone https://github.com/yourusername/vless-generator.git
cd vless-generator
sudo ./install.sh
```

This will:
- Build the binary
- Install to `/opt/vless-generator`
- Create systemd service
- Start the service automatically

## 🐳 Docker Deployment

### Environment Variables

Create a `.env` file from the template:

```bash
cp .env.example .env
```

Example `.env` configuration:

```env
VLESS_SERVER=your-server.example.com
VLESS_PORT=443
WS_PATH=/websocket
DNS_SERVER=8.8.8.8
DOH_SERVER=https://223.5.5.5/dns-query
LOG_LEVEL=info
LOG_FORMAT=json
```

### Docker Compose Configuration

```yaml
version: '3.8'

services:
  vless-generator:
    image: yourusername/vless-generator:latest
    container_name: vless-generator
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    command: [
      "-server", "your-server.com",
      "-server-port", "443",
      "-ws-path", "/websocket",
      "-log-level", "info"
    ]
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Docker Commands

```bash
# Build locally
docker build -t vless-generator .

# Run with custom configuration
docker run -d \
  --name vless-generator \
  -p 8080:8080 \
  vless-generator \
  -server your-server.com \
  -server-port 443 \
  -ws-path /websocket \
  -log-level debug

# View logs
docker logs -f vless-generator

# Health check
docker exec vless-generator wget -qO- http://localhost:8080/health
```

## ⚙️ Configuration

### Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8080` | HTTP server port |
| `-server` | `vless.example.com` | VLESS server address |
| `-server-port` | `443` | VLESS server port |
| `-ws-path` | `/websocket` | WebSocket path |
| `-dns-server` | `8.8.8.8` | Remote DNS server |
| `-doh-server` | `https://223.5.5.5/dns-query` | DNS over HTTPS server |
| `-tun-address` | `172.19.0.1/28` | TUN interface address |
| `-mixed-port` | `2080` | Mixed proxy port |
| `-tun-mtu` | `9000` | TUN interface MTU |
| `-log-level` | `info` | Log level (debug, info, warn, error) |
| `-log-format` | `json` | Log format (json, text) |

### Example Configurations

#### Development Mode
```bash
./vless-generator \
  -server localhost \
  -server-port 8443 \
  -log-level debug \
  -log-format text
```

#### Production Mode
```bash
./vless-generator \
  -server your-production-server.com \
  -server-port 443 \
  -ws-path /secure-websocket \
  -log-level info \
  -log-format json
```

#### Custom Network Configuration
```bash
./vless-generator \
  -server vpn.example.com \
  -dns-server 1.1.1.1 \
  -doh-server https://cloudflare-dns.com/dns-query \
  -tun-address 10.0.0.1/24
```

## 🌐 API Endpoints

### Configuration Pages

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/<type>/<uuid>` | Configuration page with QR code | `/vless/abc123-def456` |
| | Supported types: `vless`, `neko` | `/neko/abc123-def456` |

### Download Endpoints

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/config/<type>/<uuid>.json` | Download JSON configuration | `/config/vless/abc123-def456.json` |

### Service Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check and service info |

### Example Usage

```bash
# Visit configuration page
curl http://localhost:8080/vless/bae71742-94e0-4dd5-935f-070339819ba0

# Download configuration
curl -O http://localhost:8080/config/vless/bae71742-94e0-4dd5-935f-070339819ba0.json

# Health check
curl http://localhost:8080/health
```

### Health Check Response

```json
{
  "status": "healthy",
  "timestamp": "2025-08-31T16:32:48Z",
  "service": "vless-generator",
  "version": "1.0.0",
  "templates": ["neko", "vless"]
}
```

## 🛠️ Development

### Project Structure

```
vless-generator/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── Dockerfile                 # Container definition
├── docker-compose.yml         # Local development setup
├── .github/workflows/         # CI/CD pipelines
├── templates/                 # Configuration templates
│   ├── neko.json
│   └── vless.json
└── internal/                  # Internal packages
    ├── config/                # Configuration management
    ├── handlers/              # HTTP handlers
    ├── middleware/            # HTTP middleware
    ├── templates/             # Template management
    └── utils/                 # Utility functions
```

### Local Development

```bash
# Clone repository
git clone https://github.com/yourusername/vless-generator.git
cd vless-generator

# Install dependencies
go mod download

# Run with live reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Or run directly
go run . -server localhost -log-level debug -log-format text

# Run tests
go test ./...

# Build for production
go build -ldflags="-w -s" -o vless-generator .
```

### Adding New Configuration Types

1. Create a new template file in `templates/`:
```bash
cp templates/vless.json templates/newtype.json
# Edit newtype.json with your configuration
```

2. Update the template types in `internal/config/config.go`:
```go
cfg.Templates.Types = []string{"neko", "vless", "newtype"}
```

3. Test the new configuration:
```bash
go run . -server test.com
# Visit: http://localhost:8080/newtype/test-uuid
```

### Code Standards

- **Go Modules**: Use Go 1.21+ with modules
- **Structured Logging**: Use logrus with structured fields
- **Error Handling**: Wrap errors with context
- **Testing**: Write tests for new functionality
- **Documentation**: Document exported functions and types

## 🔄 GitHub Actions CI/CD

### Automated Workflows

The project includes automated CI/CD with the following features:

- **Testing**: Automated Go tests and linting
- **Multi-platform Builds**: Linux AMD64 and ARM64
- **Docker Hub Publishing**: Automatic image builds and pushes
- **Semantic Versioning**: Tag-based releases
- **Security Scanning**: Dependency and container scanning

### Required Secrets

Configure these secrets in your GitHub repository:

| Secret | Description |
|--------|-------------|
| `DOCKERHUB_USERNAME` | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token |

### Triggering Builds

```bash
# Push to main branch (builds latest tag)
git push origin main

# Create a release (builds versioned tag)
git tag v1.0.0
git push origin v1.0.0

# Pull request (runs tests only)
# No Docker image push on PRs
```

### Deployment Artifacts

Each build generates deployment artifacts including:
- `docker-compose.yml` for production deployment
- `README.md` with deployment instructions
- Multi-architecture Docker images

## 📊 Monitoring and Logging

### Log Formats

**JSON Format (Production)**:
```json
{
  "level": "info",
  "time": "2025-08-31T16:32:48Z",
  "component": "handlers",
  "method": "GET",
  "path": "/vless/abc123",
  "status_code": 200,
  "duration_ms": 15,
  "remote_addr": "127.0.0.1"
}
```

**Text Format (Development)**:
```
INFO[2025-08-31 16:32:48] HTTP request completed successfully component=handlers method=GET path=/vless/abc123 status_code=200
```

### Monitoring Integration

The service provides metrics suitable for:
- **Prometheus**: HTTP metrics via structured logs
- **Grafana**: Visualization of request patterns
- **ELK Stack**: Log aggregation and analysis
- **Docker Health Checks**: Container orchestration

## 🔧 Troubleshooting

### Common Issues

**Port Already in Use**:
```bash
# Check what's using the port
sudo lsof -i :8080

# Use a different port
./vless-generator -port 8081
```

**Template Loading Errors**:
```bash
# Ensure templates directory exists
ls -la templates/

# Check file permissions
chmod 644 templates/*.json
```

**Docker Build Issues**:
```bash
# Clear Docker cache
docker system prune -a

# Rebuild without cache
docker build --no-cache -t vless-generator .
```

### Debug Mode

Enable debug logging for detailed information:

```bash
# Command line
./vless-generator -log-level debug -log-format text

# Docker
docker run -d \
  --name vless-generator-debug \
  -p 8080:8080 \
  yourusername/vless-generator:latest \
  -log-level debug
```

### Health Monitoring

```bash
# Check service health
curl http://localhost:8080/health

# Monitor logs
docker logs -f vless-generator

# Check systemd service
sudo systemctl status vless-generator
sudo journalctl -u vless-generator -f
```

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** with proper tests
4. **Run tests**: `go test ./...`
5. **Commit changes**: `git commit -m "Add amazing feature"`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Setup

```bash
# Fork and clone
git clone https://github.com/yourusername/vless-generator.git
cd vless-generator

# Install pre-commit hooks (optional)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
golangci-lint run

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Logrus](https://github.com/sirupsen/logrus) for structured logging
- [go-qrcode](https://github.com/skip2/go-qrcode) for QR code generation
- [Docker](https://docker.com) for containerization
- [GitHub Actions](https://github.com/features/actions) for CI/CD

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/vless-generator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/vless-generator/discussions)
- **Documentation**: [Wiki](https://github.com/yourusername/vless-generator/wiki)

---

**Made with ❤️ by the VLESS Generator Team**

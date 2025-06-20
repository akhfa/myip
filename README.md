# My IP

[![Build Status](https://github.com/akhfa/myip/workflows/Release/badge.svg)](https://github.com/akhfa/myip/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/akhfa/myip)](https://goreportcard.com/report/github.com/akhfa/myip)
[![codecov](https://codecov.io/gh/akhfa/myip/branch/main/graph/badge.svg)](https://codecov.io/gh/akhfa/myip)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Docker Pulls](https://img.shields.io/docker/pulls/ghcr.io/akhfa/myip)](https://github.com/akhfa/myip/pkgs/container/myip)

A lightweight, fast HTTP service for detecting client IP addresses with comprehensive proxy header support and detailed IP information.

## Features

- 🌐 **Multi-Protocol Support**: Detects both IPv4 and IPv6 addresses
- 🔍 **Comprehensive Header Analysis**: Supports all major proxy headers (Cloudflare, nginx, Apache, etc.)
- 🏷️ **Multiple Output Formats**: Plain text, JSON, and detailed information endpoints
- 🛡️ **Security Focused**: Identifies private IPs, proxy chains, and Cloudflare detection
- 🚀 **High Performance**: Lightweight Go implementation with minimal dependencies
- 📊 **Health Monitoring**: Built-in health check endpoint
- 🐳 **Container Ready**: Multi-architecture Docker images (amd64, arm64)
- 🔧 **Easy Deployment**: Single binary with no external dependencies

## Quick Start

### Using Docker (Recommended)

```bash
# Run with Docker
docker run -p 8080:8080 ghcr.io/akhfa/myip:latest

# Test the service
curl http://localhost:8080
```

### Using Go Install

```bash
go install github.com/akhfa/myip@latest
myip
```

### Download Binary

Download the latest binary from the [releases page](https://github.com/akhfa/myip/releases).

## API Endpoints

| Endpoint | Description | Response Type |
|----------|-------------|---------------|
| `/` | IPv4 address only | `text/plain` |
| `/ipv6` | IPv6 address only (404 if not available) | `text/plain` |
| `/info` | Detailed IP information | `text/plain` |
| `/json` | Comprehensive JSON response | `application/json` |
| `/headers` | All HTTP headers and IP details | `text/plain` |
| `/health` | Health check endpoint | `application/json` |

### Examples

#### Get IPv4 Address
```bash
$ curl https://ip.example.com/
203.0.113.1
```

#### Get IPv6 Address
```bash
$ curl https://ip.example.com/ipv6
2001:db8::1
```

#### Get Detailed Information
```bash
$ curl https://ip.example.com/info
Your IP Address: 203.0.113.1
Detection Method: CF-Connecting-IP
Is Private IP: false
Behind Cloudflare: true
IPv4 Address: 203.0.113.1
Timestamp: 2023-12-01T12:00:00Z
```

#### Get JSON Response
```bash
$ curl https://ip.example.com/json
{
  "client_ip": "203.0.113.1",
  "detected_via": "CF-Connecting-IP",
  "ipv4_address": "203.0.113.1",
  "ipv6_address": "",
  "is_private_ip": false,
  "is_cloudflare": true,
  "user_agent": "curl/7.68.0",
  "timestamp": "2023-12-01T12:00:00Z"
}
```

## Supported Headers

My IP analyzes the following headers in order of priority:

1. `CF-Connecting-IP` (Cloudflare)
2. `True-Client-IP` (Cloudflare Enterprise)
3. `X-Real-IP` (nginx proxy/FastCGI)
4. `X-Forwarded-For` (Standard proxy header)
5. `X-Client-IP` (Apache mod_proxy_http)
6. `X-Cluster-Client-IP` (Cluster environments)
7. `X-Forwarded`, `Forwarded-For`, `Forwarded` (Less common)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional, for using Makefile)
- Docker (optional, for containerization)

### Local Development

```bash
# Clone the repository
git clone https://github.com/akhfa/myip.git
cd myip

# Install dependencies
make deps

# Run tests
make test

# Run with hot reload (requires air)
make dev

# Or run directly
make run
```

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make test          # Run tests
make test-cover    # Run tests with coverage
make check         # Run all code quality checks
make docker-build  # Build Docker image
make clean         # Clean build artifacts
```

## CI/CD Pipeline

This project features a comprehensive CI/CD pipeline with:

### Pull Request Workflow
- ✅ Automated testing and code quality checks
- 🔍 Static analysis with `staticcheck` and `golint`
- 🏗️ Multi-platform build verification
- 🐳 Docker image build validation
- 📊 Code coverage reporting

### Release Workflow
- 🚀 Automated releases with GoReleaser
- 🐳 Multi-architecture Docker images (amd64, arm64)
- 📦 Package generation (deb, rpm, apk)
- 🔐 Artifact signing with Cosign
- 🛡️ Security scanning with Gosec and Trivy
- 📋 SBOM (Software Bill of Materials) generation

### Supported Package Managers
- 🍺 **Homebrew** (macOS/Linux)
- 📦 **AUR** (Arch Linux)
- 🪟 **Winget** (Windows)
- 🐧 **APT/YUM/APK** (Linux distributions)

For detailed CI/CD documentation, see [docs/CICD.md](docs/CICD.md).

## Docker

### Multi-Architecture Support

Images are available for:
- `linux/amd64`
- `linux/arm64`

### Image Variants

```bash
# Latest stable release
docker pull ghcr.io/akhfa/myip:latest

# Specific version
docker pull ghcr.io/akhfa/myip:v1.0.0

# Development build
docker pull ghcr.io/akhfa/myip:main
```

### Docker Compose

```yaml
version: '3.8'
services:
  myip:
    image: ghcr.io/akhfa/myip:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/myip", "-health-check"]
      interval: 30s
      timeout: 3s
      retries: 3
```

## Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myip
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myip
  template:
    metadata:
      labels:
        app: myip
    spec:
      containers:
      - name: myip
        image: ghcr.io/akhfa/myip:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: myip-service
spec:
  selector:
    app: myip
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Cloudflare Workers

The service works seamlessly behind Cloudflare with proper `CF-Connecting-IP` header detection.

## Security

- 🔒 **Input Validation**: All IP addresses are validated before processing
- 🛡️ **Header Sanitization**: Prevents header injection attacks
- 🚫 **No External Dependencies**: Minimal attack surface
- 📋 **Security Scanning**: Automated vulnerability scanning in CI/CD
- ✍️ **Signed Releases**: All releases are signed with Cosign

## Performance

- ⚡ **Low Latency**: < 1ms response time for simple requests
- 🎯 **Low Memory**: < 10MB memory footprint
- 📈 **High Throughput**: Handles thousands of requests per second
- 🔄 **Concurrent Safe**: Full goroutine safety

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Run the test suite (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the GNU Affero General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 **Documentation**: Check our [docs](docs/) directory
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/akhfa/myip/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/akhfa/myip/discussions)
- 🔒 **Security Issues**: Please email security@example.com

## Acknowledgments

- Built with [Go](https://golang.org/)
- CI/CD powered by [GitHub Actions](https://github.com/features/actions)
- Releases automated with [GoReleaser](https://goreleaser.com/)
- Container images hosted on [GitHub Container Registry](https://ghcr.io/)

---

**Made with ❤️ by the community**

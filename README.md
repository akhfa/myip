# My IP

[![Build Status](https://github.com/akhfa/myip/workflows/Release/badge.svg)](https://github.com/akhfa/myip/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/akhfa/myip)](https://goreportcard.com/report/github.com/akhfa/myip)
[![codecov](https://codecov.io/gh/akhfa/myip/branch/main/graph/badge.svg)](https://codecov.io/gh/akhfa/myip)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=akhfa_myip&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=akhfa_myip)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Pulls](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/akhfa/myip/pkgs/container/myip)

A lightweight, fast HTTP service for detecting client IP addresses with comprehensive proxy header support and detailed IP information.

## Features

- 🌐 **Multi-Protocol Support**: Detects both IPv4 and IPv6 addresses
- 🔍 **Comprehensive Header Analysis**: Supports all major proxy headers (Cloudflare, nginx, Apache, etc.)
- 🏷️ **Multiple Output Formats**: Plain text, JSON, and JSONP endpoints with flexible query parameter support
- 📚 **Interactive API Documentation**: Built-in Swagger UI with OpenAPI specification
- 🛡️ **Security Focused**: Identifies private IPs, proxy chains, and Cloudflare detection
- 🚀 **High Performance**: Lightweight Go implementation with minimal dependencies
- 📊 **Health Monitoring**: Built-in health check endpoint
- 🐳 **Container Ready**: Optimized Docker images (amd64, arm64)
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
| `/?format=json` | IPv4 address in JSON format | `application/json` |
| `/?format=jsonp` | IPv4 address in JSONP format | `application/javascript` |
| `/?format=jsonp&callback=getip` | IPv4 address in JSONP format with custom callback | `application/javascript` |
| `/ipv6` | IPv6 address only (404 if not available) | `text/plain` |
| `/ipv6?format=json` | IPv6 address in JSON format | `application/json` |
| `/ipv6?format=jsonp` | IPv6 address in JSONP format | `application/javascript` |
| `/ipv6?format=jsonp&callback=getip` | IPv6 address in JSONP format with custom callback | `application/javascript` |
| `/info` | Detailed IP information | `text/plain` |
| `/json` | Comprehensive JSON response | `application/json` |
| `/headers` | All HTTP headers and IP details | `text/plain` |
| `/health` | Health check endpoint | `application/json` |
| `/swagger/` | Interactive API documentation | `text/html` |

## API Documentation

This service provides comprehensive API documentation through Swagger/OpenAPI:

- **Interactive Documentation**: Visit `/swagger/` for a web-based API explorer
- **OpenAPI Specification**: Available at `/swagger/doc.json` for programmatic access
- **API Testing**: Use the Swagger UI to test endpoints directly from your browser

### Examples

#### Get IPv4 Address
```bash
$ curl https://ip.example.com/
203.0.113.1
```

#### Get IPv4 Address in JSON Format
```bash
$ curl https://ip.example.com/?format=json
{"ip":"203.0.113.1"}
```

#### Get IPv6 Address
```bash
$ curl https://ip.example.com/ipv6
2001:db8::1
```

#### Get IPv6 Address in JSON Format
```bash
$ curl https://ip.example.com/ipv6?format=json
{"ip":"2001:db8::1"}
```

#### Get IPv4 Address in JSONP Format
```bash
$ curl https://ip.example.com/?format=jsonp
callback({"ip":"203.0.113.1"});

# With custom callback function (requires format=jsonp)
$ curl https://ip.example.com/?format=jsonp&callback=getip
getip({"ip":"203.0.113.1"});
```

#### Get IPv6 Address in JSONP Format
```bash
$ curl https://ip.example.com/ipv6?format=jsonp
callback({"ip":"2001:db8::1"});

# With custom callback function (requires format=jsonp)
$ curl https://ip.example.com/ipv6?format=jsonp&callback=getip
getip({"ip":"2001:db8::1"});
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

#### Access API Documentation
```bash
# Open interactive Swagger UI in browser
open http://localhost:8080/swagger/

# Get OpenAPI specification
curl http://localhost:8080/swagger/doc.json
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
| `HOST` | `localhost:8080` | Host configuration (used internally for server setup) |

## Development

### Prerequisites

- Go 1.24.1 or later
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
make help               # Show all available commands
make swagger            # Generate Swagger documentation
make build              # Build the application (includes swagger)
make run                # Run the application
make dev                # Run with hot reload (requires air)
make test               # Run tests
make test-race          # Run tests with race detector
make test-cover         # Run tests with coverage
make test-coverage-ci   # Run tests with coverage for CI (with race detector)
make bench              # Run benchmarks
make smoke-test         # Run comprehensive smoke tests (manual trigger)
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golint
make staticcheck        # Run staticcheck
make check              # Run all code quality checks
make build-all          # Build for all platforms
make docker-build       # Build Docker image
make docker-test-build  # Build Docker image for testing (no push)
make docker-run         # Run Docker container
make clean              # Clean build artifacts
make deps               # Download and verify dependencies
make tidy               # Tidy dependencies
make install            # Install the application
make security           # Run security checks
make security-sarif     # Run security checks with SARIF output
make ci-setup           # Setup CI environment (install tools)
make ci-test            # Run CI tests (deps, swagger, vet, staticcheck, test-race)
```

### Testing

The project includes comprehensive test coverage with multiple testing strategies:

#### Unit Tests
```bash
# Run all tests
make test

# Run tests with race detector
make test-race

# Run tests with coverage report
make test-cover
```

#### Smoke Tests
The application includes focused smoke tests that validate **IP detection accuracy** against the **live deployment** at `https://ip.2ak.me`:

```bash
# Run smoke tests manually (recommended for deployment validation)
make smoke-test

# Alternative direct command
go test -run TestSmokeTest -v ./test

# Run via GitHub Actions (manual workflow dispatch)
# Go to Actions tab → Smoke Test → Run workflow
```

**Smoke Test Validation:**
- ✅ **IPv4 Detection Accuracy**: Compares your public IPv4 from `api.ipify.org` with deployment detection
- ✅ **IPv6 Detection Accuracy**: Compares your public IPv6 from `api64.ipify.org` with deployment detection (if available)  
- ✅ **JSON Field Validation**: Tests `/json` endpoint `ipv4_address` and `ipv6_address` fields against ipify.org results
- ✅ **Endpoint Accessibility**: Validates all core endpoints (`/health`, `/info`, `/headers`) are accessible

**How It Works:**
1. **Get Real Public IP**: Fetches your actual public IPv4/IPv6 from ipify.org services
2. **Test Deployment**: Accesses `ip.2ak.me/`, `/ipv6`, and `/json` endpoints  
3. **Compare Results**: Validates deployment detects **exactly** the same IP as external services
4. **JSON Field Validation**: Re-retrieves IP from ipify before testing `/json` endpoint and validates `ipv4_address` and `ipv6_address` fields
5. **Strict Matching**: Test **fails** if IPs don't match exactly - no tolerance for differences

**Live Deployment Validation:**
The smoke test performs **strict accuracy validation** against the actual production deployment:
- ✅ **Exact IP Matching**: Requires deployment to detect **identical** IP as ipify.org
- ✅ **IPv4/IPv6 Support**: Tests both protocol versions with exact comparison
- ✅ **JSON Field Validation**: Confirms `/json` endpoint `ipv4_address` and `ipv6_address` fields match external services
- ✅ **Production Environment**: Validates live deployment configuration and accessibility
- ✅ **Network Connectivity**: Confirms real-world network behavior and response times

This focused approach ensures the deployed service **exactly matches** external IP detection services with zero tolerance for differences.

### GitHub Actions Integration

The smoke test can be executed remotely via GitHub Actions using workflow dispatch:

1. **Navigate to Actions Tab**: Go to the repository's Actions tab
2. **Select Smoke Test Workflow**: Click on "Smoke Test" workflow
3. **Run Workflow**: Click "Run workflow" button
4. **Add Description** (optional): Provide a description for the test run
5. **View Results**: Monitor the workflow execution and view detailed logs

**Workflow Features:**
- ✅ **Manual Trigger**: On-demand execution via workflow dispatch
- ✅ **Timeout Protection**: 10-minute maximum execution time  
- ✅ **Detailed Logging**: Comprehensive test output and summaries
- ✅ **Status Reporting**: Clear pass/fail indicators with error notifications
- ✅ **Environment Info**: Shows test description, target URL, and timing

This allows remote validation of the deployment without requiring local setup.

## CI/CD Pipeline

This project features a comprehensive CI/CD pipeline with two main workflows:

### Pull Request Workflow
- 🔧 **Makefile Integration**: All steps use standardized `make` commands for consistency
- 📚 **Swagger Generation**: Automatically generates API documentation via `make build`
- ✅ Comprehensive testing with `make test-coverage-ci` (race detection included)
- 🔍 Static analysis with `make staticcheck` and `make lint`
- 🏗️ Application build with `make build` (includes swagger docs)
- 🐳 Docker image build validation with `make docker-test-build`
- 📊 Code coverage reporting with Codecov integration
- ⚙️ Consistent CI setup with `make ci-setup`

### Release Workflow
- 🔧 **Makefile Integration**: Uses `make` commands for consistent testing and security
- 🚀 **Snapshot builds** on pushes to `main` branch (Docker images only)
- 🎯 **Tagged releases** with full publishing pipeline
- 📚 **Swagger Documentation**: Generated automatically via `make build`
- 🐳 Optimized Docker images (amd64, arm64) on GHCR
- 📦 Package generation (deb, rpm, apk)
- 🔐 Artifact signing with Cosign
- 🛡️ Security scanning with `make security-sarif` and Trivy
- 📋 SBOM (Software Bill of Materials) generation
- 🔄 Automatic package manager publishing

### Supported Package Managers
- 🐧 **APT/YUM/APK** (Linux distributions - deb/rpm/apk packages)

### Container Registry
All Docker images are published to GitHub Container Registry (GHCR):
- `main` and `latest` tags for main branch builds
- `v*` tags for release builds
- `commit-<sha>` tags for specific commits

For detailed CI/CD documentation, see [docs/CICD.md](docs/CICD.md).

## Security & Verification

### Verifying Signatures

All Docker images and release artifacts are signed with Cosign for security verification.

#### Verify Docker Images
```bash
# Download public key
curl -O https://raw.githubusercontent.com/akhfa/myip/main/.cosign/cosign.pub

# Verify latest image
cosign verify --key cosign.pub ghcr.io/akhfa/myip:latest

# Verify specific version
cosign verify --key cosign.pub ghcr.io/akhfa/myip:v1.0.0
```

#### Verify Release Binaries
```bash
# Download release assets (example for v1.0.0)
wget https://github.com/akhfa/myip/releases/download/v1.0.0/checksums.txt
wget https://github.com/akhfa/myip/releases/download/v1.0.0/checksums.txt.sig

# Verify checksum signature
cosign verify-blob --key cosign.pub --signature checksums.txt.sig checksums.txt

# Verify your downloaded binary matches the signed checksum
sha256sum myip_Linux_x86_64.tar.gz
grep myip_Linux_x86_64.tar.gz checksums.txt
```

For complete setup and verification instructions, see [docs/COSIGN_SETUP.md](docs/COSIGN_SETUP.md).

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
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
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

- 🔒 **Input Validation**: Comprehensive IP address validation and sanitization
- 🛡️ **Header Analysis**: Safe parsing of proxy headers with injection protection
- 🚫 **Zero Dependencies**: No external dependencies, minimal attack surface
- � **Private IP Detection**: Identifies private IP ranges (RFC 1918, 3927, 5735, 4193, 4291)
- �📋 **Security Scanning**: Automated vulnerability scanning with Gosec and Trivy in CI/CD
- ✍️ **Signed Releases**: All releases and Docker images are signed with Cosign for integrity verification
- 🛡️ **SARIF Integration**: Security findings integrated with GitHub Security tab
- 🔐 **Signature Verification**: Verify Docker images and binaries with included public key

## Performance

- ⚡ **Low Latency**: Sub-millisecond response times for simple requests
- 🎯 **Low Memory**: Minimal memory footprint (< 10MB)
- 📈 **High Throughput**: Optimized for thousands of concurrent requests
- 🔄 **Concurrent Safe**: Full goroutine safety with no external dependencies
- 🚀 **Zero Dependencies**: Pure Go standard library implementation

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

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 **Documentation**: Check our [docs](docs/) directory
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/akhfa/myip/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/akhfa/myip/discussions)
- 🔒 **Security Issues**: Please report security vulnerabilities via [GitHub Security Advisories](https://github.com/akhfa/myip/security/advisories)

## Acknowledgments

- Built with [Go](https://golang.org/)
- CI/CD powered by [GitHub Actions](https://github.com/features/actions)
- Releases automated with [GoReleaser](https://goreleaser.com/)
- Container images hosted on [GitHub Container Registry](https://ghcr.io/)

---

**Made with ❤️ by the community**

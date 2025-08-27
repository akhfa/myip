# CI/CD Pipeline Documentation

This document describes the comprehensive CI/CD pipeline setup for the My IP Go application using GitHub Actions and GoReleaser.

## Overview

The CI/CD pipeline consists of three main workflows:

1. **Pull Request Workflow** (`.github/workflows/pr.yml`) - Runs on PR creation/updates
2. **Release Workflow** (`.github/workflows/release.yml`) - Runs on pushes to master/main and tags
3. **Smoke Test Workflow** (`.github/workflows/smoke-test.yml`) - Manual deployment validation

## Workflow Details

### Pull Request Workflow

**Triggered on:** Pull requests to `master` or `main` branches

**Jobs:**
- **Test Job**: Runs comprehensive testing using Makefile commands for consistency
- **Docker Build Job**: Builds Docker image using Makefile (without pushing)

### Smoke Test Workflow

**Triggered on:** Manual workflow dispatch (`workflow_dispatch`)

**Purpose:** Validates live deployment accuracy against external IP services

**Jobs:**
- **Smoke Test Job**: Tests production deployment at `https://ip.2ak.me` (runs in separate test package)

**Features:**
- **Live Deployment Testing**: Validates IP detection against external services
- **IPv4/IPv6 Validation**: Tests both `ipv4_address` and `ipv6_address` JSON fields
- **External IP Comparison**: Compares results with `api.ipify.org` services
- **Manual Trigger**: On-demand execution via GitHub Actions interface
- **Strict Validation**: Zero tolerance for IP detection discrepancies

### Pull Request Workflow (continued)

**Features:**
- **Makefile Integration**: All build and test steps use standardized Makefile commands
- **Swagger Documentation**: Automatically generates API documentation before builds
- Go module caching for faster builds
- Static code analysis with `make staticcheck` and `make lint`
- Race condition detection with `make test-coverage-ci`
- Code coverage reporting to Codecov
- Docker image build validation with `make docker-test-build`
- Consistent CI environment setup with `make ci-setup`

### Release Workflow

**Triggered on:** 
- Pushes to `main` branch
- Git tags matching `v*`

**Jobs:**

#### 1. Test Job
- **Makefile Integration**: Uses standardized Makefile commands for all testing
- **Swagger Generation**: Automatically generates API docs via `make build` 
- Comprehensive testing with `make test-coverage-ci` (includes race detection)
- Code quality checks with `make vet`, `make staticcheck`
- Code coverage reporting with Codecov integration
- Consistent CI environment setup with `make ci-setup`

#### 2. Docker Standalone Job (Branch builds only)
- **Condition**: Only runs for pushes to `main` branch (not for tags)
- Builds and pushes Docker images to GitHub Container Registry (GHCR)
- Multi-architecture support (amd64, arm64)
- Automatic tagging with:
  - `main` (for main branch)
  - `commit-<sha>` (for specific commits)
  - `latest` (for default branch)

#### 3. Release Job
- **Condition**: Runs for both tagged releases and main branch pushes
- Uses GoReleaser with different modes:
  - **Tagged releases**: Full release with publishing and announcements
  - **Main branch**: Snapshot builds without publishing
- Creates GitHub releases with release notes (tagged releases only)
- Builds binaries for multiple platforms
- Generates checksums
- Creates package distributions (deb, rpm, apk)
- Publishes to package managers (Homebrew, AUR, Winget) - tagged releases only
- Docker image building integrated via GoReleaser

#### 4. Security Job
- **Makefile Integration**: Uses `make security-sarif` for consistent security scanning
- Runs Gosec security scanner with SARIF output format
- Performs Trivy vulnerability scanning
- Uploads SARIF results to GitHub Security tab

## GoReleaser Configuration

The `.goreleaser.yaml` file defines:

### Build Configuration
- Cross-platform compilation (Linux, macOS, Windows)
- Multiple architectures (amd64, arm64, arm)
- Optimized build flags with version information
- Static linking for minimal dependencies

### Docker Images
- Multi-architecture Docker images
- Automated tagging and versioning
- Image signing with Cosign
- GHCR registry integration

### Release Assets
- Compressed archives (tar.gz, zip)
- Package formats (deb, rpm, apk)
- Checksums and signatures
- Comprehensive release notes

### Package Manager Integration
- **Homebrew**: Automatic formula creation
- **AUR**: Arch Linux package publishing
- **Winget**: Windows package manager integration

## Security Features

### Code Security
- Static analysis with Gosec
- Dependency vulnerability scanning with Trivy
- SARIF report generation for GitHub Security tab

### Artifact Security
- **Artifact signing with Cosign**: All Docker images and release checksums are signed
- **Docker image signing**: Both branch builds and releases are signed with Cosign
- **Release binary signing**: Checksum files are signed for integrity verification
- **Checksum generation**: SHA256 checksums for all releases
- **SBOM generation**: Software Bill of Materials for transparency (planned)

## Usage Instructions

### For Contributors

#### Creating Pull Requests
1. Create a feature branch from `master`/`main`
2. Make your changes and commit
3. Push to your branch and create a PR
4. The PR workflow will automatically run all checks
5. Address any failing checks before merging

#### Local Development
Use the provided Makefile for local development:

```bash
# Setup CI environment (install all tools)
make ci-setup

# Run all checks locally
make check

# Run tests with coverage (CI version with race detector)
make test-coverage-ci

# Build the application (includes swagger generation)
make build

# Generate swagger documentation only
make swagger

# Run in development mode
make dev

# Test Docker build
make docker-test-build

# Run security checks
make security
```

### For Maintainers

#### Creating Releases

##### Automatic Snapshot Builds
1. Merge PRs to `main` branch
2. The release workflow automatically creates snapshot builds
3. Docker images are built and pushed to GHCR with `main` and `latest` tags
4. No GitHub release is created (snapshot mode only)

##### Tagged Releases (Full Release)
1. Create and push a git tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
2. The release workflow will trigger and create a full GitHub release
3. Publishes to package managers and creates release assets

#### Release Artifacts

**Tagged Releases** include:
- **Binaries**: For Linux, macOS, and Windows (multiple architectures)
- **Docker Images**: Multi-architecture images on GHCR with Cosign signatures
- **Packages**: deb, rpm, and apk packages
- **Checksums**: SHA256 checksums for all artifacts with Cosign signatures
- **Security Verification**: Cosign signatures for Docker images and checksums
- **GitHub Release**: With comprehensive release notes

**Snapshot Builds** (main branch) include:
- **Docker Images**: Multi-architecture images on GHCR (tagged as `main`, `latest`, and `commit-<sha>`)
- **Binaries**: Built but not published (available as workflow artifacts)

## Configuration Requirements

### GitHub Repository Settings

#### Secrets
No additional secrets are required. The workflow uses:
- `GITHUB_TOKEN` (automatically provided)

#### Permissions
The following permissions are automatically configured:
- `contents: write` (for creating releases)
- `packages: write` (for pushing Docker images)
- `security-events: write` (for security scanning)

#### Branch Protection
Recommended branch protection rules for `master`/`main`:
- Require PR reviews
- Require status checks to pass
- Require branches to be up to date
- Restrict force pushes

### Cosign Signing Setup (Required for Maintainers)

#### Initial Setup
1. Generate Cosign key pair using the instructions in [COSIGN_SETUP.md](COSIGN_SETUP.md)
2. Add the following GitHub repository secrets:
   - `COSIGN_PRIVATE_KEY`: The entire private key content
   - `COSIGN_PASSWORD`: The password for the private key
3. Commit the public key to `.cosign/cosign.pub` in the repository

#### Signature Verification
```bash
# Verify Docker images
cosign verify --key .cosign/cosign.pub ghcr.io/akhfa/myip:latest

# Verify release checksums
cosign verify-blob --key .cosign/cosign.pub --signature checksums.txt.sig checksums.txt
```

For complete setup instructions, see [COSIGN_SETUP.md](COSIGN_SETUP.md).

### Package Manager Setup (Optional)

#### Homebrew
1. Create a `homebrew-tap` repository
2. The workflow will automatically create pull requests for new releases

#### AUR
1. Set up AUR SSH key as `AUR_KEY` secret
2. Configure AUR repository access

#### Winget
1. Fork the `microsoft/winget-pkgs` repository
2. The workflow will create pull requests automatically

## Docker Usage

### Pulling Images
```bash
# Latest version
docker pull ghcr.io/[username]/myip:latest

# Specific version
docker pull ghcr.io/[username]/myip:v1.0.0
```

### Running Containers
```bash
# Basic usage
docker run -p 8080:8080 ghcr.io/[username]/myip:latest

# With environment variables
docker run -p 8080:8080 -e PORT=3000 ghcr.io/[username]/myip:latest
```

## Monitoring and Troubleshooting

### Workflow Monitoring
- Check the Actions tab for workflow status
- Review logs for failed builds
- Monitor security alerts in the Security tab

### Common Issues

#### Test Failures
- Review test output in the workflow logs
- Run tests locally with `make test` or `make test-cover`
- Ensure all code quality checks pass

#### Docker Build Failures
- Check Dockerfile syntax
- Verify base image availability
- Test builds locally with `make docker-build`

#### Release Failures
- Verify GoReleaser configuration with `make release-dry`
- Check for missing dependencies or invalid configurations
- Ensure proper tagging for releases

### Performance Optimization

#### Build Caching
- Go module caching reduces build times
- Docker layer caching optimizes image builds
- Artifact caching speeds up workflows

#### Parallel Execution
- Jobs run in parallel where possible
- Matrix builds for multiple Go versions
- Independent test and build processes

## Best Practices

### Code Quality
- Always run `make check` before committing
- Maintain test coverage above 80%
- Follow Go coding standards and conventions

### Security
- Regularly update dependencies
- Monitor security advisories
- Review and address security scan results

### Releases
- Use semantic versioning for tags
- Write comprehensive release notes
- Test releases in staging environments

### Documentation
- Keep CI/CD documentation updated
- Document any workflow modifications
- Maintain clear commit messages and PR descriptions

## Future Enhancements

### Planned Features
- Integration testing with test environments
- Performance benchmarking in CI
- Multi-environment deployment pipelines
- Enhanced security scanning and reporting

### Monitoring Integration
- Build status badges
- Performance metrics collection
- Automated dependency updates with Dependabot

This comprehensive CI/CD setup ensures code quality, security, and reliable automated releases while maintaining flexibility for future enhancements.

# CI/CD Pipeline Documentation

This document describes the comprehensive CI/CD pipeline setup for the IP Detector Go application using GitHub Actions and GoReleaser.

## Overview

The CI/CD pipeline consists of two main workflows:

1. **Pull Request Workflow** (`.github/workflows/pr.yml`) - Runs on PR creation/updates
2. **Release Workflow** (`.github/workflows/release.yml`) - Runs on pushes to master/main and tags

## Workflow Details

### Pull Request Workflow

**Triggered on:** Pull requests to `master` or `main` branches

**Jobs:**
- **Test Job**: Runs comprehensive testing including unit tests, linting, and code quality checks
- **Build Job**: Cross-compiles binaries for multiple platforms
- **Docker Build Job**: Builds Docker image (without pushing)

**Features:**
- Go module caching for faster builds
- Static code analysis with `staticcheck` and `golint`
- Race condition detection
- Code coverage reporting to Codecov
- Multi-platform binary compilation
- Docker image build validation

### Release Workflow

**Triggered on:** 
- Pushes to `master`/`main` branches
- Git tags matching `v*`

**Jobs:**

#### 1. Test Job
- Comprehensive testing identical to PR workflow
- Code coverage reporting

#### 2. Docker Job
- Builds and pushes Docker images to GitHub Container Registry (GHCR)
- Multi-architecture support (amd64, arm64)
- Automatic tagging with:
  - `latest` (for default branch)
  - `commit-<sha>` (for specific commits)
  - `v<version>` (for tagged releases)

#### 3. Release Job
- Uses GoReleaser for automated releases
- Creates GitHub releases with release notes
- Builds binaries for multiple platforms
- Generates checksums
- Creates package distributions (deb, rpm, apk)
- Publishes to package managers (Homebrew, AUR, Winget)
- Signs artifacts with Cosign

#### 4. Security Job
- Runs Gosec security scanner
- Performs Trivy vulnerability scanning
- Uploads results to GitHub Security tab

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
- Artifact signing with Cosign
- Docker image signing
- Checksum generation for all releases
- SBOM (Software Bill of Materials) generation

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
# Run all checks locally
make check

# Run tests with coverage
make test-cover

# Build the application
make build

# Run in development mode
make dev
```

### For Maintainers

#### Creating Releases

##### Automatic Releases (Recommended)
1. Merge PRs to `master`/`main` branch
2. The release workflow automatically creates releases
3. Docker images are built and pushed to GHCR

##### Manual Tagged Releases
1. Create and push a git tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
2. The release workflow will trigger and create a full release

#### Release Artifacts

Each release includes:
- **Binaries**: For Linux, macOS, and Windows (multiple architectures)
- **Docker Images**: Multi-architecture images on GHCR
- **Packages**: deb, rpm, and apk packages
- **Checksums**: SHA256 checksums for all artifacts
- **Signatures**: Cosign signatures for security verification

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
docker pull ghcr.io/[username]/ip-detector:latest

# Specific version
docker pull ghcr.io/[username]/ip-detector:v1.0.0
```

### Running Containers
```bash
# Basic usage
docker run -p 8080:8080 ghcr.io/[username]/ip-detector:latest

# With environment variables
docker run -p 8080:8080 -e PORT=3000 ghcr.io/[username]/ip-detector:latest
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

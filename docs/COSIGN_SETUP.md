# Cosign Signing Setup Guide

This guide explains how to set up Cosign signing for the MyIP project, including key generation, GitHub Secrets configuration, and signature verification.

## Overview

The MyIP project uses [Cosign](https://github.com/sigstore/cosign) to sign:
- **Docker images** (both branch builds and releases)
- **Release binaries** (checksums and archives)

All signatures are created using a private key stored securely in GitHub Secrets.

## Prerequisites

- [Cosign CLI](https://docs.sigstore.dev/cosign/installation/) installed locally
- Repository admin access to configure GitHub Secrets
- Basic understanding of public key cryptography

## 1. Generate Cosign Key Pair

### Option A: Generate Keys Locally (Recommended)

```bash
# Generate a new key pair
cosign generate-key-pair

# This creates:
# - cosign.key (private key - keep secret!)
# - cosign.pub (public key - can be shared)
```

When prompted:
- **Enter password**: Use a strong password (you'll add this to GitHub Secrets)
- **Confirm password**: Re-enter the same password

### Option B: Generate Keys without Password (Less Secure)

```bash
# Generate keys without password protection
COSIGN_PASSWORD="" cosign generate-key-pair
```

⚠️ **Security Warning**: Passwordless keys are less secure. Use only for testing.

## 2. Configure GitHub Repository Secrets

Navigate to your GitHub repository → Settings → Secrets and variables → Actions

### Required Secrets

Add these repository secrets:

#### `COSIGN_PRIVATE_KEY`
```bash
# Copy the entire private key file content
cat cosign.key
```
Copy the **entire output** including the `-----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----` and `-----END ENCRYPTED COSIGN PRIVATE KEY-----` lines.

#### `COSIGN_PASSWORD`
Enter the password you used when generating the key pair.

### Secret Configuration Steps

1. Go to your repository on GitHub
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add `COSIGN_PRIVATE_KEY`:
   - Name: `COSIGN_PRIVATE_KEY`
   - Secret: Paste the entire content of `cosign.key`
5. Add `COSIGN_PASSWORD`:
   - Name: `COSIGN_PASSWORD`
   - Secret: Your key password

## 3. Distribute Public Key

The public key (`cosign.pub`) should be made available for signature verification.

### Option A: Commit to Repository (Recommended)

```bash
# Add public key to repository
mkdir -p .cosign
cp cosign.pub .cosign/cosign.pub
git add .cosign/cosign.pub
git commit -m "Add Cosign public key for signature verification"
git push
```

### Option B: Include in Documentation

Add the public key content to your README or documentation:

```bash
cat cosign.pub
```

## 4. Test the Setup

### Test Key Generation
```bash
# Verify key files exist and are valid
cosign public-key --key cosign.key

# Should output the public key content
```

### Test Signing Locally
```bash
# Create a test file
echo "test content" > test.txt

# Sign the file
cosign sign-blob --key cosign.key test.txt

# Verify the signature
cosign verify-blob --key cosign.pub --signature test.txt.sig test.txt
```

## 5. How Signing Works in CI/CD

### Docker Images
1. **Branch builds** (main): Signs images after Docker build
2. **Tagged releases**: Signs images via GoReleaser

### Binary Releases
1. GoReleaser creates checksums of all release artifacts
2. Cosign signs the checksum file
3. Signature is uploaded as a release asset

### Workflow Integration
```yaml
# Example of how signing is integrated
- name: Install Cosign
  uses: sigstore/cosign-installer@v3
  with:
    cosign-release: 'v2.2.4'

- name: Sign Docker images
  env:
    COSIGN_PRIVATE_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}
    COSIGN_PASSWORD: ${{ secrets.COSIGN_PASSWORD }}
  run: |
    cosign sign --yes --key env://COSIGN_PRIVATE_KEY "$image"
```

## 6. Verifying Signatures

### Verify Docker Images

```bash
# Download the public key
curl -O https://raw.githubusercontent.com/[username]/myip/main/.cosign/cosign.pub

# Verify a Docker image signature
cosign verify --key cosign.pub ghcr.io/[username]/myip:latest
```

### Verify Release Binaries

```bash
# Download release assets
wget https://github.com/[username]/myip/releases/download/v1.0.0/checksums.txt
wget https://github.com/[username]/myip/releases/download/v1.0.0/checksums.txt.sig

# Verify checksum signature
cosign verify-blob --key cosign.pub --signature checksums.txt.sig checksums.txt

# Verify specific binary checksum
sha256sum myip_Linux_x86_64.tar.gz
grep myip_Linux_x86_64.tar.gz checksums.txt
```

## 7. Security Best Practices

### Key Management
- 🔐 **Store private keys securely**: Never commit private keys to repositories
- 🔑 **Use strong passwords**: For encrypted private keys
- 🔄 **Rotate keys regularly**: Generate new keys periodically
- 📋 **Backup keys securely**: Store copies in secure locations

### Repository Security
- 🛡️ **Restrict secret access**: Only trusted maintainers should have access
- 📊 **Monitor secret usage**: Review Actions logs for unusual activity
- 🔍 **Audit regularly**: Check which secrets are configured and used

### Verification
- ✅ **Always verify signatures**: Before using downloaded binaries
- 🔗 **Use trusted sources**: Download public keys from official repositories
- 📝 **Document verification steps**: Include in your project documentation

## 8. Troubleshooting

### Common Issues

#### "private key not found" Error
```bash
# Check if COSIGN_PRIVATE_KEY secret is set correctly
# Ensure the entire key content is copied, including headers
```

#### "invalid password" Error
```bash
# Verify COSIGN_PASSWORD secret matches the key password
# Try regenerating keys if password is forgotten
```

#### "permission denied" Errors
```bash
# Check GitHub Actions permissions:
# - contents: write (for releases)
# - packages: write (for container registry)
```

### Debug Commands

```bash
# Test key access locally
export COSIGN_PRIVATE_KEY="$(cat cosign.key)"
export COSIGN_PASSWORD="your-password"
echo "test" | cosign sign-blob --key env://COSIGN_PRIVATE_KEY -

# Verify public key format
cosign public-key --key cosign.key
```

## 9. Advanced Configuration

### Keyless Signing (Future)
Consider migrating to [keyless signing](https://docs.sigstore.dev/cosign/keyless/) for enhanced security:

```yaml
# Keyless signing example
- name: Sign with keyless mode
  run: |
    cosign sign --yes ghcr.io/[username]/myip:latest
```

### Multiple Signatures
For enhanced security, consider implementing multiple signatures from different maintainers.

### Integration with SLSA
Combine Cosign with [SLSA](https://slsa.dev/) for supply chain security.

## 10. Resources

- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [Sigstore Project](https://www.sigstore.dev/)
- [GoReleaser Cosign Integration](https://goreleaser.com/customization/sign/)
- [GitHub Actions Security](https://docs.github.com/en/actions/security-guides)

## Support

If you encounter issues with Cosign setup:
1. Check the [troubleshooting section](#8-troubleshooting) above
2. Review GitHub Actions logs for detailed error messages
3. Open an issue in the repository with error details
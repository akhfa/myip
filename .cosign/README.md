# Cosign Public Key

This directory contains the public key used to verify Cosign signatures for MyIP releases and Docker images.

## Usage

Download the public key and verify signatures:

```bash
# Download public key
curl -O https://raw.githubusercontent.com/akhfa/myip/main/.cosign/cosign.pub

# Verify Docker image
cosign verify --key cosign.pub ghcr.io/akhfa/myip:latest

# Verify release checksums
cosign verify-blob --key cosign.pub --signature checksums.txt.sig checksums.txt
```

## Key Information

- **Algorithm**: ECDSA-P256
- **Purpose**: Signing Docker images and release artifacts
- **Maintained by**: Repository maintainers

## Security

This public key is safe to distribute and use for verification. The corresponding private key is securely stored in GitHub Secrets and never exposed.

For detailed setup and verification instructions, see [COSIGN_SETUP.md](../docs/COSIGN_SETUP.md).
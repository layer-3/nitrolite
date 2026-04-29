# GCP Cloud KMS Signer

This package provides a `sign.Signer` implementation backed by Google Cloud KMS. The private key never leaves the HSM — only signing requests are sent to the KMS API.

## Key Setup

### 1. Create a Key Ring and Key

```bash
# Set your project and location
PROJECT_ID="your-project-id"
LOCATION="us-east1"
KEY_RING="nitronode"
KEY_NAME="signer"

# Create key ring
gcloud kms keyrings create $KEY_RING \
  --location $LOCATION \
  --project $PROJECT_ID

# Create secp256k1 signing key with HSM protection
gcloud kms keys create $KEY_NAME \
  --keyring $KEY_RING \
  --location $LOCATION \
  --project $PROJECT_ID \
  --purpose asymmetric-signing \
  --default-algorithm ec-sign-secp256k1-sha256 \
  --protection-level hsm
```

### 2. Get the Key Version Resource Name

```bash
# List key versions
gcloud kms keys versions list \
  --key $KEY_NAME \
  --keyring $KEY_RING \
  --location $LOCATION \
  --project $PROJECT_ID

# The resource name follows this format:
# projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}/cryptoKeyVersions/{version}
```

### 3. Get the Ethereum Address

```bash
# Fetch the public key
gcloud kms keys versions get-public-key 1 \
  --key $KEY_NAME \
  --keyring $KEY_RING \
  --location $LOCATION \
  --project $PROJECT_ID \
  --output-file /tmp/kms-pub.pem

# The Ethereum address can be derived from the public key using standard tools.
# Nitronode will log the address on startup.
```

## IAM Permissions

The service account running nitronode needs these permissions on the KMS key:

- `cloudkms.cryptoKeyVersions.useToSign` — to sign data
- `cloudkms.cryptoKeyVersions.viewPublicKey` — to fetch the public key at startup

You can grant these with the predefined role:

```bash
SERVICE_ACCOUNT="nitronode-sa@${PROJECT_ID}.iam.gserviceaccount.com"

gcloud kms keys add-iam-policy-binding $KEY_NAME \
  --keyring $KEY_RING \
  --location $LOCATION \
  --project $PROJECT_ID \
  --member "serviceAccount:${SERVICE_ACCOUNT}" \
  --role "roles/cloudkms.signerVerifier"
```

## Nitronode Configuration

Set these environment variables:

```bash
# Use GCP KMS instead of a raw private key
CLEARNODE_SIGNER_TYPE=gcp-kms

# Full key version resource name
NITRONODE_GCP_KMS_KEY_NAME=projects/my-project/locations/us-east1/keyRings/nitronode/cryptoKeys/signer/cryptoKeyVersions/1
```

When running on GKE with Workload Identity, no additional credential configuration is needed. For other environments, set `GOOGLE_APPLICATION_CREDENTIALS` to point to a service account key file.

## Helm Values Example

```yaml
config:
  extraEnvs:
    CLEARNODE_SIGNER_TYPE: gcp-kms
    NITRONODE_GCP_KMS_KEY_NAME: projects/my-project/locations/us-east1/keyRings/nitronode/cryptoKeys/signer/cryptoKeyVersions/1
```

## Key Requirements

| Property         | Value                          |
|-----------------|--------------------------------|
| Algorithm       | `EC_SIGN_SECP256K1_SHA256`    |
| Purpose         | `ASYMMETRIC_SIGN`             |
| Protection Level| `HSM` (recommended)           |
| Curve           | secp256k1                      |

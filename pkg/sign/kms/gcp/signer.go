// Package gcpkms implements [sign.Signer] using Google Cloud KMS.
//
// It wraps a GCP Cloud KMS asymmetric signing key (secp256k1) and converts
// the KMS-produced DER signatures into Ethereum-compatible 65-byte format.
package gcp

import (
	"context"
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"hash/crc32"
	"math/big"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/layer-3/nitrolite/pkg/sign"
	kmssign "github.com/layer-3/nitrolite/pkg/sign/kms"
)

// crc32cTable is pre-computed once to avoid re-creating it on every Sign() call.
var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

// GCPKMSSigner implements [sign.Signer] using a GCP Cloud KMS secp256k1 key.
//
// The key must be created with:
//   - Algorithm: EC_SIGN_SECP256K1_SHA256
//   - Purpose: ASYMMETRIC_SIGN
//   - Protection level: HSM (recommended)
type GCPKMSSigner struct {
	client      *kms.KeyManagementClient
	keyName     string // full resource name including version
	publicKey   sign.EthereumPublicKey
	ecPublicKey *ecdsa.PublicKey
}

// DefaultGRPCPoolSize is the number of gRPC connections to open to KMS.
// Multiple connections allow concurrent signing requests to avoid head-of-line
// blocking on a single HTTP/2 connection.
const DefaultGRPCPoolSize = 4

// NewSigner creates a new GCP KMS signer.
//
// keyResourceName must be the full key version resource name:
// projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}/cryptoKeyVersions/{version}
//
// Authentication is handled automatically by the GCP SDK via Application Default
// Credentials (GOOGLE_APPLICATION_CREDENTIALS env var, Workload Identity, etc.).
func NewSigner(ctx context.Context, keyResourceName string, opts ...option.ClientOption) (*GCPKMSSigner, error) {
	// Default to a connection pool for better concurrent throughput.
	allOpts := append([]option.ClientOption{
		option.WithGRPCConnectionPool(DefaultGRPCPoolSize),
	}, opts...)
	client, err := kms.NewKeyManagementClient(ctx, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	signer, err := newSignerWithClient(ctx, client, keyResourceName)
	if err != nil {
		client.Close()
		return nil, err
	}
	return signer, nil
}

// newSignerWithClient creates a signer with an injected KMS client (for testing).
func newSignerWithClient(ctx context.Context, client *kms.KeyManagementClient, keyResourceName string) (*GCPKMSSigner, error) {
	// Fetch the public key from KMS
	resp, err := client.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{
		Name: keyResourceName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get public key from KMS: %w", err)
	}

	// Verify CRC32C integrity of the public key PEM
	if resp.PemCrc32C != nil {
		expectedCRC := crc32c([]byte(resp.Pem))
		if int64(expectedCRC) != resp.PemCrc32C.Value {
			return nil, fmt.Errorf("KMS public key PEM CRC32C mismatch: got %d, expected %d", resp.PemCrc32C.Value, expectedCRC)
		}
	}

	// Parse PEM-encoded public key
	ecPub, err := parseECPublicKeyPEM(resp.Pem)
	if err != nil {
		return nil, fmt.Errorf("failed to parse KMS public key: %w", err)
	}

	// Validate it's on secp256k1
	if err := kmssign.ValidateSecp256k1PublicKey(ecPub); err != nil {
		return nil, fmt.Errorf("KMS key validation failed: %w", err)
	}

	return &GCPKMSSigner{
		client:      client,
		keyName:     keyResourceName,
		publicKey:   sign.NewEthereumPublicKey(ecPub),
		ecPublicKey: ecPub,
	}, nil
}

// PublicKey returns the cached Ethereum public key derived from the KMS key.
func (s *GCPKMSSigner) PublicKey() sign.PublicKey {
	return s.publicKey
}

// signTimeout is the maximum time to wait for a KMS signing operation.
const signTimeout = 15 * time.Second

// Sign signs the given hash using GCP KMS and returns an Ethereum-compatible signature.
//
// The hash should be a 32-byte digest (e.g., Keccak256). GCP KMS will sign this
// hash directly using the secp256k1 key. The returned DER signature is then
// converted to Ethereum's 65-byte R || S || V format.
//
// CRC32C integrity checks are performed on both the request digest and response
// signature to detect data corruption in transit.
func (s *GCPKMSSigner) Sign(hash []byte) (sign.Signature, error) {
	ctx, cancel := context.WithTimeout(context.Background(), signTimeout)
	defer cancel()

	digestCRC32C := crc32c(hash)

	resp, err := s.client.AsymmetricSign(ctx, &kmspb.AsymmetricSignRequest{
		Name: s.keyName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: hash,
			},
		},
		DigestCrc32C: wrapperspb.Int64(int64(digestCRC32C)),
	})
	if err != nil {
		return nil, fmt.Errorf("KMS AsymmetricSign failed: %w", err)
	}

	// Verify the server confirmed our digest checksum
	if !resp.VerifiedDigestCrc32C {
		return nil, fmt.Errorf("KMS did not verify digest CRC32C, request may have been corrupted in transit")
	}

	// Verify the response signature checksum
	if resp.SignatureCrc32C != nil {
		expectedCRC := crc32c(resp.Signature)
		if int64(expectedCRC) != resp.SignatureCrc32C.Value {
			return nil, fmt.Errorf("KMS response signature CRC32C mismatch: got %d, expected %d", resp.SignatureCrc32C.Value, expectedCRC)
		}
	}

	ethSig, err := kmssign.DERToEthereumSignature(hash, resp.Signature, s.ecPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert KMS signature to Ethereum format: %w", err)
	}

	return sign.Signature(ethSig), nil
}

// Close closes the underlying KMS client connection.
func (s *GCPKMSSigner) Close() error {
	return s.client.Close()
}

// secp256k1OID is the ASN.1 OID for the secp256k1 curve (1.3.132.0.10).
var secp256k1OID = asn1.ObjectIdentifier{1, 3, 132, 0, 10}

// subjectPublicKeyInfo mirrors the ASN.1 SubjectPublicKeyInfo structure.
type subjectPublicKeyInfo struct {
	Algorithm algorithmIdentifier
	PublicKey asn1.BitString
}

type algorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.ObjectIdentifier
}

// parseECPublicKeyPEM parses a PEM-encoded secp256k1 EC public key.
//
// Go's standard x509.ParsePKIXPublicKey does not support secp256k1 (OID 1.3.132.0.10),
// so we parse the ASN.1 SubjectPublicKeyInfo manually and extract the uncompressed
// EC point (0x04 || X || Y).
func parseECPublicKeyPEM(pemStr string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	var spki subjectPublicKeyInfo
	if _, err := asn1.Unmarshal(block.Bytes, &spki); err != nil {
		return nil, fmt.Errorf("failed to parse SubjectPublicKeyInfo: %w", err)
	}

	if !spki.Algorithm.Parameters.Equal(secp256k1OID) {
		return nil, fmt.Errorf("unsupported curve OID %v, expected secp256k1 (%v)", spki.Algorithm.Parameters, secp256k1OID)
	}

	pointBytes := spki.PublicKey.Bytes
	if len(pointBytes) == 0 {
		return nil, fmt.Errorf("empty public key point")
	}

	// Expect uncompressed format: 0x04 || X (32 bytes) || Y (32 bytes)
	if pointBytes[0] != 0x04 || len(pointBytes) != 65 {
		return nil, fmt.Errorf("expected uncompressed EC point (65 bytes), got %d bytes with prefix 0x%02x", len(pointBytes), pointBytes[0])
	}

	curve := secp256k1.S256()
	x := new(big.Int).SetBytes(pointBytes[1:33])
	y := new(big.Int).SetBytes(pointBytes[33:65])

	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("public key point is not on the secp256k1 curve")
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// crc32c computes the CRC32C (Castagnoli) checksum of data.
func crc32c(data []byte) uint32 {
	return crc32.Checksum(data, crc32cTable)
}

package gcp

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/pem"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/sign"
)

// buildSecp256k1PEM creates a PEM-encoded SubjectPublicKeyInfo for a secp256k1 key,
// matching the format returned by GCP KMS.
func buildSecp256k1PEM(pub *ecdsa.PublicKey) string {
	// EC public key OID: 1.2.840.10045.2.1
	ecOID := asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

	// Build uncompressed point: 0x04 || X (32 bytes) || Y (32 bytes)
	point := make([]byte, 65)
	point[0] = 0x04
	xBytes := pub.X.Bytes()
	yBytes := pub.Y.Bytes()
	copy(point[33-len(xBytes):33], xBytes)
	copy(point[65-len(yBytes):65], yBytes)

	spki := subjectPublicKeyInfo{
		Algorithm: algorithmIdentifier{
			Algorithm:  ecOID,
			Parameters: secp256k1OID,
		},
		PublicKey: asn1.BitString{
			Bytes:     point,
			BitLength: len(point) * 8,
		},
	}

	der, err := asn1.Marshal(spki)
	if err != nil {
		panic(err)
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}))
}

func TestParseECPublicKeyPEM_Secp256k1(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	pemStr := buildSecp256k1PEM(&key.PublicKey)

	pub, err := parseECPublicKeyPEM(pemStr)
	require.NoError(t, err)
	assert.Equal(t, key.PublicKey.X, pub.X)
	assert.Equal(t, key.PublicKey.Y, pub.Y)
	assert.True(t, secp256k1.S256().IsOnCurve(pub.X, pub.Y))
}

func TestParseECPublicKeyPEM_WrongCurveOID(t *testing.T) {
	p256OID := asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
	ecOID := asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

	point := make([]byte, 65)
	point[0] = 0x04

	spki := subjectPublicKeyInfo{
		Algorithm: algorithmIdentifier{
			Algorithm:  ecOID,
			Parameters: p256OID,
		},
		PublicKey: asn1.BitString{
			Bytes:     point,
			BitLength: len(point) * 8,
		},
	}

	der, err := asn1.Marshal(spki)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})

	_, err = parseECPublicKeyPEM(string(pemBlock))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported curve OID")
}

func TestParseECPublicKeyPEM_InvalidPEM(t *testing.T) {
	_, err := parseECPublicKeyPEM("not a PEM block")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode PEM")
}

func TestParseECPublicKeyPEM_InvalidASN1(t *testing.T) {
	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte("not valid ASN.1"),
	})

	_, err := parseECPublicKeyPEM(string(pemBlock))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse SubjectPublicKeyInfo")
}

func TestParseECPublicKeyPEM_CompressedPoint(t *testing.T) {
	// Build a PEM with a compressed point (33 bytes, prefix 0x02/0x03) — should fail
	ecOID := asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	compressed := make([]byte, 33)
	compressed[0] = 0x02

	spki := subjectPublicKeyInfo{
		Algorithm: algorithmIdentifier{
			Algorithm:  ecOID,
			Parameters: secp256k1OID,
		},
		PublicKey: asn1.BitString{
			Bytes:     compressed,
			BitLength: len(compressed) * 8,
		},
	}

	der, err := asn1.Marshal(spki)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})

	_, err = parseECPublicKeyPEM(string(pemBlock))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected uncompressed EC point")
}

func TestParseECPublicKeyPEM_PointNotOnCurve(t *testing.T) {
	ecOID := asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

	// 0x04 || X || Y where X,Y are not on secp256k1
	point := make([]byte, 65)
	point[0] = 0x04
	point[1] = 0x01 // X = 1, Y = 1 — not on curve
	point[33] = 0x01

	spki := subjectPublicKeyInfo{
		Algorithm: algorithmIdentifier{
			Algorithm:  ecOID,
			Parameters: secp256k1OID,
		},
		PublicKey: asn1.BitString{
			Bytes:     point,
			BitLength: len(point) * 8,
		},
	}

	der, err := asn1.Marshal(spki)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})

	_, err = parseECPublicKeyPEM(string(pemBlock))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not on the secp256k1 curve")
}

func TestCrc32c(t *testing.T) {
	// Verify crc32c produces non-zero, deterministic values
	data := []byte("hello world")
	c1 := crc32c(data)
	c2 := crc32c(data)
	assert.Equal(t, c1, c2)
	assert.NotZero(t, c1)

	// Different data produces different checksum
	other := crc32c([]byte("different data"))
	assert.NotEqual(t, c1, other)
}

func TestGCPKMSSigner_PublicKey(t *testing.T) {
	// Verify the public key type round-trips correctly through our PEM parser
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	pemStr := buildSecp256k1PEM(&key.PublicKey)
	pub, err := parseECPublicKeyPEM(pemStr)
	require.NoError(t, err)

	ethPub := sign.NewEthereumPublicKey(pub)

	// Verify address derivation works
	addr := ethPub.Address()
	assert.NotEmpty(t, addr.String())

	// Verify X,Y coordinates match
	assert.Equal(t, key.PublicKey.X.Cmp(pub.X), 0)
	assert.Equal(t, key.PublicKey.Y.Cmp(pub.Y), 0)
}

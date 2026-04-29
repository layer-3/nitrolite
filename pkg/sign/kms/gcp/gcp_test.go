// Integration test for GCP KMS signer against a real KMS key.
//
// This test is skipped by default. To run it locally:
//
//	export GCP_KMS_KEY_NAME="projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}/cryptoKeyVersions/{version}"
//	export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
//	go test -v -run TestIntegration ./pkg/sign/kms/gcp/
//
// The key must be EC_SIGN_SECP256K1_SHA256. The service account needs
// roles/cloudkms.signerVerifier on the key.
package gcp

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/sign"
)

const envKMSKeyName = "GCP_KMS_KEY_NAME"

func skipIfNoKMS(t *testing.T) string {
	t.Helper()
	keyName := os.Getenv(envKMSKeyName)
	if keyName == "" {
		t.Skipf("skipping: set %s to run GCP KMS integration tests", envKMSKeyName)
	}
	return keyName
}

// TestIntegration_KMSSignerCreation verifies that the KMS signer initializes
// successfully and returns a valid Ethereum address.
func TestIntegration_KMSSignerCreation(t *testing.T) {
	keyName := skipIfNoKMS(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	signer, err := NewSigner(ctx, keyName)
	require.NoError(t, err)
	defer signer.Close()

	addr := signer.PublicKey().Address().String()
	t.Logf("KMS signer address: %s", addr)

	assert.True(t, strings.HasPrefix(addr, "0x"), "address should start with 0x")
	assert.Len(t, addr, 42, "Ethereum address should be 42 chars")
}

// TestIntegration_RawSignAndRecover signs a hash with the KMS signer (raw mode)
// and verifies the signature recovers to the KMS key's address.
func TestIntegration_RawSignAndRecover(t *testing.T) {
	keyName := skipIfNoKMS(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kmsSigner, err := NewSigner(ctx, keyName)
	require.NoError(t, err)
	defer kmsSigner.Close()

	kmsAddress := kmsSigner.PublicKey().Address().String()
	t.Logf("KMS address: %s", kmsAddress)

	messages := []struct {
		name string
		data []byte
	}{
		{"simple text", []byte("hello world")},
		{"empty", []byte("")},
		{"json payload", []byte(`{"channel":"0x123","amount":"100"}`)},
		{"binary", []byte{0x00, 0x01, 0xff, 0xfe}},
		{"long message", []byte(strings.Repeat("a]", 1000))},
	}

	for _, msg := range messages {
		t.Run(msg.name, func(t *testing.T) {
			hash := ethcrypto.Keccak256(msg.data)

			// Sign with KMS
			kmsSig, err := kmsSigner.Sign(hash)
			require.NoError(t, err)
			require.Len(t, kmsSig, 65, "signature should be 65 bytes")

			// V should be 27 or 28
			v := kmsSig[64]
			assert.True(t, v == 27 || v == 28, "V should be 27 or 28, got %d", v)

			// Recover address from KMS signature
			recoveredAddr, err := sign.RecoverAddressFromHash(hash, kmsSig)
			require.NoError(t, err)

			assert.True(t,
				strings.EqualFold(kmsAddress, recoveredAddr.String()),
				"recovered address %s should match KMS address %s", recoveredAddr, kmsAddress,
			)
		})
	}
}

// TestIntegration_MsgSignAndRecover wraps the KMS signer in EthereumMsgSigner
// (EIP-191) and verifies the signature recovers correctly using the matching
// EthereumMsgAddressRecoverer. This is the same path nitronode uses for state signing.
func TestIntegration_MsgSignAndRecover(t *testing.T) {
	keyName := skipIfNoKMS(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kmsSigner, err := NewSigner(ctx, keyName)
	require.NoError(t, err)
	defer kmsSigner.Close()

	// Wrap in EthereumMsgSigner — same as nitronode does for StateSigner
	msgSigner, err := sign.NewEthereumMsgSignerFromRaw(kmsSigner)
	require.NoError(t, err)

	kmsAddress := msgSigner.PublicKey().Address().String()
	t.Logf("KMS msg signer address: %s", kmsAddress)

	recoverer := &sign.EthereumMsgAddressRecoverer{}

	messages := []struct {
		name string
		data []byte
	}{
		{"simple text", []byte("hello world")},
		{"state hash", ethcrypto.Keccak256([]byte("channel state"))},
		{"json", []byte(`{"nonce":1,"data":"test"}`)},
	}

	for _, msg := range messages {
		t.Run(msg.name, func(t *testing.T) {
			sig, err := msgSigner.Sign(msg.data)
			require.NoError(t, err)
			require.Len(t, sig, 65)

			recovered, err := recoverer.RecoverAddress(msg.data, sig)
			require.NoError(t, err)

			assert.True(t,
				strings.EqualFold(kmsAddress, recovered.String()),
				"recovered address %s should match KMS address %s", recovered, kmsAddress,
			)
		})
	}
}

// TestIntegration_KMSvsLocalSigner is the core comparison test.
// It signs the same data with both a local private key signer and the KMS signer,
// then verifies:
//  1. Both produce valid 65-byte Ethereum signatures
//  2. Both signatures recover to their respective signer's address
//  3. The recovery mechanism works identically for both
//  4. Cross-recovery fails (KMS sig doesn't recover to local address and vice versa)
//
// Note: R and S values will differ because KMS uses a hardware HSM, but the
// signature FORMAT and RECOVERY behavior must be identical.
func TestIntegration_KMSvsLocalSigner(t *testing.T) {
	keyName := skipIfNoKMS(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --- KMS signer (raw) ---
	kmsSigner, err := NewSigner(ctx, keyName)
	require.NoError(t, err)
	defer kmsSigner.Close()

	// --- Local private key signer (raw) ---
	// Using a known test key — different from the KMS key
	localRawSigner, err := sign.NewEthereumRawSigner("4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)

	kmsAddr := kmsSigner.PublicKey().Address().String()
	localAddr := localRawSigner.PublicKey().Address().String()
	t.Logf("KMS address:   %s", kmsAddr)
	t.Logf("Local address: %s", localAddr)

	// Addresses should differ (different keys)
	assert.False(t,
		strings.EqualFold(kmsAddr, localAddr),
		"KMS and local addresses should differ (different keys)",
	)

	t.Run("raw signer comparison", func(t *testing.T) {
		messages := []string{
			"test message",
			"channel state update",
			`{"method":"create_app_session","params":{}}`,
		}

		for _, msg := range messages {
			t.Run(msg, func(t *testing.T) {
				hash := ethcrypto.Keccak256([]byte(msg))

				// Sign with both
				kmsSig, err := kmsSigner.Sign(hash)
				require.NoError(t, err)

				localSig, err := localRawSigner.Sign(hash)
				require.NoError(t, err)

				// Both are valid 65-byte signatures
				require.Len(t, kmsSig, 65)
				require.Len(t, localSig, 65)

				// Both have valid V values
				assert.True(t, kmsSig[64] == 27 || kmsSig[64] == 28)
				assert.True(t, localSig[64] == 27 || localSig[64] == 28)

				// KMS sig recovers to KMS address
				kmsRecovered, err := sign.RecoverAddressFromHash(hash, kmsSig)
				require.NoError(t, err)
				assert.True(t, strings.EqualFold(kmsAddr, kmsRecovered.String()),
					"KMS signature should recover to KMS address")

				// Local sig recovers to local address
				localRecovered, err := sign.RecoverAddressFromHash(hash, localSig)
				require.NoError(t, err)
				assert.True(t, strings.EqualFold(localAddr, localRecovered.String()),
					"local signature should recover to local address")

				// Cross-check: KMS sig should NOT recover to local address
				assert.False(t, strings.EqualFold(localAddr, kmsRecovered.String()),
					"KMS signature should not recover to local address")

				// Cross-check: local sig should NOT recover to KMS address
				assert.False(t, strings.EqualFold(kmsAddr, localRecovered.String()),
					"local signature should not recover to KMS address")
			})
		}
	})

	t.Run("msg signer comparison (EIP-191)", func(t *testing.T) {
		// Wrap both in EthereumMsgSigner — the nitronode StateSigner path
		kmsMsgSigner, err := sign.NewEthereumMsgSignerFromRaw(kmsSigner)
		require.NoError(t, err)

		localMsgSigner, err := sign.NewEthereumMsgSigner("4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
		require.NoError(t, err)

		recoverer := &sign.EthereumMsgAddressRecoverer{}

		messages := []string{
			"state update payload",
			`{"nonce":42,"allocations":[]}`,
		}

		for _, msg := range messages {
			t.Run(msg, func(t *testing.T) {
				data := []byte(msg)

				kmsSig, err := kmsMsgSigner.Sign(data)
				require.NoError(t, err)

				localSig, err := localMsgSigner.Sign(data)
				require.NoError(t, err)

				// Both valid
				require.Len(t, kmsSig, 65)
				require.Len(t, localSig, 65)

				// Recover from KMS msg signature
				kmsRecovered, err := recoverer.RecoverAddress(data, kmsSig)
				require.NoError(t, err)
				assert.True(t, strings.EqualFold(kmsAddr, kmsRecovered.String()),
					"KMS EIP-191 signature should recover to KMS address")

				// Recover from local msg signature
				localRecovered, err := recoverer.RecoverAddress(data, localSig)
				require.NoError(t, err)
				assert.True(t, strings.EqualFold(localAddr, localRecovered.String()),
					"local EIP-191 signature should recover to local address")
			})
		}
	})
}

// TestIntegration_SignDeterminism signs the same hash multiple times with KMS
// and verifies all signatures recover to the same address.
// Note: KMS signatures may differ each time (non-deterministic R due to RFC 6979
// nonce generation in HSM), but they must all be valid.
func TestIntegration_SignDeterminism(t *testing.T) {
	keyName := skipIfNoKMS(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kmsSigner, err := NewSigner(ctx, keyName)
	require.NoError(t, err)
	defer kmsSigner.Close()

	kmsAddr := kmsSigner.PublicKey().Address().String()
	hash := ethcrypto.Keccak256([]byte("determinism test"))

	for i := range 5 {
		sig, err := kmsSigner.Sign(hash)
		require.NoError(t, err, "signing attempt %d failed", i)

		recovered, err := sign.RecoverAddressFromHash(hash, sig)
		require.NoError(t, err, "recovery attempt %d failed", i)

		assert.True(t, strings.EqualFold(kmsAddr, recovered.String()),
			"attempt %d: recovered %s, expected %s", i, recovered, kmsAddr)
	}
}

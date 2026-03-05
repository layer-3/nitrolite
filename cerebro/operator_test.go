package main

import (
	"crypto/ecdsa"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperator_ParseChainID(t *testing.T) {
	t.Parallel()
	op := &Operator{}

	tests := []struct {
		name    string
		input   string
		want    uint64
		wantErr bool
	}{
		{"Valid chain ID", "1", 1, false},
		{"Valid large chain ID", "11155111", 11155111, false},
		{"Invalid number", "abc", 0, true},
		{"Negative number", "-1", 0, true},
		{"Empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := op.parseChainID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestOperator_ParseAmount(t *testing.T) {
	t.Parallel()
	op := &Operator{}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Valid integer", "100", "100", false},
		{"Valid decimal", "100.5", "100.5", false},
		{"Valid small decimal", "0.0001", "0.0001", false},
		{"Invalid number", "abc", "", true},
		{"Empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := op.parseAmount(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.String())
			}
		})
	}
}

func TestOperator_BuildStateSigner(t *testing.T) {
	t.Parallel()
	// Setup storage
	s, err := NewStorage(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	op := &Operator{
		store: s,
	}

	// Generate a wallet private key
	walletPK, err := generatePrivateKey()
	require.NoError(t, err)

	// Case 1: No session key -> Default signer
	signer, err := op.buildStateSigner(walletPK)
	require.NoError(t, err)
	assert.IsType(t, &core.ChannelDefaultSigner{}, signer)

	// Case 2: Session key configured -> Session key signer
	sessionPK, err := generatePrivateKey()
	require.NoError(t, err)

	// Create dummy metadata hash and auth sig
	metaHash := "0x1234"
	authSig := "0xabcd"

	err = s.SetSessionKey(sessionPK, metaHash, authSig)
	require.NoError(t, err)

	signer, err = op.buildStateSigner(walletPK)
	require.NoError(t, err)
	assert.IsType(t, &core.ChannelSessionKeySignerV1{}, signer)

	// Verify the signer uses the session key
	// We can check by signing something and verifying with the session key public key
	msg := []byte("hello")
	sig, err := signer.Sign(msg)
	require.NoError(t, err)

	// Recover public key from signature
	sessionPrivateKey, err1 := crypto.HexToECDSA(sessionPK[2:]) // remove 0x
	require.NoError(t, err1)
	pubKey := sessionPrivateKey.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pubKey)

	// The signature returned by ChannelSessionKeySignerV1 usually returns the signature of the session key.
	// We need to decode it first because ChannelSessionKeySignerV1 appends the type byte and wraps the signature.

	// core.ChannelSessionKeySignerV1 returns [TypeByte] + [EncodedTuple(skAuth, skSig)]
	// We need to skip the type byte and decode the rest if we want to extract the inner signature.
	// But simply checking that we got a signature is enough for unit test of 'buildStateSigner'.
	// We just want to ensure the right signer type was returned and it works.

	assert.NotEmpty(t, sig)
	assert.Equal(t, byte(core.ChannelSignerType_SessionKey), sig[0])

	// For deeper verification we would need to decode, but that requires core's unexported helpers or public helpers.
	// Let's stick to checking the signer address if possible.
	// ChannelSessionKeySignerV1 embeds sign.Signer.

	sessionSigner := signer.(*core.ChannelSessionKeySignerV1)
	assert.True(t, strings.EqualFold(addr.String(), sessionSigner.PublicKey().Address().String()), "Expected address %s, got %s", addr.String(), sessionSigner.PublicKey().Address().String())
}

func TestOperator_Connect_Failure(t *testing.T) {
	// Setup storage with a private key (required for connect)
	t.Parallel()
	s, err := NewStorage(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	pk, err := generatePrivateKey()
	require.NoError(t, err)
	err = s.SetPrivateKey(pk)
	require.NoError(t, err)

	// Attempt to connect to a non-existent server
	_, err = NewOperator("ws://localhost:12345/nonexistent", s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to clearnode")
}

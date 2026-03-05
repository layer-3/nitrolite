package apps_v1

import (
	"context"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testOwnerWallet generates a real ECDSA key pair, signs the packed app data, and returns
// the wallet address (lowercase hex), the hex-encoded signature, and the signer.
func testOwnerWallet(t *testing.T, appEntry app.AppV1) (address string, sig string) {
	t.Helper()

	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	addr := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())
	privKeyHex := hexutil.Encode(crypto.FromECDSA(key))

	signer, err := sign.NewEthereumMsgSigner(privKeyHex)
	require.NoError(t, err)

	// Update the entry with the real address for packing
	appEntry.OwnerWallet = addr
	packed, err := app.PackAppV1(appEntry)
	require.NoError(t, err)

	sigBytes, err := signer.Sign(packed)
	require.NoError(t, err)

	return addr, hexutil.Encode(sigBytes)
}

func newHandlerWithDefaults(store Store) *Handler {
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(store)
	}
	return NewHandler(store, storeTxProvider, &MockActionGateway{}, 4096)
}

func TestSubmitAppVersion_Success(t *testing.T) {
	appEntry := app.AppV1{
		ID:       "test-app",
		Metadata: "0x0000000000000000000000000000000000000000000000000000000000000000",
		Version:  1,
	}

	addr, sig := testOwnerWallet(t, appEntry)

	mockStore := &MockStore{
		createAppFn: func(entry app.AppV1) error {
			assert.Equal(t, "test-app", entry.ID)
			assert.Equal(t, addr, entry.OwnerWallet)
			return nil
		},
	}

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(mockStore, storeTxProvider, &MockActionGateway{}, 4096)

	reqPayload := rpc.AppsV1SubmitAppVersionRequest{
		App: rpc.AppV1{
			ID:          "test-app",
			OwnerWallet: addr,
			Metadata:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			Version:     "1",
		},
		OwnerSig: sig,
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1SubmitAppVersionMethod), payload),
	}

	handler.SubmitAppVersion(ctx)

	require.NotNil(t, ctx.Response)
	require.NoError(t, ctx.Response.Error())
}

func TestSubmitAppVersion_MissingOwnerWallet(t *testing.T) {
	mockStore := &MockStore{}
	handler := newHandlerWithDefaults(mockStore)

	reqPayload := rpc.AppsV1SubmitAppVersionRequest{
		App: rpc.AppV1{
			ID:       "test-app",
			Metadata: "0x00",
			Version:  "1",
		},
		OwnerSig: "0xdeadbeef",
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1SubmitAppVersionMethod), payload),
	}

	handler.SubmitAppVersion(ctx)

	require.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner_wallet")
}

func TestSubmitAppVersion_MissingOwnerSig(t *testing.T) {
	mockStore := &MockStore{}
	handler := newHandlerWithDefaults(mockStore)

	reqPayload := rpc.AppsV1SubmitAppVersionRequest{
		App: rpc.AppV1{
			ID:          "test-app",
			OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata:    "0x00",
			Version:     "1",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1SubmitAppVersionMethod), payload),
	}

	handler.SubmitAppVersion(ctx)

	require.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner_sig")
}

func TestSubmitAppVersion_InvalidAppID(t *testing.T) {
	mockStore := &MockStore{}
	handler := newHandlerWithDefaults(mockStore)

	reqPayload := rpc.AppsV1SubmitAppVersionRequest{
		App: rpc.AppV1{
			ID:          "INVALID_APP!!", // doesn't match regex
			OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata:    "0x00",
			Version:     "1",
		},
		OwnerSig: "0xdeadbeef",
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1SubmitAppVersionMethod), payload),
	}

	handler.SubmitAppVersion(ctx)

	require.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid app ID")
}

func TestSubmitAppVersion_InvalidVersion(t *testing.T) {
	mockStore := &MockStore{}
	handler := newHandlerWithDefaults(mockStore)

	reqPayload := rpc.AppsV1SubmitAppVersionRequest{
		App: rpc.AppV1{
			ID:          "test-app",
			OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata:    "0x00",
			Version:     "2", // Only version 1 is supported
		},
		OwnerSig: "0xdeadbeef",
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1SubmitAppVersionMethod), payload),
	}

	handler.SubmitAppVersion(ctx)

	require.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version 1")
}

func TestSubmitAppVersion_InvalidSignature(t *testing.T) {
	mockStore := &MockStore{}
	handler := newHandlerWithDefaults(mockStore)

	reqPayload := rpc.AppsV1SubmitAppVersionRequest{
		App: rpc.AppV1{
			ID:          "test-app",
			OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			Version:     "1",
		},
		OwnerSig: "0x" + strings.Repeat("ab", 65), // valid hex, wrong signature
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1SubmitAppVersionMethod), payload),
	}

	handler.SubmitAppVersion(ctx)

	require.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid owner signature")
}

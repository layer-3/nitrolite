package app

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackCreateAppSessionRequestV1(t *testing.T) {
	t.Parallel()
	def := AppDefinitionV1{
		ApplicationID: "chess-v1",
		Participants: []AppParticipantV1{
			{WalletAddress: "0x1111111111111111111111111111111111111111", SignatureWeight: 1},
			{WalletAddress: "0x2222222222222222222222222222222222222222", SignatureWeight: 1},
		},
		Quorum: 2,
		Nonce:  1,
	}
	sessionData := "{\"foo\":\"bar\"}"

	hash, err := PackCreateAppSessionRequestV1(def, sessionData)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 32)
}

func TestPackAppStateUpdateV1(t *testing.T) {
	t.Parallel()
	update := AppStateUpdateV1{
		AppSessionID: "0x3333333333333333333333333333333333333333333333333333333333333333",
		Intent:       AppStateUpdateIntentDeposit,
		Version:      5,
		Allocations: []AppAllocationV1{
			{Participant: "0x1111111111111111111111111111111111111111", Asset: "USDC", Amount: decimal.NewFromInt(100)},
		},
		SessionData: "{}",
	}

	hash, err := PackAppStateUpdateV1(update)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 32)
}

func TestGenerateAppSessionIDV1(t *testing.T) {
	t.Parallel()
	def := AppDefinitionV1{
		ApplicationID: "chess-v1",
		Participants: []AppParticipantV1{
			{WalletAddress: "0x1111111111111111111111111111111111111111", SignatureWeight: 1},
		},
		Quorum: 1,
		Nonce:  1,
	}

	id1, err := GenerateAppSessionIDV1(def)
	require.NoError(t, err)
	assert.NotEmpty(t, id1)

	// Same definition should produce same ID
	id2, err := GenerateAppSessionIDV1(def)
	require.NoError(t, err)
	assert.Equal(t, id1, id2)

	// Different nonce should produce different ID
	def.Nonce = 2
	id3, err := GenerateAppSessionIDV1(def)
	require.NoError(t, err)
	assert.NotEqual(t, id1, id3)
}

func TestGenerateRebalanceBatchIDV1(t *testing.T) {
	t.Parallel()
	versions := []AppSessionVersionV1{
		{SessionID: "0x1111111111111111111111111111111111111111111111111111111111111111", Version: 1},
		{SessionID: "0x2222222222222222222222222222222222222222222222222222222222222222", Version: 2},
	}

	id1, err := GenerateRebalanceBatchIDV1(versions)
	require.NoError(t, err)
	assert.NotEmpty(t, id1)

	// Deterministic
	id2, err := GenerateRebalanceBatchIDV1(versions)
	require.NoError(t, err)
	assert.Equal(t, id1, id2)

	// Different version should produce a different ID
	versions[0].Version = 99
	id3, err := GenerateRebalanceBatchIDV1(versions)
	require.NoError(t, err)
	assert.NotEqual(t, id1, id3)
}

func TestGenerateRebalanceTransactionIDV1(t *testing.T) {
	t.Parallel()
	batchID := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID := "0x2222222222222222222222222222222222222222222222222222222222222222"
	asset := "USDC"

	id1, err := GenerateRebalanceTransactionIDV1(batchID, sessionID, asset)
	require.NoError(t, err)
	assert.NotEmpty(t, id1)

	id2, err := GenerateRebalanceTransactionIDV1(batchID, sessionID, asset)
	require.NoError(t, err)
	assert.Equal(t, id1, id2)
}

func TestEnums(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "operate", AppStateUpdateIntentOperate.String())
	assert.Equal(t, "deposit", AppStateUpdateIntentDeposit.String())
	assert.Equal(t, "withdraw", AppStateUpdateIntentWithdraw.String())
	assert.Equal(t, "close", AppStateUpdateIntentClose.String())
	assert.Equal(t, "rebalance", AppStateUpdateIntentRebalance.String())
	assert.Equal(t, "unknown", AppStateUpdateIntent(255).String())

	assert.Equal(t, "", AppSessionStatusVoid.String())
	assert.Equal(t, "open", AppSessionStatusOpen.String())
	assert.Equal(t, "closed", AppSessionStatusClosed.String())
	assert.Equal(t, "unknown", AppSessionStatus(255).String())
}

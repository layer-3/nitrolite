package evm

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSigner implements sign.Signer
type MockSigner struct {
	mock.Mock
}

func (m *MockSigner) Sign(data []byte) (sign.Signature, error) {
	args := m.Called(data)
	if v := args.Get(0); v != nil {
		return v.(sign.Signature), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockSigner) PublicKey() sign.PublicKey {
	args := m.Called()
	return args.Get(0).(sign.PublicKey)
}

// MockPublicKey implements sign.PublicKey
type MockPublicKey struct {
	addr     common.Address
	pubBytes []byte
}

func (m *MockPublicKey) Address() sign.Address {
	return sign.NewEthereumAddress(m.addr)
}

func (m *MockPublicKey) Bytes() []byte {
	return m.pubBytes
}

func TestNewBlockchainClient(t *testing.T) {
	t.Parallel()
	mockEVMClient := new(MockEVMClient)
	mockAssetStore := new(MockAssetStore)
	mockSigner := new(MockSigner)

	setupMockSigner(t, mockSigner)

	contractAddress := common.HexToAddress("0x123")
	nodeAddress := "0x456"
	blockchainID := uint64(1337)

	// Mock models simulate an EVM client where NewClient/NewChannelHub return a local struct without external calls
	client, err := NewBlockchainClient(
		contractAddress,
		mockEVMClient,
		mockSigner,
		blockchainID,
		nodeAddress,
		mockAssetStore,
	)

	require.NoError(t, err)
	assert.NotNil(t, client)

	mock.AssertExpectationsForObjects(t, mockEVMClient, mockAssetStore, mockSigner)
}

func TestBlockchainClient_GetNodeBalance(t *testing.T) {
	t.Parallel()
	mockEVMClient := new(MockEVMClient)
	mockAssetStore := new(MockAssetStore)
	mockSigner := new(MockSigner)

	setupMockSigner(t, mockSigner)

	client, err := NewBlockchainClient(common.Address{}, mockEVMClient, mockSigner, 1, "0xNode", mockAssetStore)
	require.NoError(t, err)

	token := "0xToken"
	mockAssetStore.On("GetTokenDecimals", uint64(1), token).Return(uint8(18), nil)

	// Mock GetAccountBalance call
	ret := common.LeftPadBytes(big.NewInt(1000000000000000000).Bytes(), 32) // 1 ETH
	mockEVMClient.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return(ret, nil)

	balance, err := client.GetNodeBalance(token)
	require.NoError(t, err)
	assert.Equal(t, "1", balance.String())

	mock.AssertExpectationsForObjects(t, mockEVMClient, mockAssetStore, mockSigner)
}

func TestBlockchainClient_GetOpenChannels(t *testing.T) {
	t.Parallel()
	mockEVMClient := new(MockEVMClient)
	mockAssetStore := new(MockAssetStore)
	mockSigner := new(MockSigner)

	setupMockSigner(t, mockSigner)

	client, err := NewBlockchainClient(common.Address{}, mockEVMClient, mockSigner, 1, "0xNode", mockAssetStore)
	require.NoError(t, err)

	// Mock GetOpenChannels return: bytes32[]
	// Let's return 1 channel ID
	chanID := common.HexToHash("0x1234")
	// ABI encoding for dynamic array: offset, length, data
	// offset to data (32 bytes)
	offset := common.LeftPadBytes(big.NewInt(32).Bytes(), 32)
	// length (1)
	length := common.LeftPadBytes(big.NewInt(1).Bytes(), 32)
	// data (chanID)
	data := chanID.Bytes()

	ret := append(offset, length...)
	ret = append(ret, data...)

	mockEVMClient.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return(ret, nil)

	channels, err := client.GetOpenChannels("0xUser")
	require.NoError(t, err)
	assert.Len(t, channels, 1)
	assert.Equal(t, strings.ToLower(hexutil.Encode(chanID[:])), strings.ToLower(channels[0]))

	mock.AssertExpectationsForObjects(t, mockEVMClient, mockAssetStore, mockSigner)
}

func setupMockSigner(t *testing.T, mockSigner *MockSigner) {
	t.Helper()
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	pubBytes := crypto.FromECDSAPub(&privKey.PublicKey)
	mockSigner.On("PublicKey").Return(&MockPublicKey{addr: addr, pubBytes: pubBytes})
}

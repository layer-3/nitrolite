package channel_v1

import (
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"github.com/layer-3/nitrolite/clearnode/action_gateway"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mock.Mock
}

func (m *MockStore) BeginTx() (Store, func() error, func() error) {
	args := m.Called()
	return args.Get(0).(Store), args.Get(1).(func() error), args.Get(2).(func() error)
}

func (m *MockStore) LockUserState(wallet, asset string) (decimal.Decimal, error) {
	args := m.Called(wallet, asset)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockStore) GetLastUserState(wallet, asset string, signed bool) (*core.State, error) {
	args := m.Called(wallet, asset, signed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	state := args.Get(0).(core.State)
	return &state, args.Error(1)
}

func (m *MockStore) CheckOpenChannel(wallet, asset string) (string, bool, error) {
	args := m.Called(wallet, asset)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockStore) StoreUserState(state core.State) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockStore) EnsureNoOngoingStateTransitions(wallet, asset string) error {
	args := m.Called(wallet, asset)
	return args.Error(0)
}

func (m *MockStore) ScheduleInitiateEscrowWithdrawal(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

func (m *MockStore) RecordTransaction(tx core.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockStore) CreateChannel(channel core.Channel) error {
	args := m.Called(channel)
	return args.Error(0)
}

func (m *MockStore) GetChannelByID(channelID string) (*core.Channel, error) {
	args := m.Called(channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Channel), args.Error(1)
}

func (m *MockStore) GetActiveHomeChannel(wallet, asset string) (*core.Channel, error) {
	args := m.Called(wallet, asset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Channel), args.Error(1)
}

func (m *MockStore) GetUserChannels(wallet string, status *core.ChannelStatus, asset *string, channelType *core.ChannelType, limit, offset uint32) ([]core.Channel, uint32, error) {
	args := m.Called(wallet, status, asset, channelType, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]core.Channel), args.Get(1).(uint32), args.Error(2)
}

func (m *MockStore) StoreChannelSessionKeyState(state core.ChannelSessionKeyStateV1) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockStore) GetLastChannelSessionKeyVersion(wallet, sessionKey string) (uint64, error) {
	args := m.Called(wallet, sessionKey)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockStore) GetLastChannelSessionKeyStates(wallet string, sessionKey *string) ([]core.ChannelSessionKeyStateV1, error) {
	args := m.Called(wallet, sessionKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]core.ChannelSessionKeyStateV1), args.Error(1)
}

func (m *MockStore) ValidateChannelSessionKeyForAsset(wallet, sessionKey, asset, metadataHash string) (bool, error) {
	args := m.Called(wallet, sessionKey, asset, metadataHash)
	return args.Bool(0), args.Error(1)
}

func (m *MockStore) GetAppCount(ownerWallet string) (uint64, error) {
	args := m.Called(ownerWallet)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockStore) GetTotalUserStaked(wallet string) (decimal.Decimal, error) {
	args := m.Called(wallet)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockStore) RecordAction(wallet string, gatedAction core.GatedAction) error {
	args := m.Called(wallet, gatedAction)
	return args.Error(0)
}

func (m *MockStore) GetUserActionCount(wallet string, gatedAction core.GatedAction, window time.Duration) (uint64, error) {
	args := m.Called(wallet, gatedAction, window)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockStore) GetUserActionCounts(userWallet string, window time.Duration) (map[core.GatedAction]uint64, error) {
	args := m.Called(userWallet, window)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[core.GatedAction]uint64), args.Error(1)
}

func NewMockSigner() sign.Signer {
	key, _ := crypto.GenerateKey()

	signer, _ := sign.NewEthereumMsgSigner(hexutil.Encode(crypto.FromECDSA(key)))
	return signer
}

// MockSigValidator is a mock implementation of the SigValidator interface
type MockSigValidator struct {
	mock.Mock
}

func (m *MockSigValidator) Verify(wallet string, data, sig []byte) error {
	args := m.Called(wallet, data, sig)
	return args.Error(0)
}

// MockMemoryStore is a mock implementation of the MemoryStore interface
type MockMemoryStore struct {
	mock.Mock
}

func (m *MockMemoryStore) IsAssetSupported(asset, tokenAddress string, blockchainID uint64) (bool, error) {
	args := m.Called(asset, tokenAddress, blockchainID)
	return args.Bool(0), args.Error(1)
}

func (m *MockMemoryStore) GetAssetDecimals(asset string) (uint8, error) {
	args := m.Called(asset)
	return args.Get(0).(uint8), args.Error(1)
}

func (m *MockMemoryStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	args := m.Called(blockchainID, tokenAddress)
	return args.Get(0).(uint8), args.Error(1)
}

// MockAssetStore is a mock implementation of the core.AssetStore interface
type MockAssetStore struct {
	mock.Mock
}

func (m *MockAssetStore) GetAssetDecimals(asset string) (uint8, error) {
	args := m.Called(asset)
	return args.Get(0).(uint8), args.Error(1)
}

func (m *MockAssetStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	args := m.Called(blockchainID, tokenAddress)
	return args.Get(0).(uint8), args.Error(1)
}

type MockStatePacker struct {
	mock.Mock
}

func (m *MockStatePacker) PackState(state core.State) ([]byte, error) {
	args := m.Called(state)
	return args.Get(0).([]byte), args.Error(1)
}

type MockActionGateway struct {
	Err error
}

func (m *MockActionGateway) AllowAction(_ action_gateway.Store, _ string, _ core.GatedAction) error {
	return m.Err
}

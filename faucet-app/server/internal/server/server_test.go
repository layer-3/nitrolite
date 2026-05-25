package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/faucet-app/server/internal/config"
	"github.com/layer-3/nitrolite/faucet-app/server/internal/nitronode"
	"github.com/layer-3/nitrolite/pkg/log"
)

// mockNitronodeClient is a simple in-memory mock implementing NitronodeClient.
type mockNitronodeClient struct {
	ownerAddress   string
	connErr        error
	operationalErr error
	transferResult *nitronode.TransferResult
	transferErr    error
	capturedDest   string
	capturedAsset  string
	capturedAmount decimal.Decimal
}

func (m *mockNitronodeClient) GetOwnerAddress() string  { return m.ownerAddress }
func (m *mockNitronodeClient) EnsureConnected() error   { return m.connErr }
func (m *mockNitronodeClient) EnsureOperational() error { return m.operationalErr }
func (m *mockNitronodeClient) Transfer(dest, asset string, amount decimal.Decimal) (*nitronode.TransferResult, error) {
	m.capturedDest = dest
	m.capturedAsset = asset
	m.capturedAmount = amount
	return m.transferResult, m.transferErr
}

func defaultConfig() *config.Config {
	return &config.Config{
		ServerPort:               "0",
		OwnerPrivateKey:          "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		NitronodeURL:             "ws://localhost:0",
		TokenSymbol:              "usdc",
		StandardTipAmount:        "10",
		StandardTipAmountDecimal: decimal.RequireFromString("10"),
		CooldownPeriod:           "24h",
		CooldownPeriodDuration:   24 * time.Hour,
		Log:                      log.Config{Level: log.LevelDebug},
	}
}

func defaultMock() *mockNitronodeClient {
	return &mockNitronodeClient{
		ownerAddress: "0x9fc51BEE23Fb53569c46CcF013400f0E19524bd2",
		transferResult: &nitronode.TransferResult{
			TxID:   "tx-abc123",
			Amount: "10",
			Asset:  "usdc",
		},
	}
}

func TestRequestTokens_Success(t *testing.T) {
	mock := defaultMock()
	srv := NewServer(log.NewNoopLogger(), defaultConfig(), mock)

	testAddress := common.HexToAddress("0x742D35CC6634c0532925a3B8c17D18fBe3b78890").Hex()
	body, err := json.Marshal(FaucetRequest{UserAddress: testAddress})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp FaucetResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, MsgTokensSentSuccessfully, resp.Message)
	assert.Equal(t, "tx-abc123", resp.TxID)
	assert.Equal(t, "10", resp.Amount)
	assert.Equal(t, "usdc", resp.Asset)
	assert.Equal(t, testAddress, resp.Destination)

	assert.Equal(t, testAddress, mock.capturedDest)
	assert.Equal(t, "usdc", mock.capturedAsset)
	assert.True(t, decimal.RequireFromString("10").Equal(mock.capturedAmount))
}

func TestRequestTokens_InvalidAddress(t *testing.T) {
	srv := NewServer(log.NewNoopLogger(), defaultConfig(), defaultMock())

	body, err := json.Marshal(FaucetRequest{UserAddress: "not-an-address"})
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, ErrInvalidAddressFormat, resp.Error)
}

func TestRequestTokens_MissingField(t *testing.T) {
	srv := NewServer(log.NewNoopLogger(), defaultConfig(), defaultMock())

	body, err := json.Marshal(map[string]string{"wrongField": "0x1234"})
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, ErrInvalidRequestFormat, resp.Error)
}

func TestRequestTokens_ConnectionFailure(t *testing.T) {
	mock := defaultMock()
	mock.connErr = assert.AnError
	srv := NewServer(log.NewNoopLogger(), defaultConfig(), mock)

	body, err := json.Marshal(FaucetRequest{UserAddress: "0x742d35Cc6634C0532925a3b8c17d18fBE3b78890"})
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, ErrNitronodeConnectionFailed, resp.Error)
}

func TestRequestTokens_OperationalFailure(t *testing.T) {
	mock := defaultMock()
	mock.operationalErr = assert.AnError
	srv := NewServer(log.NewNoopLogger(), defaultConfig(), mock)

	body, err := json.Marshal(FaucetRequest{UserAddress: "0x742d35Cc6634C0532925a3b8c17d18fBE3b78890"})
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, ErrServiceUnavailable, resp.Error)
}

func TestRequestTokens_TransferFailure(t *testing.T) {
	mock := defaultMock()
	mock.transferResult = nil
	mock.transferErr = assert.AnError
	srv := NewServer(log.NewNoopLogger(), defaultConfig(), mock)

	body, err := json.Marshal(FaucetRequest{UserAddress: "0x742d35Cc6634C0532925a3b8c17d18fBE3b78890"})
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, ErrTransferFailed, resp.Error)
}

func TestRateLimiting(t *testing.T) {
	cfg := defaultConfig()
	cfg.CooldownPeriodDuration = 24 * time.Hour

	t.Run("second request from same wallet is rejected", func(t *testing.T) {
		mock := defaultMock()
		srv := NewServer(log.NewNoopLogger(), cfg, mock)

		testAddress := common.HexToAddress("0x742D35CC6634c0532925a3B8c17D18fBe3b78890").Hex()
		body, err := json.Marshal(FaucetRequest{UserAddress: testAddress})
		require.NoError(t, err)

		// First request — should succeed
		req1 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		srv.router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request same wallet — should be rate limited
		req2 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		srv.router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)

		var resp ErrorResponse
		require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
		assert.Equal(t, ErrRateLimitExceeded, resp.Error)
	})

	t.Run("failed transfer consumes rate limit slot", func(t *testing.T) {
		mock := defaultMock()
		mock.transferResult = nil
		mock.transferErr = assert.AnError
		srv := NewServer(log.NewNoopLogger(), cfg, mock)

		testAddress := common.HexToAddress("0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF").Hex()
		body, err := json.Marshal(FaucetRequest{UserAddress: testAddress})
		require.NoError(t, err)

		// First request fails at transfer but still consumes the rate-limit slot.
		req1 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		srv.router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusInternalServerError, w1.Code)

		// Second request is rate-limited because the slot was consumed on the first attempt.
		req2 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		srv.router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	})

	t.Run("different wallets from different IPs are not rate limited by each other", func(t *testing.T) {
		mock := defaultMock()
		srv := NewServer(log.NewNoopLogger(), cfg, mock)

		addr1 := common.HexToAddress("0x742D35CC6634c0532925a3B8c17D18fBe3b78890").Hex()
		addr2 := common.HexToAddress("0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF").Hex()

		body1, _ := json.Marshal(FaucetRequest{UserAddress: addr1})
		body2, _ := json.Marshal(FaucetRequest{UserAddress: addr2})

		req1 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body1))
		req1.Header.Set("Content-Type", "application/json")
		req1.RemoteAddr = "10.0.0.1:1234"
		w1 := httptest.NewRecorder()
		srv.router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		req2 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")
		req2.RemoteAddr = "10.0.0.2:1234"
		w2 := httptest.NewRecorder()
		srv.router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("same IP with different wallet is still rate limited", func(t *testing.T) {
		mock := defaultMock()
		srv := NewServer(log.NewNoopLogger(), cfg, mock)

		addr1 := common.HexToAddress("0x742D35CC6634c0532925a3B8c17D18fBe3b78890").Hex()
		addr2 := common.HexToAddress("0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF").Hex()

		body1, _ := json.Marshal(FaucetRequest{UserAddress: addr1})
		body2, _ := json.Marshal(FaucetRequest{UserAddress: addr2})

		req1 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body1))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		srv.router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Different wallet, same IP — should be blocked by IP limit
		req2 := httptest.NewRequest("POST", "/requestTokens", bytes.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		srv.router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	})
}

func TestInfoEndpoint(t *testing.T) {
	mock := defaultMock()
	srv := NewServer(log.NewNoopLogger(), defaultConfig(), mock)

	req := httptest.NewRequest("GET", "/info", nil)
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Nitrolite Faucet Server", resp["service"])
	assert.Equal(t, "1.0.0", resp["version"])
	assert.Equal(t, mock.ownerAddress, resp["faucet_address"])
	assert.Equal(t, "10", resp["standard_tip_amount"])
	assert.Equal(t, "usdc", resp["token_symbol"])
}

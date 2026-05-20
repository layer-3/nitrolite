package server

import (
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"faucet-server/internal/clearnode"
	"faucet-server/internal/config"
	"faucet-server/internal/logger"
)

// ClearnodeClient is the interface the server uses to interact with Clearnode.
type ClearnodeClient interface {
	GetOwnerAddress() string
	EnsureConnected() error
	EnsureOperational() error
	Transfer(destination, asset string, amount decimal.Decimal) (*clearnode.TransferResult, error)
}

// Error message constants
const (
	ErrInvalidRequestFormat      = "Invalid request format. Expected JSON with 'userAddress' field."
	ErrInvalidAddressFormat      = "Invalid address format."
	ErrClearnodeConnectionFailed = "Failed to connect to Clearnode."
	ErrServiceUnavailable        = "Faucet service is currently unavailable."
	ErrTransferFailed            = "Failed to send tokens."
	ErrRateLimitExceeded         = "Rate limit exceeded. Please try again later."
	MsgTokensSentSuccessfully    = "Tokens sent successfully"
)

type Server struct {
	config          *config.Config
	clearnodeClient ClearnodeClient
	router          *gin.Engine
	rateLimiter     *rateLimiter
}

type FaucetRequest struct {
	UserAddress string `json:"userAddress" binding:"required"`
}

type FaucetResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	TxID        string `json:"txId,omitempty"`
	Amount      string `json:"amount,omitempty"`
	Asset       string `json:"asset,omitempty"`
	Destination string `json:"destination,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(cfg *config.Config, client ClearnodeClient) *Server {
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	// Disable X-Forwarded-For trust so c.ClientIP() uses RemoteAddr.
	// Configure with actual LB IP(s) if deployed behind a trusted reverse proxy.
	router.SetTrustedProxies(nil)

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(corsMiddleware())

	server := &Server{
		config:          cfg,
		clearnodeClient: client,
		router:          router,
		rateLimiter:     newRateLimiter(cfg.CooldownPeriodDuration),
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	s.router.POST("/requestTokens", s.requestTokens)
	s.router.GET("/info", s.getInfo)
}

func (s *Server) getInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":             "Nitrolite Faucet Server",
		"version":             "1.0.0",
		"faucet_address":      s.clearnodeClient.GetOwnerAddress(),
		"standard_tip_amount": s.config.StandardTipAmountDecimal.String(),
		"token_symbol":        s.config.TokenSymbol,
		"endpoints":           []string{"/requestTokens"},
	})
}

func (s *Server) requestTokens(c *gin.Context) {
	var req FaucetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid request format: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrInvalidRequestFormat,
		})
		return
	}

	// Validate the user address
	userAddress := strings.TrimSpace(req.UserAddress)
	if !common.IsHexAddress(userAddress) {
		logger.Warnf("Invalid address format: %s", userAddress)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrInvalidAddressFormat,
		})
		return
	}

	userAddress = common.HexToAddress(userAddress).Hex()

	// Atomically check-and-record both keys under one lock. This prevents a
	// blocked IP from burning the wallet's cooldown slot and eliminates TOCTOU.
	// Every accepted request (including ones that later fail) consumes a slot,
	// preventing unlimited probing via induced failures.
	clientIP := c.ClientIP()
	if allowed, blocked := s.rateLimiter.checkAndRecordBoth(userAddress, clientIP); !allowed {
		if blocked == "address" {
			logger.Warnf("Rate limit exceeded for address %s", userAddress)
		} else {
			logger.Warnf("Rate limit exceeded for IP %s (address: %s)", clientIP, userAddress)
		}
		c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: ErrRateLimitExceeded})
		return
	}

	logger.Infof("Processing faucet request for address: %s", userAddress)

	// Ensure client is connected
	if err := s.clearnodeClient.EnsureConnected(); err != nil {
		logger.Errorf("Connection failed for %s: %v", userAddress, err)
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: ErrClearnodeConnectionFailed,
		})
		return
	}

	// Ensure client is operational
	if err := s.clearnodeClient.EnsureOperational(); err != nil {
		logger.Errorf("Service not operational for %s: %v", userAddress, err)
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: ErrServiceUnavailable,
		})
		return
	}

	// Perform the transfer
	result, err := s.clearnodeClient.Transfer(
		userAddress,
		s.config.TokenSymbol,
		s.config.StandardTipAmountDecimal,
	)
	if err != nil {
		logger.Errorf("Transfer failed for %s: %v", userAddress, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrTransferFailed,
		})
		return
	}
	if result == nil {
		logger.Errorf("Transfer returned nil result for %s", userAddress)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrTransferFailed,
		})
		return
	}

	txID := result.TxID
	amount := result.Amount
	asset := result.Asset

	logger.Infof("Successfully sent %s %s to %s (txID: %s)",
		amount, asset, userAddress, txID)

	c.JSON(http.StatusOK, FaucetResponse{
		Success:     true,
		Message:     MsgTokensSentSuccessfully,
		TxID:        txID,
		Amount:      amount,
		Asset:       asset,
		Destination: userAddress,
	})
}

func (s *Server) Start() error {
	addr := ":" + s.config.ServerPort
	logger.Infof("Starting HTTP server on port %s", s.config.ServerPort)
	return s.router.Run(addr)
}

// Middleware functions

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request
		logger.Debugf("%s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
		// Log response status
		logger.Debugf("%s %s - %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

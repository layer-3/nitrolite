package server

import (
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/faucet-app/server/internal/config"
	"github.com/layer-3/nitrolite/faucet-app/server/internal/nitronode"
	"github.com/layer-3/nitrolite/pkg/log"
)

// NitronodeClient is the interface the server uses to interact with Nitronode.
type NitronodeClient interface {
	GetOwnerAddress() string
	EnsureConnected() error
	EnsureOperational() error
	Transfer(destination, asset string, amount decimal.Decimal) (*nitronode.TransferResult, error)
}

// Error message constants
const (
	ErrInvalidRequestFormat      = "Invalid request format. Expected JSON with 'userAddress' field."
	ErrInvalidAddressFormat      = "Invalid address format."
	ErrNitronodeConnectionFailed = "Failed to connect to Nitronode."
	ErrServiceUnavailable        = "Faucet service is currently unavailable."
	ErrTransferFailed            = "Failed to send tokens."
	ErrRateLimitExceeded         = "Rate limit exceeded. Please try again later."
	MsgTokensSentSuccessfully    = "Tokens sent successfully"
)

type Server struct {
	logger          log.Logger
	config          *config.Config
	nitronodeClient NitronodeClient
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

func NewServer(logger log.Logger, cfg *config.Config, client NitronodeClient) *Server {
	if cfg.Log.Level == log.LevelDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	if len(cfg.TrustedProxyList) > 0 {
		router.SetTrustedProxies(cfg.TrustedProxyList)
	} else {
		// No proxies configured: c.ClientIP() uses RemoteAddr directly.
		// Set TRUSTED_PROXIES if the faucet is behind an ingress or load balancer,
		// otherwise IP-based rate limiting will collapse to one bucket per proxy.
		router.SetTrustedProxies(nil)
	}

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))
	router.Use(corsMiddleware())

	server := &Server{
		logger:          logger,
		config:          cfg,
		nitronodeClient: client,
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
		"faucet_address":      s.nitronodeClient.GetOwnerAddress(),
		"standard_tip_amount": s.config.StandardTipAmountDecimal.String(),
		"token_symbol":        s.config.TokenSymbol,
		"endpoints":           []string{"/requestTokens"},
	})
}

func (s *Server) requestTokens(c *gin.Context) {
	var req FaucetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Warn("invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrInvalidRequestFormat,
		})
		return
	}

	// Validate the user address
	userAddress := strings.TrimSpace(req.UserAddress)
	if !common.IsHexAddress(userAddress) {
		s.logger.Warn("invalid address format", "address", userAddress)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrInvalidAddressFormat,
		})
		return
	}

	userAddress = common.HexToAddress(userAddress).Hex()

	// Atomically check-and-record. When the per-IP bucket is enabled, both
	// keys are checked under one lock so a blocked IP cannot burn the wallet's
	// cooldown slot. When disabled (ingress handles per-IP flood protection),
	// only the wallet bucket is checked. Every accepted request — including
	// ones that later fail — consumes the slot, blocking probing via induced
	// failures.
	clientIP := c.ClientIP()
	var (
		allowed bool
		blocked string
	)
	if s.config.IPRateLimitEnabled {
		allowed, blocked = s.rateLimiter.checkAndRecordBoth(userAddress, clientIP)
	} else {
		allowed = s.rateLimiter.checkAndRecord(userAddress)
		blocked = "address"
	}
	if !allowed {
		if blocked == "address" {
			s.logger.Warn("rate limit exceeded", "key", "address", "address", userAddress)
		} else {
			s.logger.Warn("rate limit exceeded", "key", "ip", "ip", clientIP, "address", userAddress)
		}
		c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: ErrRateLimitExceeded})
		return
	}

	s.logger.Info("processing faucet request", "address", userAddress)

	// Ensure client is connected
	if err := s.nitronodeClient.EnsureConnected(); err != nil {
		s.logger.Error("connection failed", "address", userAddress, "error", err)
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: ErrNitronodeConnectionFailed,
		})
		return
	}

	// Ensure client is operational
	if err := s.nitronodeClient.EnsureOperational(); err != nil {
		s.logger.Error("service not operational", "address", userAddress, "error", err)
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: ErrServiceUnavailable,
		})
		return
	}

	// Perform the transfer
	result, err := s.nitronodeClient.Transfer(
		userAddress,
		s.config.TokenSymbol,
		s.config.StandardTipAmountDecimal,
	)
	if err != nil {
		s.logger.Error("transfer failed", "address", userAddress, "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrTransferFailed,
		})
		return
	}
	if result == nil {
		s.logger.Error("transfer returned nil result", "address", userAddress)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrTransferFailed,
		})
		return
	}

	txID := result.TxID
	amount := result.Amount
	asset := result.Asset

	s.logger.Info("transfer successful",
		"address", userAddress,
		"amount", amount,
		"asset", asset,
		"txID", txID,
	)

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
	s.logger.Info("starting HTTP server", "port", s.config.ServerPort)
	return s.router.Run(addr)
}

// Middleware functions

func requestLogger(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)
		c.Next()
		logger.Debug("response",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)
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

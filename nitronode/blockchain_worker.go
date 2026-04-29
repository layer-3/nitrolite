package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/layer-3/nitrolite/nitronode/store/database"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

type BlockchainWorkerStore interface {
	GetActions(limit uint8, chainID uint64) ([]database.BlockchainAction, error)
	GetStateByID(stateID string) (*core.State, error)
	GetChannelByID(channelID string) (*core.Channel, error)
	Complete(actionID int64, txHash string) error
	Fail(actionID int64, err string) error
	FailNoRetry(actionID int64, err string) error
	RecordAttempt(actionID int64, err string) error
}

type MetricsExporter interface {
	IncBlockchainAction(asset string, blockchainID uint64, actionType string, success bool)
}

const (
	// actionBatchSize determines how many blockchain actions to process at once
	actionBatchSize = 20

	// maxActionRetries is the maximum number of times to retry a failed action
	maxActionRetries = 5

	// blockchainWorkerTickInterval is how frequently the worker checks for new actions
	blockchainWorkerTickInterval = 30 * time.Second
)

type BlockchainWorker struct {
	blockchainID uint64
	client       core.BlockchainClient
	store        BlockchainWorkerStore
	logger       log.Logger
	metrics      MetricsExporter
}

func NewBlockchainWorker(blockchainID uint64, client core.BlockchainClient, store BlockchainWorkerStore, logger log.Logger, m MetricsExporter) *BlockchainWorker {
	return &BlockchainWorker{
		blockchainID: blockchainID,
		client:       client,
		store:        store,
		logger:       logger.WithName("bw").WithKV("blockchainID", blockchainID),
		metrics:      m,
	}
}

func (w *BlockchainWorker) Start(ctx context.Context, handleClosure func(err error)) {
	w.logger.Info("starting blockchain worker")

	childCtx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	wg.Add(1)

	var closureErr error
	var closureErrMu sync.Mutex
	childHandleClosure := func(err error) {
		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		if err != nil && closureErr == nil {
			closureErr = err
		}

		cancel()
		wg.Done()
	}

	go func() {
		defer childHandleClosure(nil)
		w.run(childCtx)
	}()

	go func() {
		wg.Wait()

		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		handleClosure(closureErr)
	}()
}

func (w *BlockchainWorker) run(ctx context.Context) {
	ticker := time.NewTicker(blockchainWorkerTickInterval)
	defer ticker.Stop()

	// Process immediately on start
	w.processActions(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("blockchain worker stopped")
			return
		case <-ticker.C:
			for w.processActions(ctx) {
			}
		}
	}
}

func (w *BlockchainWorker) processActions(ctx context.Context) bool {
	actions, err := w.store.GetActions(actionBatchSize, w.blockchainID)
	if err != nil {
		w.logger.Error("failed to get pending actions", "error", err)
		return false
	}
	if len(actions) == 0 {
		return false
	}

	w.logger.Debug("processing batch of actions", "count", len(actions))
	allSuccess := true
	for _, action := range actions {
		if ctx.Err() != nil {
			w.logger.Info("context cancelled, stopping batch processing")
			return false
		}

		ok := w.processAction(ctx, action)
		if !ok {
			allSuccess = false
		}
	}

	return allSuccess
}

func (w *BlockchainWorker) processAction(_ context.Context, action database.BlockchainAction) bool {
	logger := w.logger.
		WithKV("actionID", action.ID).
		WithKV("type", action.Type).
		WithKV("state", action.StateID).
		WithKV("attempt", action.Retries)

	state, err := w.store.GetStateByID(action.StateID)
	if err != nil {
		logger.Error("failed to get state for action", "error", err)
		if failErr := w.store.FailNoRetry(action.ID, err.Error()); failErr != nil {
			logger.Error("failed to mark action as failed", "error", failErr)
		}
		return false
	}
	if state == nil {
		errMsg := fmt.Sprintf("state not found: %s", action.StateID)
		logger.Error(errMsg)
		if failErr := w.store.FailNoRetry(action.ID, errMsg); failErr != nil {
			logger.Error("failed to mark action as failed", "error", failErr)
		}
		return false
	}

	var txHash string

	switch action.Type {
	case database.ActionTypeCheckpoint:
		txHash, err = w.client.Checkpoint(*state)

	// case database.ActionTypeInitiateEscrowDeposit:
	// 	txHash, err = w.processInitiateEscrow(state, w.client.InitiateEscrowDeposit)

	// case database.ActionTypeFinalizeEscrowDeposit:
	// 	txHash, err = w.client.FinalizeEscrowDeposit(*state)

	// case database.ActionTypeInitiateEscrowWithdrawal:
	// 	txHash, err = w.processInitiateEscrow(state, w.client.InitiateEscrowWithdrawal)

	// case database.ActionTypeFinalizeEscrowWithdrawal:
	// 	txHash, err = w.client.FinalizeEscrowWithdrawal(*state)

	default:
		err = fmt.Errorf("unknown action type: %d", action.Type)
	}

	if err != nil {
		w.handleActionError(action, err, logger)
		w.metrics.IncBlockchainAction(state.Asset, w.blockchainID, action.Type.String(), false)
		return false
	}

	if completeErr := w.store.Complete(action.ID, txHash); completeErr != nil {
		logger.Error("failed to mark action as completed", "error", completeErr)
		return false
	}
	w.metrics.IncBlockchainAction(state.Asset, w.blockchainID, action.Type.String(), true)
	logger.Info("action completed successfully", "txHash", txHash)

	return true
}

// func (w *BlockchainWorker) processInitiateEscrow(state *core.State, initiate func(core.ChannelDefinition, core.State) (string, error)) (string, error) {
// 	if state.EscrowChannelID == nil {
// 		return "", fmt.Errorf("state has no escrow channel ID")
// 	}

// 	channel, err := w.store.GetChannelByID(*state.EscrowChannelID)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get escrow channel: %w", err)
// 	}
// 	if channel == nil {
// 		return "", fmt.Errorf("escrow channel not found: %s", *state.EscrowChannelID)
// 	}

// 	def := core.ChannelDefinition{
// 		Nonce:                 channel.Nonce,
// 		Challenge:             channel.ChallengeDuration,
// 		ApprovedSigValidators: channel.ApprovedSigValidators,
// 	}

// 	return initiate(def, *state)
// }

func (w *BlockchainWorker) handleActionError(action database.BlockchainAction, err error, logger log.Logger) {
	if action.Retries >= maxActionRetries {
		logger.Warn("action failed after reaching max retries", "error", err)
		finalErr := fmt.Errorf("failed after %d retries: %w", action.Retries, err)
		if saveErr := w.store.FailNoRetry(action.ID, finalErr.Error()); saveErr != nil {
			logger.Error("failed to mark action as permanently failed", "error", saveErr)
		}
	} else {
		logger.Error("processing attempt failed, will retry later", "error", err)
		if recordErr := w.store.RecordAttempt(action.ID, err.Error()); recordErr != nil {
			logger.Error("failed to record failed attempt", "error", recordErr)
		}
	}
}

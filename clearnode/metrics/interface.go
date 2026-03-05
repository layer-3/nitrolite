package metrics

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// RuntimeMetricExporter defines the interface for recording runtime metrics across various components of the system.
type RuntimeMetricExporter interface {
	// Shared
	IncUserState(asset string, homeBlockchainID uint64, transition core.TransitionType)  // +
	RecordTransaction(asset string, txType core.TransactionType, amount decimal.Decimal) // +
	IncChannelStateSigValidation(sigType core.ChannelSignerType, success bool)           // +
	IncChannelSessionKeys()                                                              // +
	IncAppSessionKeys()                                                                  // +

	// api/rpc_router
	IncRPCMessage(msgType rpc.MsgType, method string)                                 // +
	IncRPCRequest(method, path string, success bool)                                  // +
	ObserveRPCDuration(method, path string, success bool, durationSecs time.Duration) // +
	SetRPCConnections(region, origin string, count uint32)                            // +

	// api/app_session_v1
	IncAppStateUpdate(applicationID string)                                                                  // +
	IncAppSessionUpdateSigValidation(applicationID string, sigType app.AppSessionSignerTypeV1, success bool) // +

	// Blockchain Worker
	IncBlockchainAction(asset string, blockchainID uint64, actionType string, success bool) // +

	// Event Listener
	IncBlockchainEvent(blockchainID uint64, handledSuccessfully bool) // +
}

// noopRuntimeMetricExporter is a no-op implementation for use in tests.
type noopRuntimeMetricExporter struct{}

func NewNoopRuntimeMetricExporter() RuntimeMetricExporter                                         { return noopRuntimeMetricExporter{} }
func (noopRuntimeMetricExporter) IncUserState(string, uint64, core.TransitionType)                {}
func (noopRuntimeMetricExporter) RecordTransaction(string, core.TransactionType, decimal.Decimal) {}
func (noopRuntimeMetricExporter) IncChannelStateSigValidation(core.ChannelSignerType, bool)       {}
func (noopRuntimeMetricExporter) IncChannelSessionKeys()                                          {}
func (noopRuntimeMetricExporter) IncAppSessionKeys()                                              {}
func (noopRuntimeMetricExporter) IncRPCMessage(rpc.MsgType, string)                               {}
func (noopRuntimeMetricExporter) IncRPCRequest(string, string, bool)                              {}
func (noopRuntimeMetricExporter) ObserveRPCDuration(string, string, bool, time.Duration)          {}
func (noopRuntimeMetricExporter) SetRPCConnections(string, string, uint32)                        {}
func (noopRuntimeMetricExporter) IncAppStateUpdate(string)                                        {}
func (noopRuntimeMetricExporter) IncAppSessionUpdateSigValidation(string, app.AppSessionSignerTypeV1, bool) {
}
func (noopRuntimeMetricExporter) IncBlockchainAction(string, uint64, string, bool) {
}
func (noopRuntimeMetricExporter) IncBlockchainEvent(uint64, bool) {}

// StoreMetricExporter defines the interface for setting metrics that are stored and updated by a separate metric worker.
type StoreMetricExporter interface {
	SetAppSessions(applicationID string, status app.AppSessionStatus, count uint64)
	SetChannels(asset string, status core.ChannelStatus, count uint64)
}

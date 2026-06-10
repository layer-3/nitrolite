package api

import (
	"time"

	"github.com/layer-3/nitrolite/nitronode/api/app_session_v1"
	"github.com/layer-3/nitrolite/nitronode/api/channel_v1"
	"github.com/layer-3/nitrolite/nitronode/api/node_v1"
	"github.com/layer-3/nitrolite/nitronode/api/user_v1"
	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/nitronode/store/database"
	"github.com/layer-3/nitrolite/nitronode/store/memory"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
)

type RPCRouter struct {
	Node           rpc.Node
	lg             log.Logger
	runtimeMetrics metrics.RuntimeMetricExporter
}

type RPCRouterConfig struct {
	NodeVersion  string
	MinChallenge uint32
	MaxChallenge uint32
	MaxParticipants       int
	MaxSessionDataLen     int
	MaxSessionKeyIDs      int
	MaxSessionKeysPerUser int
}

func NewRPCRouter(
	cfg RPCRouterConfig,
	node rpc.Node,
	signer sign.Signer,
	dbStore database.DatabaseStore,
	memoryStore memory.MemoryStore,
	runtimeMetrics metrics.RuntimeMetricExporter,
	logger log.Logger,
) *RPCRouter {
	r := &RPCRouter{
		Node:           node,
		lg:             logger.WithName("rpc-router"),
		runtimeMetrics: runtimeMetrics,
	}

	r.Node.Use(r.ObservabilityMiddleware)

	// Transaction wrapper helpers for each store type.
	// wrapWithMetrics executes fn inside a DB transaction with a metricStore wrapper,
	// then flushes buffered metrics only after the transaction commits successfully.
	wrapWithMetrics := func(fn func(*metricStore) error) error {
		var ms *metricStore
		if err := dbStore.ExecuteInTransaction(func(s database.DatabaseStore) error {
			ms = &metricStore{DatabaseStore: s, m: runtimeMetrics}
			return fn(ms)
		}); err != nil {
			return err
		}
		ms.flush()
		return nil
	}
	useChannelV1StoreInTx := func(h channel_v1.StoreTxHandler) error {
		return wrapWithMetrics(func(ms *metricStore) error { return h(ms) })
	}
	useAppSessionV1StoreInTx := func(h app_session_v1.StoreTxHandler) error {
		return wrapWithMetrics(func(ms *metricStore) error { return h(ms) })
	}
	useUserV1StoreInTx := func(h user_v1.StoreTxHandler) error {
		return wrapWithMetrics(func(ms *metricStore) error { return h(ms) })
	}

	nodeAddress := signer.PublicKey().Address().String()

	statePacker := core.NewStatePackerV1(memoryStore)
	stateAdvancer := core.NewStateAdvancerV1(memoryStore)

	nodeChannelSigner, err := core.NewChannelDefaultSigner(signer)
	if err != nil {
		panic("failed to create channel wallet signer: " + err.Error())
	}

	channelV1Handler := channel_v1.NewHandler(useChannelV1StoreInTx, memoryStore, nodeChannelSigner, stateAdvancer, statePacker, nodeAddress, cfg.MinChallenge, cfg.MaxChallenge, runtimeMetrics, cfg.MaxSessionKeyIDs, cfg.MaxSessionKeysPerUser)
	appSessionV1Handler := app_session_v1.NewHandler(useAppSessionV1StoreInTx, memoryStore, nodeChannelSigner, stateAdvancer, statePacker, nodeAddress, runtimeMetrics,
		cfg.MaxParticipants, cfg.MaxSessionDataLen, cfg.MaxSessionKeyIDs, cfg.MaxSessionKeysPerUser)
	nodeV1Handler := node_v1.NewHandler(memoryStore, nodeAddress, cfg.NodeVersion)
	userV1Handler := user_v1.NewHandler(dbStore, useUserV1StoreInTx)

	appSessionV1Group := r.Node.NewGroup(rpc.AppSessionsV1Group.String())
	appSessionV1Group.Handle(rpc.AppSessionsV1SubmitDepositStateMethod.String(), appSessionV1Handler.SubmitDepositState)
	appSessionV1Group.Handle(rpc.AppSessionsV1SubmitAppStateMethod.String(), appSessionV1Handler.SubmitAppState)
	appSessionV1Group.Handle(rpc.AppSessionsV1CreateAppSessionMethod.String(), appSessionV1Handler.CreateAppSession)
	appSessionV1Group.Handle(rpc.AppSessionsV1GetAppDefinitionMethod.String(), appSessionV1Handler.GetAppDefinition)
	appSessionV1Group.Handle(rpc.AppSessionsV1GetAppSessionsMethod.String(), appSessionV1Handler.GetAppSessions)
	appSessionV1Group.Handle(rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), appSessionV1Handler.SubmitSessionKeyState)
	appSessionV1Group.Handle(rpc.AppSessionsV1GetLastKeyStatesMethod.String(), appSessionV1Handler.GetLastKeyStates)

	channelV1Group := r.Node.NewGroup(rpc.ChannelV1Group.String())
	channelV1Group.Handle(rpc.ChannelsV1GetChannelsMethod.String(), channelV1Handler.GetChannels)
	channelV1Group.Handle(rpc.ChannelsV1GetEscrowChannelMethod.String(), channelV1Handler.GetEscrowChannel)
	channelV1Group.Handle(rpc.ChannelsV1GetHomeChannelMethod.String(), channelV1Handler.GetHomeChannel)
	channelV1Group.Handle(rpc.ChannelsV1GetLatestStateMethod.String(), channelV1Handler.GetLatestState)
	channelV1Group.Handle(rpc.ChannelsV1RequestCreationMethod.String(), channelV1Handler.RequestCreation)
	channelV1Group.Handle(rpc.ChannelsV1SubmitStateMethod.String(), channelV1Handler.SubmitState)
	channelV1Group.Handle(rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), channelV1Handler.SubmitSessionKeyState)
	channelV1Group.Handle(rpc.ChannelsV1GetLastKeyStatesMethod.String(), channelV1Handler.GetLastKeyStates)

	nodeV1Group := r.Node.NewGroup(rpc.NodeV1Group.String())
	nodeV1Group.Handle(rpc.NodeV1PingMethod.String(), nodeV1Handler.Ping)
	nodeV1Group.Handle(rpc.NodeV1GetAssetsMethod.String(), nodeV1Handler.GetAssets)
	nodeV1Group.Handle(rpc.NodeV1GetConfigMethod.String(), nodeV1Handler.GetConfig)

	userV1Group := r.Node.NewGroup(rpc.UserV1Group.String())
	userV1Group.Handle(rpc.UserV1GetBalancesMethod.String(), userV1Handler.GetBalances)
	userV1Group.Handle(rpc.UserV1GetTransactionsMethod.String(), userV1Handler.GetTransactions)

	// Pre-publish per-method metric series at 0 so dashboards and absent()-style
	// alerts have defined values before any traffic arrives. Must run after all
	// group.Handle calls so RegisteredMethods is complete.
	runtimeMetrics.SeedRPCMethodMetrics(r.Node.RegisteredMethods(), MethodPathDomains())

	return r
}

func (r *RPCRouter) ObservabilityMiddleware(c *rpc.Context) {
	logger := r.lg.WithKV("requestID", c.Request.RequestID)
	c.Context = log.SetContextLogger(c.Context, logger)
	logger = log.FromContext(c.Context)

	startTime := time.Now()
	methodPath := getMethodPath(c)

	r.runtimeMetrics.IncRPCInflight(c.Request.Method)
	defer r.runtimeMetrics.DecRPCInflight(c.Request.Method)

	c.Next()

	reqDuration := time.Since(startTime)

	r.runtimeMetrics.IncRPCMessage(c.Request.Type, c.Request.Method)
	r.runtimeMetrics.IncRPCMessage(c.Response.Type, c.Response.Method)
	r.runtimeMetrics.IncRPCRequest(c.Request.Method, methodPath, c.Response.Type == rpc.MsgTypeResp)
	r.runtimeMetrics.ObserveRPCDuration(c.Request.Method, methodPath, c.Response.Type == rpc.MsgTypeResp, reqDuration)

	if c.Request.Method == rpc.NodeV1PingMethod.String() {
		// Skip logging for ping requests
		return
	}

	logger.Info("handled RPC request",
		"method", c.Request.Method,
		"success", c.Response.Type == rpc.MsgTypeResp,
		"duration", reqDuration.String())
}

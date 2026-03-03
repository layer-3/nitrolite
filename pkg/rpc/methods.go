// Package rpc provides common method and event type definitions.
package rpc

// Method represents an RPC method name that can be called on the server.
type Group string
type Method string

const (
	// Channels Group - V1 Methods
	ChannelV1Group                        Group  = "channels.v1"
	ChannelsV1GetHomeChannelMethod        Method = "channels.v1.get_home_channel"
	ChannelsV1GetEscrowChannelMethod      Method = "channels.v1.get_escrow_channel"
	ChannelsV1GetChannelsMethod           Method = "channels.v1.get_channels"
	ChannelsV1GetLatestStateMethod        Method = "channels.v1.get_latest_state"
	ChannelsV1GetStatesMethod             Method = "channels.v1.get_states"
	ChannelsV1RequestCreationMethod       Method = "channels.v1.request_creation"
	ChannelsV1SubmitStateMethod           Method = "channels.v1.submit_state"
	ChannelsV1SubmitSessionKeyStateMethod Method = "channels.v1.submit_session_key_state"
	ChannelsV1GetLastKeyStatesMethod      Method = "channels.v1.get_last_key_states"

	// App Sessions Group - V1 Methods
	AppSessionsV1Group                       Group  = "app_sessions.v1"
	AppSessionsV1SubmitDepositStateMethod    Method = "app_sessions.v1.submit_deposit_state"
	AppSessionsV1SubmitAppStateMethod        Method = "app_sessions.v1.submit_app_state"
	AppSessionsV1RebalanceAppSessionsMethod  Method = "app_sessions.v1.rebalance_app_sessions"
	AppSessionsV1GetAppDefinitionMethod      Method = "app_sessions.v1.get_app_definition"
	AppSessionsV1GetAppSessionsMethod        Method = "app_sessions.v1.get_app_sessions"
	AppSessionsV1CreateAppSessionMethod      Method = "app_sessions.v1.create_app_session"
	AppSessionsV1SubmitSessionKeyStateMethod Method = "app_sessions.v1.submit_session_key_state"
	AppSessionsV1GetLastKeyStatesMethod      Method = "app_sessions.v1.get_last_key_states"

	// Apps Group - V1 Methods
	AppsV1Group                  Group  = "apps.v1"
	AppsV1GetAppsMethod          Method = "apps.v1.get_apps"
	AppsV1SubmitAppVersionMethod Method = "apps.v1.submit_app_version"

	// User Group - V1 Methods
	UserV1Group                 Group  = "user.v1"
	UserV1GetBalancesMethod     Method = "user.v1.get_balances"
	UserV1GetTransactionsMethod Method = "user.v1.get_transactions"

	// Node Group - V1 Methods
	NodeV1Group           Group  = "node.v1"
	NodeV1PingMethod      Method = "node.v1.ping"
	NodeV1GetConfigMethod Method = "node.v1.get_config"
	NodeV1GetAssetsMethod Method = "node.v1.get_assets"
)

// String returns the string representation of the method.
func (m Method) String() string {
	return string(m)
}

// String returns the string representation of the group.
func (g Group) String() string {
	return string(g)
}

// Event represents a notification event type sent by the server.
// Events are unsolicited notifications sent to connected clients.
type Event string

const ()

// String returns the string representation of the event.
func (e Event) String() string {
	return string(e)
}

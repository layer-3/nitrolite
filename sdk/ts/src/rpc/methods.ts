/**
 * Method and Group constants for RPC API methods
 */

/**
 * Group represents an RPC API group
 */
export type Group = string;

/**
 * Method represents an RPC method name
 */
export type Method = string;

// Channels Group - V1 Methods
export const ChannelV1Group: Group = 'channels.v1';
export const ChannelsV1GetHomeChannelMethod: Method = 'channels.v1.get_home_channel';
export const ChannelsV1GetEscrowChannelMethod: Method = 'channels.v1.get_escrow_channel';
export const ChannelsV1GetChannelsMethod: Method = 'channels.v1.get_channels';
export const ChannelsV1GetLatestStateMethod: Method = 'channels.v1.get_latest_state';
export const ChannelsV1GetStatesMethod: Method = 'channels.v1.get_states';
export const ChannelsV1RequestCreationMethod: Method = 'channels.v1.request_creation';
export const ChannelsV1SubmitStateMethod: Method = 'channels.v1.submit_state';

// Channel Session Key Methods - V1
export const ChannelsV1SubmitSessionKeyStateMethod: Method = 'channels.v1.submit_session_key_state';
export const ChannelsV1GetLastKeyStatesMethod: Method = 'channels.v1.get_last_key_states';

// App Sessions Group - V1 Methods
export const AppSessionsV1Group: Group = 'app_sessions.v1';
export const AppSessionsV1SubmitDepositStateMethod: Method = 'app_sessions.v1.submit_deposit_state';
export const AppSessionsV1SubmitAppStateMethod: Method = 'app_sessions.v1.submit_app_state';
export const AppSessionsV1RebalanceAppSessionsMethod: Method = 'app_sessions.v1.rebalance_app_sessions';
export const AppSessionsV1GetAppDefinitionMethod: Method = 'app_sessions.v1.get_app_definition';
export const AppSessionsV1GetAppSessionsMethod: Method = 'app_sessions.v1.get_app_sessions';
export const AppSessionsV1CreateAppSessionMethod: Method = 'app_sessions.v1.create_app_session';

// App Session Key Methods - V1
export const AppSessionsV1SubmitSessionKeyStateMethod: Method = 'app_sessions.v1.submit_session_key_state';
export const AppSessionsV1GetLastKeyStatesMethod: Method = 'app_sessions.v1.get_last_key_states';

// Apps Group - V1 Methods
export const AppsV1Group: Group = 'apps.v1';
export const AppsV1GetAppsMethod: Method = 'apps.v1.get_apps';
export const AppsV1SubmitAppVersionMethod: Method = 'apps.v1.submit_app_version';

// User Group - V1 Methods
export const UserV1Group: Group = 'user.v1';
export const UserV1GetBalancesMethod: Method = 'user.v1.get_balances';
export const UserV1GetTransactionsMethod: Method = 'user.v1.get_transactions';
export const UserV1GetActionAllowancesMethod: Method = 'user.v1.get_action_allowances';

// Node Group - V1 Methods
export const NodeV1Group: Group = 'node.v1';
export const NodeV1PingMethod: Method = 'node.v1.ping';
export const NodeV1GetConfigMethod: Method = 'node.v1.get_config';
export const NodeV1GetAssetsMethod: Method = 'node.v1.get_assets';

/**
 * Event represents a notification event type sent by the server
 */
export type Event = string;

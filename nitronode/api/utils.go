package api

import (
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func getMethodPath(c *rpc.Context) string {
	switch c.Request.Method {
	case rpc.AppSessionsV1SubmitAppStateMethod.String():
		var reqPayload rpc.AppSessionsV1SubmitAppStateRequest
		if err := c.Request.Payload.Translate(&reqPayload); err != nil {
			break
		}
		return reqPayload.AppStateUpdate.Intent.String()
	case rpc.ChannelsV1RequestCreationMethod.String():
		var reqPayload rpc.ChannelsV1RequestCreationRequest
		if err := c.Request.Payload.Translate(&reqPayload); err != nil {
			break
		}
		return reqPayload.State.Transition.Type.String()
	case rpc.ChannelsV1SubmitStateMethod.String():
		var reqPayload rpc.ChannelsV1SubmitStateRequest
		if err := c.Request.Payload.Translate(&reqPayload); err != nil {
			break
		}
		return reqPayload.State.Transition.Type.String()
	}

	return "default"
}

// MethodPathDomains returns the bounded `path` label domain per RPC method
// whose request payload determines path (matching getMethodPath's switch).
// Methods not in the map emit only path="default". Used by metrics seeding so
// absent()-style alerts have defined values for every (method, path, result)
// tuple at cold start, not just (method, "default", result).
func MethodPathDomains() map[string][]string {
	intents := make([]string, 0, len(app.AllAppStateUpdateIntents))
	for _, i := range app.AllAppStateUpdateIntents {
		intents = append(intents, i.String())
	}
	transitions := make([]string, 0, len(core.AllTransitionTypes))
	for _, t := range core.AllTransitionTypes {
		transitions = append(transitions, t.String())
	}
	return map[string][]string{
		rpc.AppSessionsV1SubmitAppStateMethod.String(): intents,
		rpc.ChannelsV1RequestCreationMethod.String():   transitions,
		rpc.ChannelsV1SubmitStateMethod.String():       transitions,
	}
}

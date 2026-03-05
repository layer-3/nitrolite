package api

import "github.com/layer-3/nitrolite/pkg/rpc"

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

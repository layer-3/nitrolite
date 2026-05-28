package node_v1

import (
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// Ping handles the ping request and responds with the ping method.
func (h *Handler) Ping(c *rpc.Context) {
	c.Succeed(rpc.NodeV1PingMethod.String(), nil)
}

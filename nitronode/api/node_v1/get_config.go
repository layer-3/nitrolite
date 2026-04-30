package node_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetConfig retrieves the current configuration of the Node.
func (h *Handler) GetConfig(c *rpc.Context) {
	var req rpc.NodeV1GetConfigRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	blockchains, err := h.memoryStore.GetBlockchains()
	if err != nil {
		c.Fail(err, "failed to retrieve blockchains")
		return
	}

	response := rpc.NodeV1GetConfigResponse{
		NodeAddress:            h.nodeAddress,
		NodeVersion:            h.nodeVersion,
		SupportedSigValidators: core.ChannelSignerTypes,
		Blockchains:            []rpc.BlockchainInfoV1{},
	}

	for _, bc := range blockchains {
		response.Blockchains = append(response.Blockchains, mapBlockchainV1(bc))
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}

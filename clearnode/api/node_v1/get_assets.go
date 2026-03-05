package node_v1

import (
	"strconv"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetAssets retrieves the current assets of the Node.
func (h *Handler) GetAssets(c *rpc.Context) {
	var req rpc.NodeV1GetAssetsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	// Convert string BlockchainID to uint64 if provided
	var blockchainIDPtr *uint64
	if req.BlockchainID != nil {
		blockchainID, err := strconv.ParseUint(*req.BlockchainID, 10, 64)
		if err != nil {
			c.Fail(err, "invalid blockchain_id")
			return
		}
		blockchainIDPtr = &blockchainID
	}

	assets, err := h.memoryStore.GetAssets(blockchainIDPtr)
	if err != nil {
		c.Fail(err, "failed to retrieve assets")
		return
	}

	response := rpc.NodeV1GetAssetsResponse{
		Assets: []rpc.AssetV1{},
	}

	for _, asset := range assets {
		response.Assets = append(response.Assets, mapAssetV1(asset))
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}

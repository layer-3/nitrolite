package node_v1

// Handler manages channel state transitions and provides RPC endpoints for state submission.
type Handler struct {
	memoryStore MemoryStore
	nodeVersion string // Node software version
	nodeAddress string // Node's wallet address for channel ID calculation
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	memoryStore MemoryStore,
	nodeAddress string,
	nodeVersion string,
) *Handler {
	return &Handler{
		memoryStore: memoryStore,
		nodeAddress: nodeAddress,
		nodeVersion: nodeVersion,
	}
}

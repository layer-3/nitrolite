package apps_v1

// Handler manages app registry operations and provides RPC endpoints.
type Handler struct {
	store         Store
	useStoreInTx  StoreTxProvider
	actionGateway ActionGateway

	maxAppMetadataLen int
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(store Store, useStoreInTx StoreTxProvider, actionGateway ActionGateway, maxAppMetadataLen int) *Handler {
	return &Handler{
		store:             store,
		useStoreInTx:      useStoreInTx,
		actionGateway:     actionGateway,
		maxAppMetadataLen: maxAppMetadataLen,
	}
}

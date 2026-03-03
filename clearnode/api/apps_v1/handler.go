package apps_v1

// Handler manages app registry operations and provides RPC endpoints.
type Handler struct {
	store Store

	maxAppMetadataLen int
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(store Store, maxAppMetadataLen int) *Handler {
	return &Handler{
		store:             store,
		maxAppMetadataLen: maxAppMetadataLen,
	}
}

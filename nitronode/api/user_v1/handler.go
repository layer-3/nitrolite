package user_v1

// Handler manages user data operations and provides RPC endpoints.
type Handler struct {
	store        Store
	useStoreInTx StoreTxProvider
}

// NewHandler creates a new Handler instance with the provided dependencies.
func NewHandler(
	store Store,
	useStoreInTx StoreTxProvider,
) *Handler {
	return &Handler{
		store:        store,
		useStoreInTx: useStoreInTx,
	}
}

package core

import "errors"

// ErrTokenNotSupported indicates a token address is not configured in the node
// asset store for the given blockchain. Callers can use errors.Is to distinguish
// this deterministic "not configured" condition from genuine store failures.
var ErrTokenNotSupported = errors.New("token not supported")

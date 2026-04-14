package app

import (
	"fmt"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/core"
)

var AppIDV1Regex = regexp.MustCompile(`^[a-z0-9][-a-z0-9]{0,65}$`)

// AppV1 represents an application registry entry.
type AppV1 struct {
	ID                          string
	OwnerWallet                 string
	Metadata                    string
	Version                     uint64
	CreationApprovalNotRequired bool
}

// AppInfoV1 represents full application info including timestamps.
type AppInfoV1 struct {
	App       AppV1
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PackAppV1 packs the AppV1 for signing using ABI encoding.
func PackAppV1(app AppV1) ([]byte, error) {
	var err error
	app.OwnerWallet, err = core.NormalizeHexAddress(app.OwnerWallet)
	if err != nil {
		return nil, fmt.Errorf("invalid owner wallet address: %v", err)
	}

	args := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},               // id
		{Type: abi.Type{T: abi.AddressTy}},              // ownerWallet
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // metadata (bytes32)
		{Type: abi.Type{T: abi.UintTy, Size: 64}},       // version
		{Type: abi.Type{T: abi.BoolTy}},                 // creationApprovalNotRequired
	}

	appMetadataHash := crypto.Keccak256Hash([]byte(app.Metadata))

	packed, err := args.Pack(
		app.ID,
		common.HexToAddress(app.OwnerWallet),
		appMetadataHash,
		app.Version,
		app.CreationApprovalNotRequired,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack app: %w", err)
	}

	return crypto.Keccak256(packed), nil
}

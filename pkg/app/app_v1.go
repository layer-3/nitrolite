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

// ApplicationIDRegex bounds the advisory application identifier to lowercase
// letters, digits, dashes and underscores, 1..66 chars — matching the DB
// column width (VARCHAR(66), see
// clearnode/config/migrations/postgres/20260420000000_add_application_id_to_writes.sql).
var ApplicationIDRegex = regexp.MustCompile(`^[a-z0-9_-]{1,66}$`)

// IsValidApplicationID reports whether id is a well-formed advisory
// application identifier (see ApplicationIDRegex).
func IsValidApplicationID(id string) bool {
	return ApplicationIDRegex.MatchString(id)
}

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

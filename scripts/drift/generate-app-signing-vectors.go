//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/shopspring/decimal"
)

const (
	user = "0x1111111111111111111111111111111111111111"
	svc  = "0x2222222222222222222222222222222222222222"
)

type vector struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

func main() {
	definition := app.AppDefinitionV1{
		ApplicationID: "store-v1",
		Participants: []app.AppParticipantV1{
			{WalletAddress: user, SignatureWeight: 1},
			{WalletAddress: svc, SignatureWeight: 1},
		},
		Quorum: 2,
		Nonce:  123456789,
	}

	appSessionID, err := app.GenerateAppSessionIDV1(definition)
	must("generate app session id", err)

	vectors := []vector{
		hashCreate("create_session", definition, `{"cart":"demo"}`),
		{Name: "app_session_id", Hash: appSessionID},
		hashUpdate("deposit", app.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentDeposit,
			Version:      2,
			Allocations: []app.AppAllocationV1{
				{Participant: user, Asset: "YUSD", Amount: decimal.RequireFromString("1.25")},
				{Participant: svc, Asset: "YUSD", Amount: decimal.RequireFromString("0")},
			},
			SessionData: `{"intent":"deposit"}`,
		}),
		hashUpdate("operate_purchase", app.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      3,
			Allocations: []app.AppAllocationV1{
				{Participant: user, Asset: "YUSD", Amount: decimal.RequireFromString("0.35")},
				{Participant: svc, Asset: "YUSD", Amount: decimal.RequireFromString("0.90")},
			},
			SessionData: `{"intent":"purchase","item_id":1,"item_price":"0.90"}`,
		}),
		hashUpdate("withdraw", app.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      4,
			Allocations: []app.AppAllocationV1{
				{Participant: user, Asset: "YUSD", Amount: decimal.RequireFromString("0.10")},
				{Participant: svc, Asset: "YUSD", Amount: decimal.RequireFromString("0.90")},
			},
			SessionData: `{"intent":"withdraw"}`,
		}),
		hashUpdate("fractional_deposit", app.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentDeposit,
			Version:      5,
			Allocations: []app.AppAllocationV1{
				{Participant: user, Asset: "YUSD", Amount: decimal.RequireFromString("1.23456789")},
				{Participant: svc, Asset: "YUSD", Amount: decimal.RequireFromString("0")},
			},
			SessionData: `{"intent":"deposit","note":"fractional"}`,
		}),
		hashUpdate("max_uint64_version", app.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      math.MaxUint64,
			Allocations: []app.AppAllocationV1{
				{Participant: user, Asset: "YUSD", Amount: decimal.RequireFromString("0")},
				{Participant: svc, Asset: "YUSD", Amount: decimal.RequireFromString("1.25")},
			},
			SessionData: `{"intent":"withdraw","boundary":"max_uint64_version"}`,
		}),
		hashCreate("max_uint64_nonce_create_session", app.AppDefinitionV1{
			ApplicationID: definition.ApplicationID,
			Participants:  definition.Participants,
			Quorum:        definition.Quorum,
			Nonce:         math.MaxUint64,
		}, `{"cart":"max-nonce"}`),
		hashSessionKey("session_key_state", app.AppSessionKeyStateV1{
			UserAddress:    user,
			SessionKey:     svc,
			Version:        1,
			ApplicationIDs: []string{"0x00000000000000000000000000000000000000000000000000000000000000a1"},
			AppSessionIDs:  []string{"0x00000000000000000000000000000000000000000000000000000000000000b1"},
			ExpiresAt:      time.Unix(1739812234, 0).UTC(),
			UserSig:        "0xSig",
		}),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	must("encode vectors", encoder.Encode(vectors))
}

func hashCreate(name string, definition app.AppDefinitionV1, sessionData string) vector {
	hash, err := app.PackCreateAppSessionRequestV1(definition, sessionData)
	must(name, err)
	return vector{Name: name, Hash: hexutil.Encode(hash)}
}

func hashUpdate(name string, update app.AppStateUpdateV1) vector {
	hash, err := app.PackAppStateUpdateV1(update)
	must(name, err)
	return vector{Name: name, Hash: hexutil.Encode(hash)}
}

func hashSessionKey(name string, state app.AppSessionKeyStateV1) vector {
	hash, err := app.PackAppSessionKeyStateV1(state)
	must(name, err)
	return vector{Name: name, Hash: hexutil.Encode(hash)}
}

func must(action string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", action, err)
		os.Exit(1)
	}
}

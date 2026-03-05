package stress

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/sign"
)

type pipe struct {
	wallets []wallet
	signers []*app.AppSessionSignerV1
}

type appSessionSigs struct {
	definition app.AppDefinitionV1
	createSigs []string
	// deposit is always index 0, close is always last
	steps []stepSigs // deposit + N operates + close
}

type stepSigs struct {
	update app.AppStateUpdateV1
	sigs   []string
}

func deriveAppSessionKey(masterKey string, pipeIdx, walletIdx int) string {
	data := masterKey + ":appsession:" + strconv.Itoa(pipeIdx) + ":" + strconv.Itoa(walletIdx)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// buildAllocations creates allocations where the "heavy" participant (rotating)
// gets the base share plus any rounding remainder, simulating fund movement.
func buildAllocations(wallets []wallet, numParticipants, operateIdx int, asset string, amount decimal.Decimal) []app.AppAllocationV1 {
	share := amount.Div(decimal.NewFromInt(int64(numParticipants))).Truncate(6)
	remainder := amount.Sub(share.Mul(decimal.NewFromInt(int64(numParticipants))))
	heavy := operateIdx % numParticipants

	allocs := make([]app.AppAllocationV1, numParticipants)
	for i := range numParticipants {
		amt := share
		if i == heavy {
			amt = amt.Add(remainder)
		}
		allocs[i] = app.AppAllocationV1{
			Participant: wallets[i].addr,
			Asset:       asset,
			Amount:      amt,
		}
	}
	return allocs
}

func preGenerateSigs(p *pipe, numParticipants, numOperates int, nonce uint64, asset string, amount decimal.Decimal) (appSessionSigs, error) {
	participants := make([]app.AppParticipantV1, numParticipants)
	for i := range numParticipants {
		participants[i] = app.AppParticipantV1{
			WalletAddress:   p.wallets[i].addr,
			SignatureWeight: 1,
		}
	}

	definition := app.AppDefinitionV1{
		ApplicationID: "stress-test",
		Participants:  participants,
		Quorum:        uint8(numParticipants),
		Nonce:         nonce,
	}

	sessionID, err := app.GenerateAppSessionIDV1(definition)
	if err != nil {
		return appSessionSigs{}, fmt.Errorf("generate session ID: %w", err)
	}

	// Create signatures
	createReq, err := app.PackCreateAppSessionRequestV1(definition, "{}")
	if err != nil {
		return appSessionSigs{}, fmt.Errorf("pack create: %w", err)
	}
	createSigs, err := signAll(p.signers, numParticipants, createReq)
	if err != nil {
		return appSessionSigs{}, fmt.Errorf("sign create: %w", err)
	}

	// Steps: deposit + operates + close
	totalSteps := 1 + numOperates + 1
	steps := make([]stepSigs, totalSteps)
	version := uint64(2)

	// Deposit (step 0)
	depositUpdate := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      version,
		Allocations: []app.AppAllocationV1{
			{Participant: p.wallets[0].addr, Asset: asset, Amount: amount},
		},
	}
	depositReq, err := app.PackAppStateUpdateV1(depositUpdate)
	if err != nil {
		return appSessionSigs{}, fmt.Errorf("pack deposit: %w", err)
	}
	depositSigs, err := signAll(p.signers, numParticipants, depositReq)
	if err != nil {
		return appSessionSigs{}, fmt.Errorf("sign deposit: %w", err)
	}
	steps[0] = stepSigs{update: depositUpdate, sigs: depositSigs}
	version++

	// Operates (steps 1..numOperates)
	var lastAllocations []app.AppAllocationV1
	for i := range numOperates {
		allocs := buildAllocations(p.wallets, numParticipants, i, asset, amount)
		lastAllocations = allocs

		update := app.AppStateUpdateV1{
			AppSessionID: sessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      version,
			Allocations:  allocs,
		}
		req, err := app.PackAppStateUpdateV1(update)
		if err != nil {
			return appSessionSigs{}, fmt.Errorf("pack operate %d: %w", i, err)
		}
		sigs, err := signAll(p.signers, numParticipants, req)
		if err != nil {
			return appSessionSigs{}, fmt.Errorf("sign operate %d: %w", i, err)
		}
		steps[1+i] = stepSigs{update: update, sigs: sigs}
		version++
	}

	// Close (last step) — same allocations as last operate
	closeUpdate := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      version,
		Allocations:  lastAllocations,
	}
	closeReq, err := app.PackAppStateUpdateV1(closeUpdate)
	if err != nil {
		return appSessionSigs{}, fmt.Errorf("pack close: %w", err)
	}
	closeSigs, err := signAll(p.signers, numParticipants, closeReq)
	if err != nil {
		return appSessionSigs{}, fmt.Errorf("sign close: %w", err)
	}
	steps[totalSteps-1] = stepSigs{update: closeUpdate, sigs: closeSigs}

	return appSessionSigs{
		definition: definition,
		createSigs: createSigs,
		steps:      steps,
	}, nil
}

func signAll(signers []*app.AppSessionSignerV1, n int, data []byte) ([]string, error) {
	sigs := make([]string, n)
	for i := range n {
		sig, err := signers[i].Sign(data)
		if err != nil {
			return nil, fmt.Errorf("signer %d: %w", i, err)
		}
		sigs[i] = sig.String()
	}
	return sigs, nil
}

// RunAppSessionLifecycleStress runs a parallel app session lifecycle stress test.
//
// Spec format: app-session-lifecycle:app_sessions:participants:operates:asset[:amount]
//   - app_sessions: number of concurrent app session lifecycles
//   - participants: wallets per pipe (weight 1 each, quorum = all must sign)
//   - operates: number of operate state updates per session
//
// Each pipe = create + deposit + N operates + close = N + 3 operations.
func RunAppSessionLifecycleStress(ctx context.Context, cfg *Config, spec TestSpec) (Report, error) {
	if os.Getenv("STRESS_PRIVATE_KEY") == "" {
		return Report{}, fmt.Errorf("STRESS_PRIVATE_KEY is required for app-session-lifecycle (sender must have funds)")
	}

	numAppSessions := spec.TotalReqs
	if numAppSessions < 1 {
		numAppSessions = 1
	}
	numParticipants := spec.Connections
	if numParticipants < 1 {
		numParticipants = 3
	}
	if numParticipants > 255 {
		return Report{}, fmt.Errorf("participants must be 1-255, got %d", numParticipants)
	}

	if len(spec.ExtraArgs) < 2 {
		return Report{}, fmt.Errorf("usage: app-session-lifecycle:app_sessions:participants:operates:asset[:amount]")
	}

	numOperates, err := strconv.Atoi(spec.ExtraArgs[0])
	if err != nil || numOperates < 1 {
		return Report{}, fmt.Errorf("invalid operates %q: must be positive integer", spec.ExtraArgs[0])
	}

	asset := spec.ExtraArgs[1]

	amount := decimal.NewFromFloat(0.000003)
	if len(spec.ExtraArgs) > 2 {
		parsed, err := decimal.NewFromString(spec.ExtraArgs[2])
		if err != nil {
			return Report{}, fmt.Errorf("invalid amount %q: %w", spec.ExtraArgs[2], err)
		}
		amount = parsed
	}

	senderAddr, err := cfg.WalletAddress()
	if err != nil {
		return Report{}, err
	}

	opsPerAppSession := numOperates + 3 // create + deposit + operates + close
	totalOps := numAppSessions * opsPerAppSession

	fmt.Printf("  Sender:       %s\n", senderAddr)
	fmt.Printf("  AppSessions:        %d (%d participants each)\n", numAppSessions, numParticipants)
	fmt.Printf("  Asset:        %s\n", asset)
	fmt.Printf("  Amount:       %s per deposit\n", amount.String())
	fmt.Printf("  Operates:     %d per session\n", numOperates)
	fmt.Printf("  Operations:   %d (%d per pipe)\n", totalOps, opsPerAppSession)

	// Derive wallets and create signers
	fmt.Println("  Deriving wallets...")
	sessions := make([]pipe, numAppSessions)
	for p := range numAppSessions {
		sessions[p].wallets = make([]wallet, numParticipants)
		sessions[p].signers = make([]*app.AppSessionSignerV1, numParticipants)
		for w := range numParticipants {
			key := deriveAppSessionKey(cfg.PrivateKey, p, w)
			addr, err := walletAddressFromKey(key)
			if err != nil {
				return Report{}, fmt.Errorf("derive wallet pipe=%d wallet=%d: %w", p, w, err)
			}
			sessions[p].wallets[w] = wallet{key: key, addr: addr}

			msgSigner, err := sign.NewEthereumMsgSigner(key)
			if err != nil {
				return Report{}, fmt.Errorf("create msg signer pipe=%d wallet=%d: %w", p, w, err)
			}
			appSigner, err := app.NewAppSessionWalletSignerV1(msgSigner)
			if err != nil {
				return Report{}, fmt.Errorf("create app signer pipe=%d wallet=%d: %w", p, w, err)
			}
			sessions[p].signers[w] = appSigner
		}
	}

	// Pre-generate all signatures
	fmt.Printf("  Pre-generating signatures (%d app_sessions x %d participants x %d steps)...\n", numAppSessions, numParticipants, opsPerAppSession)
	baseNonce := uint64(time.Now().UnixNano())
	allSigs := make([]appSessionSigs, numAppSessions)
	for p := range numAppSessions {
		sigs, err := preGenerateSigs(&sessions[p], numParticipants, numOperates, baseNonce+uint64(p), asset, amount)
		if err != nil {
			return Report{}, fmt.Errorf("pre-generate pipe %d: %w", p, err)
		}
		allSigs[p] = sigs
	}
	fmt.Println("  Signatures ready.")

	// Connect sender
	fmt.Println("  Connecting sender...")
	senderClient, err := createClient(cfg.WsURL, cfg.PrivateKey)
	if err != nil {
		return Report{}, fmt.Errorf("failed to connect sender: %w", err)
	}
	defer senderClient.Close()

	// Connect pipe leads
	fmt.Printf("  Connecting %d pipe leads...\n", numAppSessions)
	for p := range numAppSessions {
		c, err := createClient(cfg.WsURL, sessions[p].wallets[0].key)
		if err != nil {
			for pp := range p {
				if sessions[pp].wallets[0].client != nil {
					sessions[pp].wallets[0].client.Close()
				}
			}
			return Report{}, fmt.Errorf("connect pipe %d: %w", p, err)
		}
		sessions[p].wallets[0].client = c
		if (p+1)%10 == 0 || p+1 == numAppSessions {
			fmt.Printf("\r    Connected: %d/%d  ", p+1, numAppSessions)
		}
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println()
	defer func() {
		for p := range numAppSessions {
			if sessions[p].wallets[0].client != nil {
				sessions[p].wallets[0].client.Close()
			}
		}
	}()

	// Fund app session leads
	fmt.Printf("  Funding %d app session leads (%s each)...\n", numAppSessions, amount.String())
	for p := range numAppSessions {
		if ctx.Err() != nil {
			return Report{}, ctx.Err()
		}
		_, err := senderClient.Transfer(ctx, sessions[p].wallets[0].addr, asset, amount)
		if err != nil {
			return Report{}, fmt.Errorf("fund pipe %d: %w", p, err)
		}
		if (p+1)%10 == 0 || p+1 == numAppSessions {
			fmt.Printf("\r    Funded: %d/%d  ", p+1, numAppSessions)
		}
	}
	fmt.Println()

	// Stress test
	fmt.Printf("  Stress test (%d app_sessions x %d ops)...\n", numAppSessions, opsPerAppSession)

	results := make([]Result, totalOps)
	var completed int64
	start := time.Now()

	var wg sync.WaitGroup
	for p := range numAppSessions {
		wg.Add(1)
		go func(pipeIdx int) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}

			client := sessions[pipeIdx].wallets[0].client
			base := pipeIdx * opsPerAppSession
			rs := allSigs[pipeIdx]
			step := int64(totalOps)/20 + 1

			record := func(idx int, reqStart time.Time, err error) {
				results[base+idx] = Result{Duration: time.Since(reqStart), Err: err}
				c := atomic.AddInt64(&completed, 1)
				if c%step == 0 || c == int64(totalOps) {
					fmt.Printf("\r    Progress: %d/%d (%.0f%%)  ", c, totalOps, float64(c)/float64(totalOps)*100)
				}
			}

			skipRemaining := func(fromIdx int, cause error) {
				for k := fromIdx; k < opsPerAppSession; k++ {
					results[base+k] = Result{Err: fmt.Errorf("skipped: %w", cause)}
					atomic.AddInt64(&completed, 1)
				}
			}

			// 1. Create (op 0)
			t := time.Now()
			_, _, _, err := client.CreateAppSession(ctx, rs.definition, "{}", rs.createSigs)
			record(0, t, err)
			if err != nil {
				skipRemaining(1, err)
				return
			}

			// 2. Deposit + Operates + Close (steps 0..len-1 map to ops 1..len)
			for i, s := range rs.steps {
				if ctx.Err() != nil {
					skipRemaining(1+i, ctx.Err())
					return
				}

				t = time.Now()
				if s.update.Intent == app.AppStateUpdateIntentDeposit {
					_, err = client.SubmitAppSessionDeposit(ctx, s.update, s.sigs, asset, amount)
				} else {
					err = client.SubmitAppState(ctx, s.update, s.sigs)
				}
				record(1+i, t, err)

				if err != nil {
					skipRemaining(2+i, err)
					return
				}
			}
		}(p)
	}

	wg.Wait()
	totalTime := time.Since(start)
	fmt.Println()

	return ComputeReport("app-session-lifecycle", totalOps, numAppSessions, results, totalTime), nil
}

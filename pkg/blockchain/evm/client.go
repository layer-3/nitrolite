package evm

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
)

var _ core.Client = &Client{}

type Client struct {
	contract        *ChannelHub
	transactOpts    *bind.TransactOpts
	nodeAddress     common.Address
	contractAddress common.Address
	blockchainID    uint64
	assetStore      AssetStore
	evmClient       EVMClient

	requireCheckAllowance bool
	requireCheckBalance   bool
	checkFeeFn            func(ctx context.Context, account common.Address) error
}

type ClientOption interface {
	apply(c *Client)
}

func NewClient(
	contractAddress common.Address,
	evmClient EVMClient,
	txSigner sign.Signer,
	blockchainID uint64,
	nodeAddress string,
	assetStore AssetStore,
	opts ...ClientOption,
) (*Client, error) {
	contract, err := NewChannelHub(contractAddress, evmClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ChannelHub contract instance")
	}
	client := &Client{
		contract:        contract,
		transactOpts:    signerTxOpts(txSigner, blockchainID),
		nodeAddress:     common.HexToAddress(nodeAddress),
		contractAddress: contractAddress,
		blockchainID:    blockchainID,
		assetStore:      assetStore,
		evmClient:       evmClient,

		requireCheckAllowance: true,
		requireCheckBalance:   true,
		checkFeeFn:            getDefaultCheckFeeFn(evmClient),
	}
	for _, opt := range opts {
		opt.apply(client)
	}

	return client, nil
}

// ========= Getters - IVault =========

func (c *Client) GetAccountsBalances(accounts []string, tokens []string) ([][]decimal.Decimal, error) {
	if len(accounts) == 0 || len(tokens) == 0 {
		return [][]decimal.Decimal{}, nil
	}

	result := make([][]decimal.Decimal, len(accounts))
	for i, account := range accounts {
		result[i] = make([]decimal.Decimal, len(tokens))
		accountAddr := common.HexToAddress(account)

		for j, token := range tokens {
			tokenAddr := common.HexToAddress(token)
			balance, err := c.contract.GetAccountBalance(nil, accountAddr, tokenAddr)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get balance for account %s and token %s", account, token)
			}
			result[i][j] = decimal.NewFromBigInt(balance, 0)
		}
	}

	return result, nil
}

func (c *Client) getAllowance(asset string, owner string) (decimal.Decimal, error) {
	tokenAddrHex, err := c.assetStore.GetTokenAddress(asset, c.blockchainID)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to get token address")
	}
	tokenAddr := common.HexToAddress(tokenAddrHex)

	// Native tokens don't require allowance
	if tokenAddr == (common.Address{}) {
		return decimal.New(1, 18), nil
	}

	ownerAddr := common.HexToAddress(owner)
	erc20Contract, err := NewIERC20(tokenAddr, c.evmClient)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to instantiate token contract")
	}
	allowance, err := erc20Contract.Allowance(&bind.CallOpts{}, ownerAddr, c.contractAddress)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get allowance for token %s", asset)
	}

	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, tokenAddrHex)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get token decimals")
	}

	return decimal.NewFromBigInt(allowance, -int32(decimals)), nil
}

func (c *Client) GetTokenBalance(asset string, walletAddress string) (decimal.Decimal, error) {
	tokenAddrHex, err := c.assetStore.GetTokenAddress(asset, c.blockchainID)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to get token address")
	}

	tokenAddr := common.HexToAddress(tokenAddrHex)
	walletAddr := common.HexToAddress(walletAddress)

	// Native token (zero address) — query ETH balance directly
	if tokenAddr == (common.Address{}) {
		balance, err := c.evmClient.BalanceAt(context.Background(), walletAddr, nil)
		if err != nil {
			return decimal.Zero, errors.Wrapf(err, "failed to get native balance for wallet %s", walletAddress)
		}
		// Native tokens use 18 decimals
		return decimal.NewFromBigInt(balance, -18), nil
	}

	tokenContract, err := NewIERC20(tokenAddr, c.evmClient)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to instantiate token contract")
	}

	balance, err := tokenContract.BalanceOf(&bind.CallOpts{}, walletAddr)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get balance for wallet %s on token %s", walletAddress, tokenAddrHex)
	}
	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, tokenAddrHex)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get token decimals for %s", tokenAddrHex)
	}

	return decimal.NewFromBigInt(balance, -int32(decimals)), nil
}

// ========= Getters - ChannelsHub =========

func (c *Client) GetNodeBalance(token string) (decimal.Decimal, error) {
	tokenAddr := common.HexToAddress(token)
	balance, err := c.contract.GetAccountBalance(nil, c.nodeAddress, tokenAddr)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get node balance for token %s", token)
	}
	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, token)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get token decimals for token %s", token)
	}
	return decimal.NewFromBigInt(balance, -int32(decimals)), nil
}

func (c *Client) GetOpenChannels(user string) ([]string, error) {
	userAddr := common.HexToAddress(user)
	channelIDs, err := c.contract.GetOpenChannels(nil, userAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get open channels for user %s", user)
	}

	result := make([]string, len(channelIDs))
	for i, id := range channelIDs {
		result[i] = hexutil.Encode(id[:])
	}
	return result, nil
}

func (c *Client) GetHomeChannelData(homeChannelID string) (core.HomeChannelDataResponse, error) {
	channelIDBytes, err := hexToBytes32(homeChannelID)
	if err != nil {
		return core.HomeChannelDataResponse{}, errors.Wrap(err, "invalid channel ID")
	}

	data, err := c.contract.GetChannelData(nil, channelIDBytes)
	if err != nil {
		return core.HomeChannelDataResponse{}, errors.Wrapf(err, "failed to get channel data for channel %s", homeChannelID)
	}

	lastState, err := contractStateToCoreState(data.LastState, homeChannelID, nil)
	if err != nil {
		return core.HomeChannelDataResponse{}, errors.Wrap(err, "failed to convert contract state")
	}

	return core.HomeChannelDataResponse{
		Definition: core.ChannelDefinition{
			Nonce:     data.Definition.Nonce,
			Challenge: data.Definition.ChallengeDuration,
		},
		Node:            data.Definition.Node.Hex(),
		LastState:       *lastState,
		ChallengeExpiry: data.ChallengeExpiry.Uint64(),
	}, nil
}

func (c *Client) GetEscrowDepositData(escrowChannelID string) (core.EscrowDepositDataResponse, error) {
	escrowIDBytes, err := hexToBytes32(escrowChannelID)
	if err != nil {
		return core.EscrowDepositDataResponse{}, errors.Wrap(err, "invalid escrow ID")
	}

	data, err := c.contract.GetEscrowDepositData(nil, escrowIDBytes)
	if err != nil {
		return core.EscrowDepositDataResponse{}, errors.Wrapf(err, "failed to get escrow deposit data for escrow %s", escrowChannelID)
	}

	lastState, err := contractStateToCoreState(data.InitState, "", &escrowChannelID)
	if err != nil {
		return core.EscrowDepositDataResponse{}, errors.Wrap(err, "failed to convert contract state")
	}

	return core.EscrowDepositDataResponse{
		EscrowChannelID: escrowChannelID,
		Node:            c.contractAddress.Hex(),
		LastState:       *lastState,
		UnlockExpiry:    data.UnlockAt,
		ChallengeExpiry: data.ChallengeExpiry,
	}, nil
}

func (c *Client) GetEscrowWithdrawalData(escrowChannelID string) (core.EscrowWithdrawalDataResponse, error) {
	escrowIDBytes, err := hexToBytes32(escrowChannelID)
	if err != nil {
		return core.EscrowWithdrawalDataResponse{}, errors.Wrap(err, "invalid escrow ID")
	}

	data, err := c.contract.GetEscrowWithdrawalData(nil, escrowIDBytes)
	if err != nil {
		return core.EscrowWithdrawalDataResponse{}, errors.Wrapf(err, "failed to get escrow withdrawal data for escrow %s", escrowChannelID)
	}

	lastState, err := contractStateToCoreState(data.InitState, "", &escrowChannelID)
	if err != nil {
		return core.EscrowWithdrawalDataResponse{}, errors.Wrap(err, "failed to convert contract state")
	}

	return core.EscrowWithdrawalDataResponse{
		EscrowChannelID: escrowChannelID,
		Node:            c.contractAddress.Hex(),
		LastState:       *lastState,
	}, nil
}

// ========= IVault Functions =========

func (c *Client) Deposit(node, token string, amount decimal.Decimal) (string, error) {
	nodeAddr := common.HexToAddress(node)
	tokenAddr := common.HexToAddress(token)

	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, token)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get token decimals for token %s", token)
	}
	amountBig, err := core.DecimalToBigInt(amount, decimals)
	if err != nil {
		return "", errors.Wrapf(err, "failed to convert amount %s to big.Int", amount.String())
	}

	if tokenAddr == (common.Address{}) {
		c.transactOpts.Value = amountBig
		defer func() { c.transactOpts.Value = nil }()
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.DepositToVault(c.transactOpts, nodeAddr, tokenAddr, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to deposit to vault")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) Withdraw(node, token string, amount decimal.Decimal) (string, error) {
	nodeAddr := common.HexToAddress(node)
	tokenAddr := common.HexToAddress(token)

	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, token)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get token decimals for token %s", token)
	}
	amountBig, err := core.DecimalToBigInt(amount, decimals)
	if err != nil {
		return "", errors.Wrapf(err, "failed to convert amount %s to big.Int", amount.String())
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.WithdrawFromVault(c.transactOpts, nodeAddr, tokenAddr, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to withdraw from vault")
	}

	return tx.Hash().Hex(), nil
}

// ========= Getters - ERC20 =========

func (c *Client) Approve(asset string, amount decimal.Decimal) (string, error) {
	tokenAddrHex, err := c.assetStore.GetTokenAddress(asset, c.blockchainID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token address")
	}

	tokenAddr := common.HexToAddress(tokenAddrHex)
	if tokenAddr == (common.Address{}) {
		return "", errors.New("native tokens do not require approval")
	}

	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, tokenAddrHex)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get token decimals for %s", asset)
	}

	amountBig, err := core.DecimalToBigInt(amount, decimals)
	if err != nil {
		return "", errors.Wrapf(err, "failed to convert amount %s to big.Int", amount.String())
	}

	erc20Contract, err := NewIERC20(tokenAddr, c.evmClient)
	if err != nil {
		return "", errors.Wrap(err, "failed to instantiate token contract")
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := erc20Contract.Approve(c.transactOpts, c.contractAddress, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to approve token spending")
	}

	return tx.Hash().Hex(), nil
}

// ========= Channel Lifecycle =========

func (c *Client) Create(def core.ChannelDefinition, initCCS core.State) (string, error) {
	contractDef, err := coreDefToContractDef(def, initCCS.Asset, initCCS.UserWallet, c.nodeAddress)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert channel definition")
	}

	contractState, err := coreStateToContractState(initCCS, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert state")
	}

	switch contractState.Intent {
	case core.INTENT_OPERATE:
	case core.INTENT_WITHDRAW:
	case core.INTENT_DEPOSIT:
		if c.requireCheckAllowance {
			allowance, err := c.getAllowance(initCCS.Asset, initCCS.UserWallet)
			if err != nil {
				return "", errors.Wrap(err, "failed to get allowance")
			}
			if allowance.LessThan(initCCS.Transition.Amount) {
				return "", errors.New("allowance is not sufficient to cover the deposit amount")
			}

		}
		if c.requireCheckBalance {
			tokenBalance, err := c.GetTokenBalance(initCCS.Asset, initCCS.UserWallet)
			if err != nil {
				return "", errors.Wrap(err, "failed to check token balance")
			}
			if tokenBalance.LessThan(initCCS.Transition.Amount) {
				return "", errors.New("balance is not sufficient to cover the deposit amount")
			}
		}

		if contractState.HomeLedger.Token == (common.Address{}) {
			value, err := core.DecimalToBigInt(initCCS.Transition.Amount, contractState.HomeLedger.Decimals)
			if err != nil {
				return "", errors.Wrap(err, "failed to convert native deposit amount to wei")
			}
			c.transactOpts.Value = value
			defer func() { c.transactOpts.Value = nil }()
		}

	default:
		return "", errors.New("unsupported intent for create: " + string(contractState.Intent))
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.CreateChannel(c.transactOpts, contractDef, contractState)
	if err != nil {
		return "", errors.Wrap(err, "failed to create channel")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) MigrateChannelHere(def core.ChannelDefinition, candidate core.State) (string, error) {
	contractDef, err := coreDefToContractDef(def, candidate.Asset, candidate.UserWallet, c.nodeAddress)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert channel definition")
	}

	contractCandidate, err := coreStateToContractState(candidate, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert candidate state")
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.InitiateMigration(c.transactOpts, contractDef, contractCandidate)
	if err != nil {
		return "", errors.Wrap(err, "failed to initiate migration")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) Checkpoint(candidate core.State) (string, error) {
	if candidate.HomeChannelID == nil {
		return "", errors.New("candidate state must have a home channel ID")
	}

	channelIDBytes, err := hexToBytes32(*candidate.HomeChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid channel ID")
	}

	contractCandidate, err := coreStateToContractState(candidate, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert candidate state")
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	var tx *types.Transaction
	switch contractCandidate.Intent {
	case core.INTENT_OPERATE:
		// TODO: recheck proofs logic
		tx, err = c.contract.CheckpointChannel(c.transactOpts, channelIDBytes, contractCandidate)
	case core.INTENT_DEPOSIT:
		if c.requireCheckAllowance {
			allowance, err := c.getAllowance(candidate.Asset, candidate.UserWallet)
			if err != nil {
				return "", errors.Wrap(err, "failed to get allowance")
			}
			if allowance.LessThan(candidate.Transition.Amount) {
				return "", errors.New("allowance is not sufficient to cover the deposit amount")
			}

		}
		if c.requireCheckBalance {
			tokenBalance, err := c.GetTokenBalance(candidate.Asset, candidate.UserWallet)
			if err != nil {
				return "", errors.Wrap(err, "failed to check token balance")
			}
			if tokenBalance.LessThan(candidate.Transition.Amount) {
				return "", errors.New("balance is not sufficient to cover the deposit amount")
			}
		}

		if contractCandidate.HomeLedger.Token == (common.Address{}) {
			value, valueErr := core.DecimalToBigInt(candidate.Transition.Amount, contractCandidate.HomeLedger.Decimals)
			if valueErr != nil {
				return "", errors.Wrap(valueErr, "failed to convert native deposit amount to wei")
			}
			c.transactOpts.Value = value
			defer func() { c.transactOpts.Value = nil }()
		}

		tx, err = c.contract.DepositToChannel(c.transactOpts, channelIDBytes, contractCandidate)
	case core.INTENT_WITHDRAW:
		tx, err = c.contract.WithdrawFromChannel(c.transactOpts, channelIDBytes, contractCandidate)
	default:
		return "", errors.New("unsupported intent for checkpointing: " + string(contractCandidate.Intent))
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to checkpoint channel")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) Challenge(candidate core.State, challengerSig []byte, challengerIdx core.ChannelParticipant) (string, error) {
	if candidate.HomeChannelID == nil {
		return "", errors.New("candidate state must have a home channel ID")
	}

	channelIDBytes, err := hexToBytes32(*candidate.HomeChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid channel ID")
	}

	contractCandidate, err := coreStateToContractState(candidate, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert candidate state")
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	// TODO: recheck proofs logic
	tx, err := c.contract.ChallengeChannel(c.transactOpts, channelIDBytes, contractCandidate, challengerSig, uint8(challengerIdx))
	if err != nil {
		return "", errors.Wrap(err, "failed to challenge channel")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) Close(candidate core.State) (string, error) {
	if candidate.HomeChannelID == nil {
		return "", errors.New("candidate state must have a home channel ID")
	}

	channelIDBytes, err := hexToBytes32(*candidate.HomeChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid channel ID")
	}

	contractCandidate, err := coreStateToContractState(candidate, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert candidate state")
	}

	if contractCandidate.Intent != core.INTENT_CLOSE {
		return "", errors.New("unsupported intent for close: " + string(contractCandidate.Intent))
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	// TODO: recheck proof logic
	tx, err := c.contract.CloseChannel(c.transactOpts, channelIDBytes, contractCandidate)
	if err != nil {
		return "", errors.Wrap(err, "failed to close channel")
	}

	return tx.Hash().Hex(), nil
}

// ========= Escrow Deposit =========

func (c *Client) InitiateEscrowDeposit(def core.ChannelDefinition, initCCS core.State) (string, error) {
	contractDef, err := coreDefToContractDef(def, initCCS.Asset, initCCS.UserWallet, c.nodeAddress)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert channel definition")
	}

	contractState, err := coreStateToContractState(initCCS, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert state")
	}

	if contractState.Intent != core.INTENT_INITIATE_ESCROW_DEPOSIT {
		return "", errors.New("unsupported intent for initiate escrow deposit: " + string(contractState.Intent))
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.InitiateEscrowDeposit(c.transactOpts, contractDef, contractState)
	if err != nil {
		return "", errors.Wrap(err, "failed to initiate escrow deposit")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) ChallengeEscrowDeposit(candidate core.State, challengerSig []byte, challengerIdx core.ChannelParticipant) (string, error) {
	if candidate.EscrowChannelID == nil {
		return "", errors.New("candidate state must have an escrow channel ID")
	}

	escrowIDBytes, err := hexToBytes32(*candidate.EscrowChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid escrow ID")
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.ChallengeEscrowDeposit(c.transactOpts, escrowIDBytes, challengerSig, uint8(challengerIdx))
	if err != nil {
		return "", errors.Wrap(err, "failed to challenge escrow deposit")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) FinalizeEscrowDeposit(candidate core.State) (string, error) {
	if candidate.EscrowChannelID == nil {
		return "", errors.New("candidate state must have an escrow channel ID")
	}

	escrowIDBytes, err := hexToBytes32(*candidate.EscrowChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid escrow ID")
	}

	contractCandidate, err := coreStateToContractState(candidate, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert candidate state")
	}

	if contractCandidate.Intent != core.INTENT_FINALIZE_ESCROW_DEPOSIT {
		return "", errors.New("unsupported intent for finalize escrow deposit: " + string(contractCandidate.Intent))
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.FinalizeEscrowDeposit(c.transactOpts, escrowIDBytes, contractCandidate)
	if err != nil {
		return "", errors.Wrap(err, "failed to finalize escrow deposit")
	}

	return tx.Hash().Hex(), nil
}

// ========= Escrow Withdrawal =========

func (c *Client) InitiateEscrowWithdrawal(def core.ChannelDefinition, initCCS core.State) (string, error) {
	contractDef, err := coreDefToContractDef(def, initCCS.Asset, initCCS.UserWallet, c.nodeAddress)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert channel definition")
	}

	contractState, err := coreStateToContractState(initCCS, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert state")
	}

	if contractState.Intent != core.INTENT_INITIATE_ESCROW_WITHDRAWAL {
		return "", errors.New("unsupported intent for initiate escrow withdrawal: " + string(contractState.Intent))
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.InitiateEscrowWithdrawal(c.transactOpts, contractDef, contractState)
	if err != nil {
		return "", errors.Wrap(err, "failed to initiate escrow withdrawal")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) ChallengeEscrowWithdrawal(candidate core.State, challengerSig []byte, challengerIdx core.ChannelParticipant) (string, error) {
	if candidate.EscrowChannelID == nil {
		return "", errors.New("candidate state must have an escrow channel ID")
	}

	escrowIDBytes, err := hexToBytes32(*candidate.EscrowChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid escrow ID")
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.ChallengeEscrowWithdrawal(c.transactOpts, escrowIDBytes, challengerSig, uint8(challengerIdx))
	if err != nil {
		return "", errors.Wrap(err, "failed to challenge escrow withdrawal")
	}

	return tx.Hash().Hex(), nil
}

func (c *Client) FinalizeEscrowWithdrawal(candidate core.State) (string, error) {
	if candidate.EscrowChannelID == nil {
		return "", errors.New("candidate state must have an escrow channel ID")
	}
	if candidate.EscrowLedger == nil {
		return "", errors.New("candidate state must have an escrow channel ID")
	}

	escrowIDBytes, err := hexToBytes32(*candidate.EscrowChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid escrow ID")
	}

	contractCandidate, err := coreStateToContractState(candidate, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert candidate state")
	}

	if contractCandidate.Intent != core.INTENT_INITIATE_ESCROW_WITHDRAWAL {
		return "", errors.New("unsupported intent for initiate escrow withdrawal: " + string(contractCandidate.Intent))
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.FinalizeEscrowWithdrawal(c.transactOpts, escrowIDBytes, contractCandidate)
	if err != nil {
		return "", errors.Wrap(err, "failed to finalize escrow withdrawal")
	}

	return tx.Hash().Hex(), nil
}

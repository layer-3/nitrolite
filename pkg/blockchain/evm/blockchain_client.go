package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
)

var _ core.BlockchainClient = &BlockchainClient{}

type BlockchainClient struct {
	BaseClient
	contract                  *ChannelHub
	transactOpts              *bind.TransactOpts
	txSigner                  sign.Signer
	nodeAddress               common.Address
	channelHubContractAddress common.Address
	assetStore                AssetStore

	requireCheckAllowance bool
	requireCheckBalance   bool
	checkFeeFn            func(ctx context.Context, account common.Address) error
}

type ClientOption interface {
	apply(c *BlockchainClient)
}

func NewBlockchainClient(
	contractAddress common.Address,
	evmClient EVMClient,
	txSigner sign.Signer,
	blockchainID uint64,
	nodeAddress string,
	assetStore AssetStore,
	opts ...ClientOption,
) (*BlockchainClient, error) {
	contract, err := NewChannelHub(contractAddress, evmClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ChannelHub contract instance")
	}
	client := &BlockchainClient{
		BaseClient: BaseClient{
			evmClient:    evmClient,
			blockchainID: blockchainID,
		},
		contract:                  contract,
		transactOpts:              signerTxOpts(txSigner, blockchainID),
		txSigner:                  txSigner,
		nodeAddress:               common.HexToAddress(nodeAddress),
		channelHubContractAddress: contractAddress,
		assetStore:                assetStore,

		requireCheckAllowance: true,
		requireCheckBalance:   true,
		checkFeeFn:            getDefaultCheckFeeFn(evmClient),
	}
	for _, opt := range opts {
		opt.apply(client)
	}

	return client, nil
}

func (c *BlockchainClient) getAllowance(asset string, owner string) (decimal.Decimal, error) {
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
	allowance, err := erc20Contract.Allowance(&bind.CallOpts{}, ownerAddr, c.channelHubContractAddress)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get allowance for token %s", asset)
	}

	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, tokenAddrHex)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get token decimals")
	}

	return decimal.NewFromBigInt(allowance, -int32(decimals)), nil
}

func (c *BlockchainClient) GetTokenBalance(asset string, walletAddress string) (decimal.Decimal, error) {
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

func (c *BlockchainClient) GetNodeBalance(token string) (decimal.Decimal, error) {
	tokenAddr := common.HexToAddress(token)
	balance, err := c.contract.GetNodeBalance(nil, tokenAddr)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get node balance for token %s", token)
	}
	decimals, err := c.assetStore.GetTokenDecimals(c.blockchainID, token)
	if err != nil {
		return decimal.Zero, errors.Wrapf(err, "failed to get token decimals for token %s", token)
	}
	return decimal.NewFromBigInt(balance, -int32(decimals)), nil
}

func (c *BlockchainClient) GetOpenChannels(user string) ([]string, error) {
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

func (c *BlockchainClient) GetHomeChannelData(homeChannelID string) (core.HomeChannelDataResponse, error) {
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

func (c *BlockchainClient) GetEscrowDepositData(escrowChannelID string) (core.EscrowDepositDataResponse, error) {
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
		Node:            c.nodeAddress.Hex(),
		LastState:       *lastState,
		UnlockExpiry:    data.UnlockAt,
		ChallengeExpiry: data.ChallengeExpiry,
	}, nil
}

func (c *BlockchainClient) GetEscrowWithdrawalData(escrowChannelID string) (core.EscrowWithdrawalDataResponse, error) {
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
		Node:            c.nodeAddress.Hex(),
		LastState:       *lastState,
	}, nil
}

// ========= IVault Functions =========

func (c *BlockchainClient) Deposit(token string, amount decimal.Decimal) (string, error) {
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

	tx, err := c.contract.DepositToNode(c.transactOpts, tokenAddr, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to deposit to node")
	}

	return tx.Hash().Hex(), nil
}

func (c *BlockchainClient) Withdraw(to, token string, amount decimal.Decimal) (string, error) {
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

	tx, err := c.contract.WithdrawFromNode(c.transactOpts, common.HexToAddress(to), tokenAddr, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to withdraw from node")
	}

	return tx.Hash().Hex(), nil
}

// ========= Getters - ERC20 =========

func (c *BlockchainClient) Approve(asset string, amount decimal.Decimal) (string, error) {
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

	tx, err := erc20Contract.Approve(c.transactOpts, c.channelHubContractAddress, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to approve token spending")
	}

	return tx.Hash().Hex(), nil
}

// ========= Channel Lifecycle =========

func (c *BlockchainClient) Create(def core.ChannelDefinition, initCCS core.State) (string, error) {
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

func (c *BlockchainClient) MigrateChannelHere(def core.ChannelDefinition, candidate core.State) (string, error) {
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

func (c *BlockchainClient) Checkpoint(candidate core.State) (string, error) {
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

func (c *BlockchainClient) Challenge(candidate core.State, challengerSig []byte, challengerIdx core.ChannelParticipant) (string, error) {
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

func (c *BlockchainClient) Close(candidate core.State) (string, error) {
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

func (c *BlockchainClient) InitiateEscrowDeposit(def core.ChannelDefinition, initCCS core.State) (string, error) {
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

func (c *BlockchainClient) ChallengeEscrowDeposit(candidate core.State, challengerSig []byte, challengerIdx core.ChannelParticipant) (string, error) {
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

func (c *BlockchainClient) FinalizeEscrowDeposit(candidate core.State) (string, error) {
	if candidate.HomeChannelID == nil {
		return "", errors.New("candidate state must have a home channel ID")
	}
	if candidate.EscrowChannelID == nil {
		return "", errors.New("candidate state must have an escrow channel ID")
	}

	channelIDBytes, err := hexToBytes32(*candidate.HomeChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid channel ID")
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

	tx, err := c.contract.FinalizeEscrowDeposit(c.transactOpts, channelIDBytes, escrowIDBytes, contractCandidate)
	if err != nil {
		return "", errors.Wrap(err, "failed to finalize escrow deposit")
	}

	return tx.Hash().Hex(), nil
}

// ========= Escrow Withdrawal =========

func (c *BlockchainClient) InitiateEscrowWithdrawal(def core.ChannelDefinition, initCCS core.State) (string, error) {
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

func (c *BlockchainClient) ChallengeEscrowWithdrawal(candidate core.State, challengerSig []byte, challengerIdx core.ChannelParticipant) (string, error) {
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

func (c *BlockchainClient) FinalizeEscrowWithdrawal(candidate core.State) (string, error) {
	if candidate.HomeChannelID == nil {
		return "", errors.New("candidate state must have a home channel ID")
	}
	if candidate.EscrowChannelID == nil {
		return "", errors.New("candidate state must have an escrow channel ID")
	}
	if candidate.EscrowLedger == nil {
		return "", errors.New("candidate state must have an escrow ledger")
	}

	channelIDBytes, err := hexToBytes32(*candidate.HomeChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid channel ID")
	}

	escrowIDBytes, err := hexToBytes32(*candidate.EscrowChannelID)
	if err != nil {
		return "", errors.Wrap(err, "invalid escrow ID")
	}

	contractCandidate, err := coreStateToContractState(candidate, c.assetStore.GetTokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert candidate state")
	}

	if contractCandidate.Intent != core.INTENT_FINALIZE_ESCROW_WITHDRAWAL {
		return "", errors.New("unsupported intent for finalize escrow withdrawal: " + string(contractCandidate.Intent))
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return "", err
	}

	tx, err := c.contract.FinalizeEscrowWithdrawal(c.transactOpts, channelIDBytes, escrowIDBytes, contractCandidate)
	if err != nil {
		return "", errors.Wrap(err, "failed to finalize escrow withdrawal")
	}

	return tx.Hash().Hex(), nil
}

func (c *BlockchainClient) EnsureSigValidatorRegistered(validatorID uint8, validatorAddress string, checkOnly bool) error {
	validatorAddr := common.HexToAddress(validatorAddress)

	validatorInfo, err := c.contract.GetNodeValidator(nil, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to check if validator %d is registered", validatorID)
	}
	if validatorInfo.Validator.Hex() == validatorAddr.Hex() {
		return nil
	} else if validatorInfo.Validator != (common.Address{}) {
		return errors.Errorf("validator ID %d is already registered with a different address %s", validatorID, validatorInfo.Validator.Hex())
	}

	if checkOnly {
		return errors.Errorf("validator ID %d with address %s is not registered; run 'clearnode operator register-validator' to register", validatorID, validatorAddress)
	}

	if err := c.checkFeeFn(context.Background(), c.transactOpts.From); err != nil {
		return err
	}

	uint8Type, _ := abi.NewType("uint8", "", nil)
	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	args := abi.Arguments{
		{Type: uint256Type},
		{Type: addressType},
		{Type: uint8Type},
		{Type: addressType},
	}
	message, err := args.Pack(new(big.Int).SetUint64(c.blockchainID), c.channelHubContractAddress, validatorID, validatorAddr)
	if err != nil {
		return errors.Wrap(err, "failed to encode validator registration message")
	}

	sig, err := c.txSigner.Sign(sign.ComputeEthereumSignedMessageHash(message))
	if err != nil {
		return errors.Wrap(err, "failed to sign validator registration message")
	}

	_, err = c.contract.RegisterNodeValidator(c.transactOpts, validatorID, validatorAddr, sig)
	if err != nil {
		return errors.Wrapf(err, "failed to register validator %d with address %s", validatorID, validatorAddress)
	}

	return nil
}

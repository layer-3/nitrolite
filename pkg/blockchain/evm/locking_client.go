package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// LockingClient provides access to a Locking contract.
type LockingClient struct {
	BaseClient
	lockingContractAddress common.Address
	transactOpts           *bind.TransactOpts

	tokenAddress  common.Address
	tokenDecimals uint8
}

// NewLockingClient creates a new LockingClient.
// If txSigner is provided, the client can perform write operations (lock, relock, unlock, withdraw).
func NewLockingClient(lockingContractAddress common.Address, evmClient EVMClient, blockchainID uint64, txSigner ...sign.Signer) (*LockingClient, error) {
	c := &LockingClient{
		BaseClient: BaseClient{
			evmClient:    evmClient,
			blockchainID: blockchainID,
		},
		lockingContractAddress: lockingContractAddress,
	}
	if len(txSigner) > 0 && txSigner[0] != nil {
		c.transactOpts = signerTxOpts(txSigner[0], blockchainID)
	}

	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate Locking contract")
	}

	tokenAddress, err := lockingContract.Asset(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get asset from the Locking contract")
	}

	erc20Contract, err := NewIERC20(tokenAddress, c.evmClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate ERC20 contract")
	}

	decimals, err := erc20Contract.Decimals(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decimals for token %s", tokenAddress.Hex())
	}

	c.tokenAddress = tokenAddress
	c.tokenDecimals = decimals

	return c, nil
}

// GetTokenDecimals returns the number of decimals for the token used in the AppRegistry.
// This is needed to convert between human-readable amounts and the raw integer amounts used in transactions.
func (c *LockingClient) GetTokenDecimals() (uint8, error) {
	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return 0, errors.Wrap(err, "failed to instantiate Locking contract")
	}

	tokenAddress, err := lockingContract.Asset(&bind.CallOpts{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to get asset from the Locking contract")
	}

	erc20Contract, err := NewIERC20(tokenAddress, c.evmClient)
	if err != nil {
		return 0, errors.Wrap(err, "failed to instantiate ERC20 contract")
	}

	decimals, err := erc20Contract.Decimals(&bind.CallOpts{})
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get decimals for token %s", tokenAddress.Hex())
	}

	return decimals, nil
}

// Lock locks tokens into the Locking contract for the specified target address.
// The caller must have approved the Locking contract to spend the token beforehand.
func (c *LockingClient) Lock(targetWalletAddress string, amount decimal.Decimal) (string, error) {
	if !common.IsHexAddress(targetWalletAddress) {
		return "", errors.Errorf("invalid address %q", targetWalletAddress)
	}

	targetAddr := common.HexToAddress(targetWalletAddress)
	if c.transactOpts == nil {
		return "", errors.New("transaction signer not configured")
	}

	amountBig, err := core.DecimalToUint256(amount, c.tokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert amount with decimal precision")
	}

	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return "", errors.Wrap(err, "failed to instantiate Locking contract")
	}

	tx, err := lockingContract.Lock(c.transactOpts, targetAddr, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to send lock transaction")
	}
	return tx.Hash().Hex(), nil
}

// Relock re-locks tokens that are in the unlocking state back to the locked state.
func (c *LockingClient) Relock() (string, error) {
	if c.transactOpts == nil {
		return "", errors.New("transaction signer not configured")
	}

	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return "", errors.Wrap(err, "failed to instantiate Locking contract")
	}

	tx, err := lockingContract.Relock(c.transactOpts)
	if err != nil {
		return "", errors.Wrap(err, "failed to send relock transaction")
	}
	return tx.Hash().Hex(), nil
}

// Unlock initiates the unlock process for the caller's locked tokens.
// After the unlock period elapses, Withdraw can be called.
func (c *LockingClient) Unlock() (string, error) {
	if c.transactOpts == nil {
		return "", errors.New("transaction signer not configured")
	}

	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return "", errors.Wrap(err, "failed to instantiate Locking contract")
	}

	tx, err := lockingContract.Unlock(c.transactOpts)
	if err != nil {
		return "", errors.Wrap(err, "failed to send unlock transaction")
	}
	return tx.Hash().Hex(), nil
}

// Withdraw withdraws unlocked tokens to the specified destination address.
// Can only be called after the unlock period has elapsed.
func (c *LockingClient) Withdraw(destination string) (string, error) {
	if c.transactOpts == nil {
		return "", errors.New("transaction signer not configured")
	}

	destinationAddr := common.HexToAddress(destination)

	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return "", errors.Wrap(err, "failed to instantiate Locking contract")
	}

	tx, err := lockingContract.Withdraw(c.transactOpts, destinationAddr)
	if err != nil {
		return "", errors.Wrap(err, "failed to send withdraw transaction")
	}
	return tx.Hash().Hex(), nil
}

// ApproveToken approves the Locking contract to spend the specified amount of tokens.
// This must be called before Lock.
func (c *LockingClient) ApproveToken(amount decimal.Decimal) (string, error) {
	if c.transactOpts == nil {
		return "", errors.New("transaction signer not configured")
	}

	amountBig, err := core.DecimalToUint256(amount, c.tokenDecimals)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert amount with decimal precision")
	}

	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return "", errors.Wrap(err, "failed to instantiate Locking contract")
	}

	tokenAddress, err := lockingContract.Asset(&bind.CallOpts{})
	if err != nil {
		return "", errors.Wrap(err, "failed to get asset from Locking contract")
	}

	erc20Contract, err := NewIERC20(tokenAddress, c.evmClient)
	if err != nil {
		return "", errors.Wrap(err, "failed to instantiate ERC20 contract")
	}

	tx, err := erc20Contract.Approve(c.transactOpts, c.lockingContractAddress, amountBig)
	if err != nil {
		return "", errors.Wrap(err, "failed to send approve transaction")
	}
	return tx.Hash().Hex(), nil
}

// GetBalance returns the locked balance of a user in the Locking contraсt.
func (c *LockingClient) GetBalance(user string) (decimal.Decimal, error) {
	userAddr := common.HexToAddress(user)
	lockingContract, err := NewAppRegistry(c.lockingContractAddress, c.evmClient)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to instantiate Locking contract")
	}

	balance, err := lockingContract.BalanceOf(nil, userAddr)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to get balance")
	}

	return decimal.NewFromBigInt(balance, -int32(c.tokenDecimals)), nil
}

// decimalToBigInt converts a decimal amount to *big.Int given token decimals.
func decimalToBigInt(amount decimal.Decimal, decimals int32) *big.Int {
	shifted := amount.Shift(decimals)
	return shifted.BigInt()
}

// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ChannelHub} from "../src/ChannelHub.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";
import {EscrowStatus} from "../src/interfaces/Types.sol";

/**
 * @title TestChannelHub
 * @notice Test harness contract that exposes internal ChannelHub functions for testing
 */
contract TestChannelHub is ChannelHub {
    constructor(ISignatureValidator _defaultSigValidator, address _node) ChannelHub(_defaultSigValidator, _node) {}

    /**
     * @notice Marks this contract as a test contract for Forge
     * @dev Prevents size limit checks from being enforced on this test harness
     */
    function IS_TEST() external pure {}

    /**
     * @notice Exposed version of _nonRevertingPushFunds for testing
     */
    function exposed_nonRevertingPushFunds(address to, address token, uint256 amount) external payable {
        _nonRevertingPushFunds(to, token, amount);
    }

    /**
     * @notice Exposed version of _pullFunds for testing
     */
    function exposed_pullFunds(address from, address token, uint256 amount) external payable {
        _pullFunds(from, token, amount);
    }

    /**
     * @notice Workaround to set a node's vault balance directly for testing
     * @dev Allows tests to set up vault state without going through depositToVault (useful for tokens
     *      that revert on transferFrom, making a normal deposit impossible)
     */
    function workaround_setNodeBalance(address node, address token, uint256 amount) external {
        _nodeBalances[node][token] = amount;
    }

    /**
     * @notice Workaround to set reclaim balance for testing
     * @dev Allows tests to set up reclaim state without going through failed transfers
     */
    function workaround_setReclaim(address account, address token, uint256 amount) external {
        _reclaims[account][token] = amount;
    }

    /**
     * @notice Workaround to write an escrow deposit record directly into storage
     * @dev Only sets the fields relevant to purge/stats logic; avoids going through the full escrow lifecycle
     */
    function workaround_setEscrowDeposit(
        bytes32 escrowId,
        bytes32 channelId,
        EscrowStatus status,
        address user_,
        address /* node_ */,
        uint64 unlockAt,
        uint64 challengeExpireAt,
        uint256 lockedAmount,
        address token
    ) external {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];
        meta.channelId = channelId;
        meta.status = status;
        meta.user = user_;
        meta.unlockAt = unlockAt;
        meta.challengeExpireAt = challengeExpireAt;
        meta.lockedAmount = lockedAmount;
        meta.initState.nonHomeLedger.token = token;
    }

    /**
     * @notice Workaround to append an escrow ID to the ordered purge queue
     */
    function workaround_addEscrowDepositId(bytes32 escrowId) external {
        _escrowDepositIds.push(escrowId);
    }

    /**
     * @notice Exposes the internal _escrowDepositIds array for assertions
     */
    function harness_escrowDepositIds() external view returns (bytes32[] memory) {
        return _escrowDepositIds;
    }

    /**
     * @notice Exposes the internal _purgeEscrowDeposits for direct invocation in tests
     */
    function harness_purgeEscrowDeposits(uint256 maxSteps) external {
        _purgeEscrowDeposits(maxSteps);
    }

    /**
     * @notice Exposed version of _extractValidator for testing
     * @dev Returns only the resolved validator address; sigData slice is not useful to callers
     */
    function exposed_extractValidator(bytes calldata signature, uint256 approvedSignatureValidators)
        external
        view
        returns (ISignatureValidator validator)
    {
        (validator,) = _extractValidator(signature, approvedSignatureValidators);
    }
}

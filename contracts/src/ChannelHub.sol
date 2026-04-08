// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

import {IVault} from "./interfaces/IVault.sol";
import {ISignatureValidator, ValidationResult, VALIDATION_FAILURE} from "./interfaces/ISignatureValidator.sol";
import {
    ChannelDefinition,
    ChannelStatus,
    DEFAULT_SIG_VALIDATOR_ID,
    EscrowStatus,
    State,
    StateIntent,
    ParticipantIndex
} from "./interfaces/Types.sol";

import {Utils} from "./Utils.sol";
import {ChannelEngine} from "./ChannelEngine.sol";
import {EscrowDepositEngine} from "./EscrowDepositEngine.sol";
import {EscrowWithdrawalEngine} from "./EscrowWithdrawalEngine.sol";
import {EcdsaSignatureUtils} from "./sigValidators/EcdsaSignatureUtils.sol";

/**
 * @title ChannelHub
 * @notice Main contract implementing the Nitrolite state channel protocol (single-chain operations)
 * @dev Uses unified transition pattern with ChannelEngine library for validation
 */
contract ChannelHub is IVault, ReentrancyGuard {
    using EnumerableSet for EnumerableSet.Bytes32Set;
    using SafeERC20 for IERC20;
    using SafeCast for int256;
    using SafeCast for uint256;
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;

    event EscrowDepositsPurged(uint256 purgedCount);

    event ChannelCreated(
        bytes32 indexed channelId,
        address indexed user,
        address indexed node,
        ChannelDefinition definition,
        State initialState
    );
    event ChannelDeposited(bytes32 indexed channelId, State candidate);
    event ChannelWithdrawn(bytes32 indexed channelId, State candidate);
    event ChannelCheckpointed(bytes32 indexed channelId, State candidate);
    event ChannelChallenged(bytes32 indexed channelId, State candidate, uint64 challengeExpireAt);
    event ChannelClosed(bytes32 indexed channelId, State finalState);

    event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowDepositChallenged(bytes32 indexed escrowId, State state, uint64 challengeExpireAt);
    event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);

    event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowWithdrawalChallenged(bytes32 indexed escrowId, State state, uint64 challengeExpireAt);
    event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);

    event MigrationOutInitiated(bytes32 indexed channelId, State state);
    event MigrationInInitiated(bytes32 indexed channelId, State state);
    event MigrationOutFinalized(bytes32 indexed channelId, State state);
    event MigrationInFinalized(bytes32 indexed channelId, State state);

    event ValidatorRegistered(uint8 indexed validatorId, ISignatureValidator indexed validator);
    event TransferFailed(address indexed recipient, address indexed token, uint256 amount);
    event FundsClaimed(address indexed account, address indexed token, address indexed destination, uint256 amount);

    event NodeBalanceUpdated(address indexed node, address indexed token, uint256 amount);

    error InvalidAddress();
    error IncorrectAmount();
    error IncorrectValue();
    error NativeTransferFailed(address to, uint256 amount);
    error AddressCollision(address collision);
    error IncorrectChallengeDuration();

    error InvalidValidatorId();
    error ValidatorAlreadyRegistered(uint8 validatorId);
    error ValidatorNotRegistered(uint8 validatorId);
    error ValidatorNotApproved();
    error ValidatorNotActive(uint8 validatorId, uint64 activatesAt);

    error EmptySignature();
    error IncorrectSignature();
    error InsufficientBalance();
    error IncorrectStateIntent();
    error IncorrectChannelStatus();
    error ChallengerVersionTooLow();
    error NoChannelIdFoundForEscrow();
    error IncorrectChannelId();
    error IncorrectNode();

    struct ChannelMeta {
        ChannelStatus status;
        ChannelDefinition definition;
        State lastState;
        uint256 lockedFunds;
        uint64 challengeExpireAt;
    }

    struct EscrowDepositMeta {
        bytes32 channelId;
        EscrowStatus status;
        address user;
        uint256 approvedSignatureValidators;
        uint64 unlockAt;
        uint64 challengeExpireAt;
        uint256 lockedAmount;
        State initState;
    }

    struct EscrowWithdrawalMeta {
        bytes32 channelId;
        EscrowStatus status;
        address user;
        uint256 approvedSignatureValidators;
        uint64 challengeExpireAt;
        uint256 lockedAmount;
        State initState;
    }

    struct NodeValidator {
        ISignatureValidator validator;
        uint64 registeredAt;
    }

    // ======== Contract Storage ==========

    uint8 public constant VERSION = 1;

    ISignatureValidator public immutable DEFAULT_SIG_VALIDATOR;

    address public immutable NODE;

    // TODO: estimate these values better
    uint32 public constant MIN_CHALLENGE_DURATION = 1 days;

    uint32 public constant ESCROW_DEPOSIT_UNLOCK_DELAY = 3 hours;

    // NOTE: bounds the total number of queue entries inspected per purge call (including FINALIZED/DISPUTED entries
    // that are skipped). Every loop iteration counts against this limit, so per-call gas cost is always bounded
    // regardless of how large a FINALIZED prefix has accumulated in front of the first live entry.
    uint32 public constant MAX_DEPOSIT_ESCROW_STEPS = 64;

    // Gas limit for outbound transfers to prevent gas depletion attacks
    // Sufficient for: ETH transfers to smart wallets (6k-9k gas), ERC20 standard transfers (~50k gas),
    // ERC777 hooks (~2.6k registry lookup + <5k hook execution)
    uint256 public constant TRANSFER_GAS_LIMIT = 100000;

    // Delay between validator registration and first use.
    // Gives users a window to observe a newly registered validator on-chain and revoke
    // ERC20 approvals before a compromised NODE key can weaponise it to forge user signatures.
    uint64 public constant VALIDATOR_ACTIVATION_DELAY = 1 days;

    mapping(bytes32 channelId => ChannelMeta meta) internal _channels;
    mapping(address user => EnumerableSet.Bytes32Set channelIds) internal _userChannels;

    mapping(bytes32 escrowId => EscrowDepositMeta meta) internal _escrowDeposits;
    // sorted by `unlockAt` ascending
    bytes32[] internal _escrowDepositIds;
    // points to the first non-purged escrow deposit
    uint256 public escrowHead;

    mapping(bytes32 escrowId => EscrowWithdrawalMeta meta) internal _escrowWithdrawals;

    mapping(address node => mapping(address token => uint256 balance)) internal _nodeBalances;

    // Validator ID 0x00 is reserved for DEFAULT_SIG_VALIDATOR
    // Validator IDs 0x01-0xFF are available for node-registered validators
    mapping(uint8 validatorId => NodeValidator) internal _validatorRegistry;

    // Reclaim balances for failed outbound transfers
    // Accumulates funds when transfers fail (blacklists, hooks, gas depletion)
    // Users can claim these funds later via claimFunds()
    mapping(address account => mapping(address token => uint256 amount)) internal _reclaims;

    // ========== Constructor ==========

    constructor(ISignatureValidator _defaultSigValidator, address _node) {
        require(address(_defaultSigValidator) != address(0), InvalidAddress());
        require(_node != address(0), InvalidAddress());
        DEFAULT_SIG_VALIDATOR = _defaultSigValidator;
        NODE = _node;
    }

    // ========== Getters ==========

    function getAccountBalance(address node, address token) external view returns (uint256) {
        return _nodeBalances[node][token];
    }

    function getNodeValidator(uint8 validatorId)
        external
        view
        returns (ISignatureValidator validator, uint64 registeredAt)
    {
        NodeValidator memory validatorInfo = _validatorRegistry[validatorId];
        return (validatorInfo.validator, validatorInfo.registeredAt);
    }

    function getChannelIds(address user) external view returns (bytes32[] memory) {
        return _userChannels[user].values();
    }

    // Filter only non-closed and non-migrated-out channels
    function getOpenChannels(address user) external view returns (bytes32[] memory openChannels) {
        openChannels = _userChannels[user].values();
        uint256 count = 0;

        // Optimization: single pass filter moves open channels to front, tracks count
        for (uint256 i = 0; i < openChannels.length; i++) {
            ChannelStatus status = _channels[openChannels[i]].status;
            if (status != ChannelStatus.CLOSED && status != ChannelStatus.MIGRATED_OUT) {
                if (count < i) openChannels[count] = openChannels[i];
                count++;
            }
        }

        // Resize array to actual count
        assembly {
            mstore(openChannels, count)
        }
    }

    function getChannelData(bytes32 channelId)
        external
        view
        returns (
            ChannelStatus status,
            ChannelDefinition memory definition,
            State memory lastState,
            uint256 challengeExpiry,
            uint256 lockedFunds
        )
    {
        ChannelMeta memory meta = _channels[channelId];
        status = meta.status;
        definition = meta.definition;
        lastState = meta.lastState;
        challengeExpiry = meta.challengeExpireAt;
        lockedFunds = meta.lockedFunds;
    }

    function getEscrowDepositData(bytes32 escrowId)
        external
        view
        returns (
            bytes32 channelId,
            EscrowStatus status,
            uint64 unlockAt,
            uint64 challengeExpiry,
            uint256 lockedAmount,
            State memory initState
        )
    {
        EscrowDepositMeta memory meta = _escrowDeposits[escrowId];
        channelId = meta.channelId;
        status = meta.status;
        unlockAt = meta.unlockAt;
        challengeExpiry = meta.challengeExpireAt;
        lockedAmount = meta.lockedAmount;
        initState = meta.initState;
    }

    function getEscrowWithdrawalData(bytes32 escrowId)
        external
        view
        returns (
            bytes32 channelId,
            EscrowStatus status,
            uint64 challengeExpiry,
            uint256 lockedAmount,
            State memory initState
        )
    {
        EscrowWithdrawalMeta memory meta = _escrowWithdrawals[escrowId];
        channelId = meta.channelId;
        status = meta.status;
        challengeExpiry = meta.challengeExpireAt;
        lockedAmount = meta.lockedAmount;
        initState = meta.initState;
    }

    function getReclaimBalance(address account, address token) external view returns (uint256) {
        return _reclaims[account][token];
    }

    // ========= IVault ==========

    function depositToVault(address node, address token, uint256 amount) external payable {
        require(node != address(0), InvalidAddress());
        require(amount > 0, IncorrectAmount());

        uint256 updatedBalance = _nodeBalances[node][token] + amount;
        _nodeBalances[node][token] = updatedBalance;

        _pullFunds(msg.sender, token, amount);

        emit Deposited(node, token, amount);
        emit NodeBalanceUpdated(node, token, updatedBalance);
    }

    function withdrawFromVault(address to, address token, uint256 amount) external {
        require(to != address(0), InvalidAddress());
        require(amount > 0, IncorrectAmount());

        address node = msg.sender;
        uint256 currentBalance = _nodeBalances[node][token];
        require(currentBalance >= amount, InsufficientBalance());

        uint256 updatedBalance = currentBalance - amount;
        _nodeBalances[node][token] = updatedBalance;

        _pushFunds(to, token, amount);

        emit Withdrawn(node, token, amount);
        emit NodeBalanceUpdated(node, token, updatedBalance);
    }

    /**
     * @notice Claim accumulated funds from failed outbound transfers
     * @dev Allows users to claim funds that failed to transfer due to blacklists, gas depletion, or other reasons
     * @param token The token address (address(0) for native ETH)
     * @param destination The destination address to send funds to (can differ from msg.sender for blacklisted users)
     */
    function claimFunds(address token, address destination) external nonReentrant {
        require(destination != address(0), InvalidAddress());

        address account = msg.sender;
        uint256 amount = _reclaims[account][token];
        require(amount > 0, IncorrectAmount());

        _reclaims[account][token] = 0;

        // Transfer without gas limit or reclaim logic (user controls gas, accepts responsibility)
        if (token == address(0)) {
            (bool success,) = payable(destination).call{value: amount}("");
            require(success, NativeTransferFailed(destination, amount));
        } else {
            IERC20(token).safeTransfer(destination, amount);
        }

        emit FundsClaimed(account, token, destination, amount);
    }

    // ========= Escrow Deposit Purge ==========
    function getUnlockableEscrowDepositStats() public view returns (uint256 count, uint256 totalAmount) {
        uint256 totalDeposits = _escrowDepositIds.length;
        uint256 escrowHeadTemp = escrowHead;

        while (escrowHeadTemp < totalDeposits) {
            bytes32 escrowId = _escrowDepositIds[escrowHeadTemp];
            EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

            if (_isEscrowDepositSkippable(meta)) {
                escrowHeadTemp++;
                continue;
            }
            if (_isEscrowDepositUnlockable(meta)) {
                count++;
                totalAmount += meta.lockedAmount;
                escrowHeadTemp++;
            } else {
                break;
            }
        }
    }

    function getEscrowDepositIds(uint256 page, uint256 pageSize) external view returns (bytes32[] memory ids) {
        uint256 totalDeposits = _escrowDepositIds.length;
        uint256 start = page * pageSize;
        if (start >= totalDeposits) {
            return new bytes32[](0);
        }
        uint256 end = start + pageSize;
        if (end > totalDeposits) {
            end = totalDeposits;
        }
        ids = new bytes32[](end - start);
        for (uint256 i = start; i < end; i++) {
            ids[i - start] = _escrowDepositIds[i];
        }
    }

    function purgeEscrowDeposits(uint256 maxSteps) external {
        _purgeEscrowDeposits(maxSteps);
    }

    function _purgeEscrowDeposits() internal {
        _purgeEscrowDeposits(MAX_DEPOSIT_ESCROW_STEPS);
    }

    function _purgeEscrowDeposits(uint256 maxSteps) internal {
        uint256 purgedCount = 0;
        uint256 steps = 0;
        uint256 totalDeposits = _escrowDepositIds.length;
        uint256 escrowHeadTemp = escrowHead;

        while (escrowHeadTemp < totalDeposits && steps < maxSteps) {
            bytes32 escrowId = _escrowDepositIds[escrowHeadTemp];
            EscrowDepositMeta storage meta = _escrowDeposits[escrowId];
            steps++;

            if (_isEscrowDepositSkippable(meta)) {
                escrowHeadTemp++;
                continue;
            }
            if (_isEscrowDepositUnlockable(meta)) {
                address token = meta.initState.nonHomeLedger.token;
                uint256 updatedBalance = _nodeBalances[NODE][token] + meta.lockedAmount;
                _nodeBalances[NODE][token] = updatedBalance;

                meta.status = EscrowStatus.FINALIZED;
                meta.lockedAmount = 0;
                purgedCount++;
                escrowHeadTemp++;

                emit NodeBalanceUpdated(NODE, token, updatedBalance);
            } else {
                break;
            }
        }

        escrowHead = escrowHeadTemp;

        if (purgedCount != 0) {
            emit EscrowDepositsPurged(purgedCount);
        }
    }

    /// @dev Check if an escrow deposit can be unlocked (INITIALIZED and past unlock time)
    function _isEscrowDepositUnlockable(EscrowDepositMeta storage meta) internal view returns (bool) {
        return meta.unlockAt <= block.timestamp && meta.status == EscrowStatus.INITIALIZED;
    }

    /// @dev Check if an escrow deposit should be skipped without action (FINALIZED or DISPUTED)
    function _isEscrowDepositSkippable(EscrowDepositMeta storage meta) internal view returns (bool) {
        return meta.status == EscrowStatus.FINALIZED || meta.status == EscrowStatus.DISPUTED;
    }

    // ========= Validator Registry ==========

    /**
     * @notice Register a signature validator for a node using signature-based authorization
     * @dev Anyone can submit this transaction with a valid node signature, enabling relayer-friendly registration.
     *      The node's private key only signs the registration data, never sends transactions directly.
     *      This allows nodes to use cold storage or HSMs without exposing keys to transaction submission.
     *      The signature includes block.chainid and address(this) to prevent cross-chain and cross-deployment replay attacks.
     * @param validatorId The validator ID (0x01-0xFF, 0x00 reserved for DEFAULT)
     * @param validator The validator contract address
     * @param signature NODE's signature over abi.encode(block.chainid, address(this), validatorId, validator)
     */
    function registerNodeValidator(uint8 validatorId, ISignatureValidator validator, bytes calldata signature)
        external
    {
        require(validatorId != DEFAULT_SIG_VALIDATOR_ID, InvalidValidatorId());
        require(address(validator) != address(0), InvalidAddress());
        require(address(_validatorRegistry[validatorId].validator) == address(0), ValidatorAlreadyRegistered(validatorId));

        bytes memory message = Utils.getValidatorRegistrationMessage(address(this), validatorId, address(validator));
        require(EcdsaSignatureUtils.validateEcdsaSigner(message, signature, NODE), IncorrectSignature());

        _validatorRegistry[validatorId] = NodeValidator({validator: validator, registeredAt: uint64(block.timestamp)});

        emit ValidatorRegistered(validatorId, validator);
    }

    // ========== Channel lifecycle ==========

    // Create channel with DEPOSIT, WITHDRAW, or OPERATE intent
    // This enables users who already have off-chain virtual states with non-zero version
    // to create a channel and perform initial operation simultaneously
    function createChannel(ChannelDefinition calldata def, State calldata initState) external payable {
        require(
            initState.intent == StateIntent.DEPOSIT || initState.intent == StateIntent.WITHDRAW
                || initState.intent == StateIntent.OPERATE,
            IncorrectStateIntent()
        );

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        address user = def.user;

        _requireValidDefinition(def);
        _validateSignatures(channelId, initState, user, def.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][initState.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, initState);

        _applyEffects(channelId, def, initState, effects);
        _userChannels[user].add(channelId);

        // Emit appropriate event based on intent
        if (initState.intent == StateIntent.DEPOSIT) {
            emit ChannelDeposited(channelId, initState);
        } else if (initState.intent == StateIntent.WITHDRAW) {
            emit ChannelWithdrawn(channelId, initState);
        } else {
            emit ChannelCheckpointed(channelId, initState);
        }

        emit ChannelCreated(channelId, user, NODE, def, initState);
    }

    function depositToChannel(bytes32 channelId, State calldata candidate) public payable {
        require(candidate.intent == StateIntent.DEPOSIT, IncorrectStateIntent());

        ChannelDefinition memory def = _channels[channelId].definition;
        _validateSignatures(channelId, candidate, def.user, def.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, def, candidate, effects);

        emit ChannelDeposited(channelId, candidate);
    }

    function withdrawFromChannel(bytes32 channelId, State calldata candidate) public payable {
        require(candidate.intent == StateIntent.WITHDRAW, IncorrectStateIntent());

        ChannelDefinition memory def = _channels[channelId].definition;
        _validateSignatures(channelId, candidate, def.user, def.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, def, candidate, effects);

        emit ChannelWithdrawn(channelId, candidate);
    }

    function checkpointChannel(bytes32 channelId, State calldata candidate) external payable {
        require(candidate.intent == StateIntent.OPERATE, IncorrectStateIntent()); // Can only checkpoint operate states

        ChannelDefinition memory def = _channels[channelId].definition;
        _validateSignatures(channelId, candidate, def.user, def.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, def, candidate, effects);

        emit ChannelCheckpointed(channelId, candidate);
    }

    function challengeChannel(
        bytes32 channelId,
        State calldata candidate,
        bytes calldata challengerSig,
        ParticipantIndex challengerIdx
    ) external payable {
        ChannelMeta storage meta = _channels[channelId];
        ChannelDefinition memory def = meta.definition;
        ChannelStatus status = meta.status;

        require(status == ChannelStatus.OPERATING || status == ChannelStatus.MIGRATING_IN, IncorrectChannelStatus());

        State memory prevState = meta.lastState;
        require(candidate.version >= prevState.version, ChallengerVersionTooLow());

        address user = def.user;
        uint256 approvedSignatureValidators = def.approvedSignatureValidators;

        // If version is higher, process the new state
        if (candidate.version > prevState.version) {
            // Cannot challenge with a CLOSE intent, use `closeChannel(...)` function instead
            require(candidate.intent != StateIntent.CLOSE, IncorrectStateIntent());
            // Cannot challenge with a FINALIZE_MIGRATION intent on an old home channel, use `finalizeMigration(...)` function instead
            require(
                !(status == ChannelStatus.OPERATING && candidate.intent == StateIntent.FINALIZE_MIGRATION),
                IncorrectStateIntent()
            );

            _validateSignatures(channelId, candidate, user, approvedSignatureValidators);

            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[NODE][candidate.homeLedger.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyTransitionEffects(channelId, def, candidate, effects);
        }
        // else: challenging with same version, state already processed

        (ISignatureValidator validator, bytes calldata sigData) =
            _extractValidator(challengerSig, approvedSignatureValidators);
        _validateChallengerSignature(channelId, candidate, sigData, validator, user, challengerIdx);

        meta.status = ChannelStatus.DISPUTED;
        uint64 challengeExpiry = uint64(block.timestamp) + def.challengeDuration;
        meta.challengeExpireAt = challengeExpiry;

        emit ChannelChallenged(channelId, candidate, challengeExpiry);
    }

    function closeChannel(bytes32 channelId, State calldata candidate) external payable {
        ChannelMeta storage meta = _channels[channelId];
        ChannelDefinition memory def = meta.definition;
        ChannelStatus status = meta.status;

        State memory prevState = meta.lastState;
        address user = def.user;

        // Path 1: Unilateral closure after challenge timeout
        if (status == ChannelStatus.DISPUTED && meta.challengeExpireAt < block.timestamp) {
            meta.status = ChannelStatus.CLOSED;
            meta.lockedFunds = 0;
            meta.challengeExpireAt = 0;

            _nonRevertingPushFunds(user, prevState.homeLedger.token, prevState.homeLedger.userAllocation);
            _nonRevertingPushFunds(NODE, prevState.homeLedger.token, prevState.homeLedger.nodeAllocation);

            _userChannels[user].remove(channelId);

            emit ChannelClosed(channelId, prevState);
            return;
        }

        // Path 2: Cooperative closure with signed CLOSE state
        require(candidate.intent == StateIntent.CLOSE, IncorrectStateIntent());
        _validateSignatures(channelId, candidate, user, def.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, def, candidate, effects);
        _userChannels[user].remove(channelId);

        emit ChannelClosed(channelId, candidate);
    }

    // ========= Cross-Chain Functions ==========

    function initiateEscrowDeposit(ChannelDefinition calldata def, State calldata candidate) external payable {
        require(candidate.intent == StateIntent.INITIATE_ESCROW_DEPOSIT, IncorrectStateIntent());
        _requireValidDefinition(def);

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        address user = def.user;
        uint256 approvedSignatureValidators = def.approvedSignatureValidators;
        _validateSignatures(channelId, candidate, user, approvedSignatureValidators);

        bytes32 escrowId = Utils.getEscrowId(channelId, candidate.version);

        if (_isChannelHomeChain(channelId)) {
            _processHomeChainEscrowInitiate(channelId, candidate);
            emit EscrowDepositInitiatedOnHome(escrowId, channelId, candidate);
        } else {
            // NON-HOME CHAIN: Create escrow record - recover addresses from signatures
            EscrowDepositEngine.TransitionContext memory ctx = _buildEscrowDepositContext(escrowId);
            EscrowDepositEngine.TransitionEffects memory effects =
                EscrowDepositEngine.validateTransition(ctx, candidate);

            _applyEscrowDepositEffects(
                escrowId, channelId, candidate, effects, user, approvedSignatureValidators
            );
            _escrowDepositIds.push(escrowId);

            emit EscrowDepositInitiated(escrowId, channelId, candidate);
        }
    }

    function challengeEscrowDeposit(bytes32 escrowId, bytes calldata challengerSig, ParticipantIndex challengerIdx)
        external
    {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];
        bytes32 channelId = meta.channelId;
        require(channelId != bytes32(0), NoChannelIdFoundForEscrow());

        address user = meta.user;
        uint256 approvedSignatureValidators = meta.approvedSignatureValidators;

        (ISignatureValidator validator, bytes calldata sigData) =
            _extractValidator(challengerSig, approvedSignatureValidators);
        _validateChallengerSignature(channelId, meta.initState, sigData, validator, user, challengerIdx);

        EscrowDepositEngine.TransitionContext memory ctx = _buildEscrowDepositContext(escrowId);
        EscrowDepositEngine.TransitionEffects memory effects = EscrowDepositEngine.validateChallenge(ctx);

        _applyEscrowDepositEffects(escrowId, channelId, meta.initState, effects, user, approvedSignatureValidators);

        emit EscrowDepositChallenged(escrowId, meta.initState, effects.newChallengeExpiry);
    }

    function finalizeEscrowDeposit(bytes32 channelId, bytes32 escrowId, State calldata candidate) external {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

        if (_isEscrowDepositHomeChain(channelId, escrowId)) {
            // HOME CHAIN: Get user from channel definition
            ChannelMeta storage channelMeta = _channels[channelId];
            _processHomeChainEscrowFinalize(channelId, candidate, channelMeta.definition.user);
            emit EscrowDepositFinalizedOnHome(escrowId, channelId, candidate);
            return;
        }

        // NON-HOME CHAIN: Use escrow metadata
        require(meta.channelId == channelId, IncorrectChannelId()); // Validate consistency
        address user = meta.user;
        EscrowStatus status = meta.status;

        if (status == EscrowStatus.DISPUTED && meta.challengeExpireAt < block.timestamp) {
            // NON-HOME CHAIN: Unilateral finalization after challenge timeout
            meta.status = EscrowStatus.FINALIZED;
            uint256 lockedAmount = meta.lockedAmount;
            meta.lockedAmount = 0;
            meta.challengeExpireAt = 0;

            // Release to user as "deposit exchange" has not been signed yet (it is the "finalizeEscrowDeposit" state)
            _nonRevertingPushFunds(user, meta.initState.nonHomeLedger.token, lockedAmount);

            // Eagerly advance the queue head so FINALIZED entries don't accumulate
            _purgeEscrowDeposits();

            emit EscrowDepositFinalized(escrowId, channelId, candidate);
            return;
        }

        require(candidate.intent == StateIntent.FINALIZE_ESCROW_DEPOSIT, IncorrectStateIntent());

        uint256 approvedSignatureValidators = meta.approvedSignatureValidators;

        // NON-HOME CHAIN: Update via EscrowDepositEngine
        _validateSignatures(channelId, candidate, user, approvedSignatureValidators);

        EscrowDepositEngine.TransitionContext memory ctx = _buildEscrowDepositContext(escrowId);
        EscrowDepositEngine.TransitionEffects memory effects = EscrowDepositEngine.validateTransition(ctx, candidate);

        _applyEscrowDepositEffects(
            escrowId, channelId, candidate, effects, user, approvedSignatureValidators
        );

        emit EscrowDepositFinalized(escrowId, channelId, candidate);
    }

    function initiateEscrowWithdrawal(ChannelDefinition calldata def, State calldata candidate) external {
        require(candidate.intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL, IncorrectStateIntent());
        _requireValidDefinition(def);

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        address user = def.user;
        uint256 approvedSignatureValidators = def.approvedSignatureValidators;
        _validateSignatures(channelId, candidate, user, approvedSignatureValidators);

        bytes32 escrowId = Utils.getEscrowId(channelId, candidate.version);

        if (_isChannelHomeChain(channelId)) {
            // HOME CHAIN: Process through channel state, no escrow metadata
            _processHomeChainEscrowInitiate(channelId, candidate);
            emit EscrowWithdrawalInitiatedOnHome(escrowId, channelId, candidate);
        } else {
            // NON-HOME CHAIN
            EscrowWithdrawalEngine.TransitionContext memory ctx =
                _buildEscrowWithdrawalContext(escrowId, _nodeBalances[NODE][candidate.nonHomeLedger.token]);
            EscrowWithdrawalEngine.TransitionEffects memory effects =
                EscrowWithdrawalEngine.validateTransition(ctx, candidate);

            _applyEscrowWithdrawalEffects(
                escrowId, channelId, candidate, effects, user, approvedSignatureValidators
            );

            emit EscrowWithdrawalInitiated(escrowId, channelId, candidate);
        }
    }

    function challengeEscrowWithdrawal(bytes32 escrowId, bytes calldata challengerSig, ParticipantIndex challengerIdx)
        external
    {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];
        bytes32 channelId = meta.channelId;
        require(channelId != bytes32(0), NoChannelIdFoundForEscrow());

        EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, 0);
        EscrowWithdrawalEngine.TransitionEffects memory effects = EscrowWithdrawalEngine.validateChallenge(ctx);

        // Validate challenger signature
        address user = meta.user;
        uint256 approvedSignatureValidators = meta.approvedSignatureValidators;
        (ISignatureValidator validator, bytes calldata sigData) =
            _extractValidator(challengerSig, approvedSignatureValidators);
        _validateChallengerSignature(channelId, meta.initState, sigData, validator, user, challengerIdx);

        _applyEscrowWithdrawalEffects(
            escrowId, channelId, meta.initState, effects, user, approvedSignatureValidators
        );

        emit EscrowWithdrawalChallenged(escrowId, meta.initState, effects.newChallengeExpiry);
    }

    function finalizeEscrowWithdrawal(bytes32 channelId, bytes32 escrowId, State calldata candidate) external {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];

        if (_isEscrowWithdrawalHomeChain(channelId, escrowId)) {
            // HOME CHAIN: Get user from channel definition
            ChannelMeta storage channelMeta = _channels[channelId];
            _processHomeChainEscrowFinalize(channelId, candidate, channelMeta.definition.user);
            emit EscrowWithdrawalFinalizedOnHome(escrowId, channelId, candidate);
            return;
        }

        // NON-HOME CHAIN: Use escrow metadata
        require(meta.channelId == channelId, IncorrectChannelId()); // Validate consistency
        address user = meta.user;
        EscrowStatus status = meta.status;

        if (status == EscrowStatus.DISPUTED && meta.challengeExpireAt < block.timestamp) {
            // NON-HOME CHAIN: Unilateral finalization after challenge timeout
            meta.status = EscrowStatus.FINALIZED;
            uint256 lockedAmount = meta.lockedAmount;
            meta.lockedAmount = 0;
            meta.challengeExpireAt = 0;

            // Release locked amount back to node as "withdrawal exchange" has not been signed yet (it is the "finalizeEscrowWithdrawal" state)
            address withdrawalToken = meta.initState.nonHomeLedger.token;
            uint256 updatedWithdrawalBalance = _nodeBalances[NODE][withdrawalToken] + lockedAmount;
            _nodeBalances[NODE][withdrawalToken] = updatedWithdrawalBalance;

            emit NodeBalanceUpdated(NODE, withdrawalToken, updatedWithdrawalBalance);

            // Eagerly advance the queue head so FINALIZED entries don't accumulate
            _purgeEscrowDeposits();

            emit EscrowWithdrawalFinalized(escrowId, channelId, candidate);
            return;
        }

        require(candidate.intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL, IncorrectStateIntent());

        // NON-HOME CHAIN: Update via EscrowWithdrawalEngine
        uint256 approvedSignatureValidators = meta.approvedSignatureValidators;
        _validateSignatures(channelId, candidate, user, approvedSignatureValidators);

        EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, 0);
        EscrowWithdrawalEngine.TransitionEffects memory effects =
            EscrowWithdrawalEngine.validateTransition(ctx, candidate);

        _applyEscrowWithdrawalEffects(
            escrowId, channelId, candidate, effects, user, approvedSignatureValidators
        );

        emit EscrowWithdrawalFinalized(escrowId, channelId, candidate);
    }

    function initiateMigration(ChannelDefinition calldata def, State calldata candidate) external {
        require(candidate.intent == StateIntent.INITIATE_MIGRATION, IncorrectStateIntent());

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        address user = def.user;
        _validateSignatures(channelId, candidate, user, def.approvedSignatureValidators);

        State memory targetCandidate = candidate;
        bool isHomeChain = _isChannelHomeChain(channelId);

        if (!isHomeChain) {
            // Initiate migration IN (on new home chain)
            _requireValidDefinition(def);

            // Swap states before processing it, so that homeLedger = current chain
            targetCandidate.homeLedger = candidate.nonHomeLedger;
            targetCandidate.nonHomeLedger = candidate.homeLedger;
            targetCandidate.userSig = ""; // Invalidate signatures after swap
            targetCandidate.nodeSig = "";

            _userChannels[user].add(channelId);
        }

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][targetCandidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, targetCandidate);
        _applyEffects(channelId, def, targetCandidate, effects);

        if (isHomeChain) {
            emit MigrationOutInitiated(channelId, candidate);
        } else {
            emit MigrationInInitiated(channelId, candidate);
        }
    }

    function finalizeMigration(bytes32 channelId, State calldata candidate) external {
        require(candidate.intent == StateIntent.FINALIZE_MIGRATION, IncorrectStateIntent());

        ChannelDefinition memory def = _channels[channelId].definition;
        address user = def.user;

        _validateSignatures(channelId, candidate, user, def.approvedSignatureValidators);

        State memory targetCandidate = candidate;
        // `_isHomeChain(...)` cannot be used here as channel exists on both chains
        bool isOldHomeChain = candidate.nonHomeLedger.chainId == block.chainid;

        if (isOldHomeChain) {
            // Finalize migration OUT (on old home chain)
            // Swap states before validation to maintain invariant, so that homeLedger = current chain
            targetCandidate.homeLedger = candidate.nonHomeLedger;
            targetCandidate.nonHomeLedger = candidate.homeLedger;
            targetCandidate.userSig = ""; // Invalidate signatures after swap
            targetCandidate.nodeSig = "";

            _userChannels[user].remove(channelId);
        }

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][targetCandidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, targetCandidate);
        _applyEffects(channelId, def, targetCandidate, effects);

        if (isOldHomeChain) {
            emit MigrationOutFinalized(channelId, candidate);
        } else {
            emit MigrationInFinalized(channelId, candidate);
        }
    }

    // ========= Internal ==========

    function _validateSignatures(
        bytes32 channelId,
        State calldata state,
        address user,
        uint256 approvedSignatureValidators
    ) internal view {
        (ISignatureValidator userValidator, bytes calldata userSigData) =
            _extractValidator(state.userSig, approvedSignatureValidators);
        _validateSignature(channelId, state, userSigData, user, userValidator);
        (ISignatureValidator nodeValidator, bytes calldata nodeSigData) =
            _extractValidator(state.nodeSig, approvedSignatureValidators);
        _validateSignature(channelId, state, nodeSigData, NODE, nodeValidator);
    }

    /**
     * @notice Validates a single signature (stateless)
     * @param channelId The channel ID for signature validation
     * @param state The state to validate
     * @param sigData The signature data (without validator ID byte)
     * @param participant The expected signer's address
     * @param validator The validator to use for verification
     */
    function _validateSignature(
        bytes32 channelId,
        State calldata state,
        bytes calldata sigData,
        address participant,
        ISignatureValidator validator
    ) internal view {
        bytes memory signingData = Utils.toSigningData(state);
        ValidationResult result = validator.validateSignature(channelId, signingData, sigData, participant);
        require(ValidationResult.unwrap(result) != ValidationResult.unwrap(VALIDATION_FAILURE), IncorrectSignature());
    }

    function _extractValidator(bytes calldata signature, uint256 approvedSignatureValidators)
        internal
        view
        returns (ISignatureValidator validator, bytes calldata sigData)
    {
        require(signature.length > 0, EmptySignature());

        uint8 validatorId = uint8(signature[0]);

        if (validatorId == DEFAULT_SIG_VALIDATOR_ID) {
            // Default validator (0x00) is always available for both users and nodes
            validator = DEFAULT_SIG_VALIDATOR;
        } else {
            // Look up validator in node's registry
            require((approvedSignatureValidators >> validatorId) & 1 == 1, ValidatorNotApproved());

            NodeValidator storage entry = _validatorRegistry[validatorId];
            require(address(entry.validator) != address(0), ValidatorNotRegistered(validatorId));

            uint64 activatesAt = entry.registeredAt + VALIDATOR_ACTIVATION_DELAY;
            require(block.timestamp >= activatesAt, ValidatorNotActive(validatorId, activatesAt));

            validator = entry.validator;
        }

        sigData = _sliceCalldata(signature, 1);
    }

    function _sliceCalldata(bytes calldata data, uint256 start) internal pure returns (bytes calldata result) {
        assembly ("memory-safe") {
            result.offset := add(data.offset, start)
            result.length := sub(data.length, start)
        }
    }

    /**
     * @notice Validates a challenger's signature (stateless)
     * @param channelId The channel ID for signature validation
     * @param state The state being challenged
     * @param sigData The challenger's signature data (without validator ID byte)
     * @param validator The validator to use for verification
     * @param user The user's address
     */
    function _validateChallengerSignature(
        bytes32 channelId,
        State memory state,
        bytes calldata sigData,
        ISignatureValidator validator,
        address user,
        ParticipantIndex challengerIdx
    ) internal view {
        bytes memory signingData = Utils.toSigningData(state);
        bytes memory challengerSigningData = abi.encodePacked(signingData, "challenge");
        address challenger = challengerIdx == ParticipantIndex.USER ? user : NODE;
        ValidationResult result = validator.validateSignature(channelId, challengerSigningData, sigData, challenger);
        require(ValidationResult.unwrap(result) != ValidationResult.unwrap(VALIDATION_FAILURE), IncorrectSignature());
    }

    /// @dev Process HOME CHAIN path for escrow initiate operations
    function _processHomeChainEscrowInitiate(bytes32 channelId, State calldata candidate) internal {
        ChannelDefinition memory metaDef = _channels[channelId].definition;

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, metaDef, candidate, effects);
    }

    /// @dev Process HOME CHAIN path for escrow finalize operations
    function _processHomeChainEscrowFinalize(bytes32 channelId, State calldata candidate, address user)
        internal
    {
        ChannelDefinition memory channelDef = _channels[channelId].definition;
        _validateSignatures(channelId, candidate, user, channelDef.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[NODE][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, channelDef, candidate, effects);
    }

    function _buildChannelContext(bytes32 channelId, uint256 nodeAvailableFunds)
        internal
        view
        returns (ChannelEngine.TransitionContext memory ctx)
    {
        ChannelMeta storage meta = _channels[channelId];

        ctx.status = meta.status;
        ctx.prevState = meta.lastState;
        ctx.lockedFunds = meta.lockedFunds;
        ctx.nodeAvailableFunds = nodeAvailableFunds;
        ctx.challengeExpiry = meta.challengeExpireAt;

        return ctx;
    }

    function _buildEscrowDepositContext(bytes32 escrowId)
        internal
        view
        returns (EscrowDepositEngine.TransitionContext memory ctx)
    {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

        ctx.status = meta.status;
        ctx.initState = meta.initState;
        ctx.lockedAmount = meta.lockedAmount;
        ctx.unlockAt = meta.unlockAt;
        ctx.challengeExpiry = meta.challengeExpireAt;

        return ctx;
    }

    function _buildEscrowWithdrawalContext(bytes32 escrowId, uint256 nodeAvailableFunds)
        internal
        view
        returns (EscrowWithdrawalEngine.TransitionContext memory ctx)
    {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];

        ctx.status = meta.status;
        ctx.initState = meta.initState;
        ctx.lockedAmount = meta.lockedAmount;
        ctx.challengeExpiry = meta.challengeExpireAt;
        ctx.nodeAvailableFunds = nodeAvailableFunds;

        return ctx;
    }

    function _applyEffects(
        bytes32 channelId,
        ChannelDefinition memory def,
        State memory candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        ChannelMeta storage meta = _channels[channelId];

        if (meta.status == ChannelStatus.VOID) {
            meta.definition = def;
        }

        _applyTransitionEffects(channelId, def, candidate, effects);

        if (effects.newStatus != ChannelStatus.VOID && meta.status != effects.newStatus) {
            meta.status = effects.newStatus;
        }

        if (meta.challengeExpireAt != effects.newChallengeExpiry) {
            meta.challengeExpireAt = effects.newChallengeExpiry;
        }

        if (effects.closeChannel) {
            meta.lockedFunds = 0;
        }
    }

    function _applyTransitionEffects(
        bytes32 channelId,
        ChannelDefinition memory def,
        State memory candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        ChannelMeta storage meta = _channels[channelId];

        if (effects.updateLastState) {
            meta.lastState = candidate;
        }

        address user = def.user;
        address token = candidate.homeLedger.token;

        // Process POSITIVE deltas first (additions to lockedFunds) to prevent underflow
        if (effects.userFundsDelta > 0) {
            uint256 amount = effects.userFundsDelta.toUint256();
            _pullFunds(user, token, amount);
            meta.lockedFunds += amount;
        }

        if (effects.nodeFundsDelta > 0) {
            uint256 amount = effects.nodeFundsDelta.toUint256();
            uint256 updatedBalance = _nodeBalances[NODE][token] - amount;
            _nodeBalances[NODE][token] = updatedBalance;
            meta.lockedFunds += amount;

            emit NodeBalanceUpdated(NODE, token, updatedBalance);
        }

        // Then process NEGATIVE deltas (subtractions from lockedFunds)
        if (effects.userFundsDelta < 0) {
            uint256 amount = (-effects.userFundsDelta).toUint256();
            _nonRevertingPushFunds(user, token, amount);
            meta.lockedFunds -= amount;
        }

        if (effects.nodeFundsDelta < 0) {
            uint256 amount = (-effects.nodeFundsDelta).toUint256();
            uint256 updatedBalance = _nodeBalances[NODE][token] + amount;
            _nodeBalances[NODE][token] = updatedBalance;
            meta.lockedFunds -= amount;

            emit NodeBalanceUpdated(NODE, token, updatedBalance);
        }

        // NOTE: purge escrow deposits to unlock unutilized node liquidity
        _purgeEscrowDeposits();
    }

    function _applyEscrowDepositEffects(
        bytes32 escrowId,
        bytes32 channelId,
        State memory candidate,
        EscrowDepositEngine.TransitionEffects memory effects,
        address user,
        uint256 approvedSignatureValidators
    ) internal {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

        if (effects.newStatus != EscrowStatus.VOID) {
            meta.status = effects.newStatus;
        }

        if (effects.updateInitState) {
            _initEscrowDepositMetadata(escrowId, channelId, candidate, user, approvedSignatureValidators);
        }

        if (effects.newUnlockAt > 0) {
            meta.unlockAt = effects.newUnlockAt;
        }

        if (effects.newChallengeExpiry > 0) {
            meta.challengeExpireAt = effects.newChallengeExpiry;
        }

        // Determine the correct token to use (from init state for finalization, from candidate for initiation)
        address token = effects.updateInitState ? candidate.nonHomeLedger.token : meta.initState.nonHomeLedger.token;

        // Handle user funds (positive = pull from user)
        if (effects.userFundsDelta > 0) {
            uint256 amount = effects.userFundsDelta.toUint256();
            _pullFunds(user, token, amount);
            meta.lockedAmount += amount;
        } else if (effects.userFundsDelta < 0) {
            uint256 amount = (-effects.userFundsDelta).toUint256();
            _nonRevertingPushFunds(user, token, amount);
            meta.lockedAmount -= amount;
        }

        // Handle node funds (positive = pull from node vault, negative = release to vault)
        if (effects.nodeFundsDelta > 0) {
            uint256 amount = effects.nodeFundsDelta.toUint256();
            uint256 updatedBalance = _nodeBalances[NODE][token] - amount;
            _nodeBalances[NODE][token] = updatedBalance;
            meta.lockedAmount += amount;
            emit NodeBalanceUpdated(NODE, token, updatedBalance);
        } else if (effects.nodeFundsDelta < 0) {
            uint256 amount = (-effects.nodeFundsDelta).toUint256();
            uint256 updatedBalance = _nodeBalances[NODE][token] + amount;
            _nodeBalances[NODE][token] = updatedBalance;
            meta.lockedAmount -= amount;
            emit NodeBalanceUpdated(NODE, token, updatedBalance);
        }

        // NOTE: purge escrow deposits to unlock unutilized node liquidity
        _purgeEscrowDeposits();
    }

    function _applyEscrowWithdrawalEffects(
        bytes32 escrowId,
        bytes32 channelId,
        State memory candidate,
        EscrowWithdrawalEngine.TransitionEffects memory effects,
        address user,
        uint256 approvedSignatureValidators
    ) internal {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];

        if (effects.newStatus != EscrowStatus.VOID) {
            meta.status = effects.newStatus;
        }

        if (effects.updateInitState) {
            _initEscrowWithdrawalMetadata(escrowId, channelId, candidate, user, approvedSignatureValidators);
        }

        if (effects.newChallengeExpiry > 0) {
            meta.challengeExpireAt = effects.newChallengeExpiry;
        }

        // Determine the correct token to use (from init state for finalization, from candidate for initiation)
        address token = effects.updateInitState ? candidate.nonHomeLedger.token : meta.initState.nonHomeLedger.token;

        // Handle user funds (negative = push to user)
        if (effects.userFundsDelta > 0) {
            uint256 amount = effects.userFundsDelta.toUint256();
            _pullFunds(user, token, amount);
            meta.lockedAmount += amount;
        } else if (effects.userFundsDelta < 0) {
            uint256 amount = (-effects.userFundsDelta).toUint256();
            _nonRevertingPushFunds(user, token, amount);
            meta.lockedAmount -= amount;
        }

        // Handle node funds (positive = pull from node vault, negative = release to vault)
        if (effects.nodeFundsDelta > 0) {
            uint256 amount = effects.nodeFundsDelta.toUint256();
            uint256 updatedBalance = _nodeBalances[NODE][token] - amount;
            _nodeBalances[NODE][token] = updatedBalance;
            meta.lockedAmount += amount;
            emit NodeBalanceUpdated(NODE, token, updatedBalance);
        } else if (effects.nodeFundsDelta < 0) {
            uint256 amount = (-effects.nodeFundsDelta).toUint256();
            uint256 updatedBalance = _nodeBalances[NODE][token] + amount;
            _nodeBalances[NODE][token] = updatedBalance;
            meta.lockedAmount -= amount;
            emit NodeBalanceUpdated(NODE, token, updatedBalance);
        }

        // NOTE: purge escrow deposits to unlock unutilized node liquidity
        _purgeEscrowDeposits();
    }

    function _initEscrowDepositMetadata(
        bytes32 escrowId,
        bytes32 channelId,
        State memory candidate,
        address user,
        uint256 approvedSignatureValidators
    ) internal {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];
        meta.channelId = channelId;
        meta.initState = candidate;
        meta.user = user;
        meta.approvedSignatureValidators = approvedSignatureValidators;
    }

    function _initEscrowWithdrawalMetadata(
        bytes32 escrowId,
        bytes32 channelId,
        State memory candidate,
        address user,
        uint256 approvedSignatureValidators
    ) internal {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];
        meta.channelId = channelId;
        meta.initState = candidate;
        meta.user = user;
        meta.approvedSignatureValidators = approvedSignatureValidators;
    }

    function _requireValidDefinition(ChannelDefinition calldata def) internal view {
        address user = def.user;

        require(user != address(0), InvalidAddress());
        require(def.node == NODE, IncorrectNode());
        require(user != NODE, AddressCollision(user));
        require(def.challengeDuration >= MIN_CHALLENGE_DURATION, IncorrectChallengeDuration());
    }

    /// @dev Returns true when the channel is active and considers the current chain its home chain.
    /// VOID and MIGRATED_OUT are excluded: no channel exists here yet, or the channel has moved away.
    /// MIGRATING_IN is intentionally treated as home chain: initiateMigration() on the new home chain
    /// swaps homeLedger <-> nonHomeLedger before storing, so lastState.homeLedger already represents
    /// the current chain. ChannelEngine processes all subsequent operations (deposit, withdraw,
    /// checkpoint, escrow, close) for MIGRATING_IN using home chain logic — the only thing pending is
    /// a finalizeMigration() call to advance the status to OPERATING.
    function _isChannelHomeChain(bytes32 channelId) internal view returns (bool) {
        ChannelStatus status = _channels[channelId].status;
        if (status == ChannelStatus.VOID || status == ChannelStatus.MIGRATED_OUT) {
            return false;
        }

        return _channels[channelId].lastState.homeLedger.chainId == block.chainid;
    }

    /// @dev Returns true when a finalizeEscrowDeposit call should be routed through the home chain path.
    /// Escrow metadata is only written on the non-home chain during initiateEscrowDeposit, so its
    /// presence means the escrow must always follow the non-home path — even if _isChannelHomeChain()
    /// later returns true because of a subsequent migration.
    function _isEscrowDepositHomeChain(bytes32 channelId, bytes32 escrowId) internal view returns (bool) {
        return _escrowDeposits[escrowId].channelId == bytes32(0) && _isChannelHomeChain(channelId);
    }

    /// @dev Returns true when a finalizeEscrowWithdrawal call should be routed through the home chain path.
    /// Escrow metadata is only written on the non-home chain during initiateEscrowWithdrawal, so its
    /// presence means the escrow must always follow the non-home path — even if _isChannelHomeChain()
    /// later returns true because of a subsequent migration.
    function _isEscrowWithdrawalHomeChain(bytes32 channelId, bytes32 escrowId) internal view returns (bool) {
        return _escrowWithdrawals[escrowId].channelId == bytes32(0) && _isChannelHomeChain(channelId);
    }

    function _pullFunds(address from, address token, uint256 amount) internal nonReentrant {
        if (amount == 0) return;

        if (token == address(0)) {
            require(msg.value == amount, IncorrectValue());
        } else {
            require(msg.value == 0, IncorrectValue());
        }

        if (token != address(0)) {
            IERC20(token).safeTransferFrom(from, address(this), amount);
        }
    }

    /// @dev Reverts if the transfer fails. Used in non-adversarial contexts where atomicity is required
    /// (e.g. voluntary vault withdrawals where the caller controls the destination).
    function _pushFunds(address to, address token, uint256 amount) internal nonReentrant {
        if (amount == 0) return;

        if (token == address(0)) {
            (bool success,) = payable(to).call{value: amount}("");
            require(success, NativeTransferFailed(to, amount));
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    /// @dev Never reverts. On failure, accumulates funds in `_reclaims[to]` for later recovery via `claimFunds()`.
    /// Used in adversarial contexts (e.g. channel settlement) where a reverting recipient must not block progress.
    function _nonRevertingPushFunds(address to, address token, uint256 amount) internal nonReentrant {
        if (amount == 0) return;

        if (token == address(0)) {
            // Native token: limit gas to prevent depletion attacks
            (bool success,) = payable(to).call{value: amount, gas: TRANSFER_GAS_LIMIT}("");
            if (!success) {
                _reclaims[to][token] += amount;
                emit TransferFailed(to, token, amount);
                return;
            }
        } else {
            if (!_trySafeTransfer(token, to, amount)) {
                _reclaims[to][token] += amount;
                emit TransferFailed(to, token, amount);
            }
        }
    }

    /// @dev Gas-capped variant of SafeERC20.trySafeTransfer. Forwards at most TRANSFER_GAS_LIMIT gas
    /// to prevent depletion attacks from malicious ERC777/ERC1363 hooks.
    /// Returns true if the transfer succeeded, false otherwise.
    /// - no return value: success if token is a contract (handles no-return-value tokens like old USDT)
    /// - explicit return value: decoded as uint256, non-zero treated as success.
    ///   Decoding as uint256 (not bool) avoids an abi.decode revert on non-canonical bool encodings
    ///   (e.g. a token returning 2), which would otherwise break _nonRevertingPushFunds' no-revert guarantee.
    function _trySafeTransfer(address token, address to, uint256 amount) internal returns (bool) {
        (bool success, bytes memory returnData) =
            address(token).call{gas: TRANSFER_GAS_LIMIT}(abi.encodeCall(IERC20.transfer, (to, amount)));

        if (!success) return false;
        if (returnData.length == 0) return address(token).code.length > 0;
        // Solidity 0.8's ABI decoder validates canonical bool encoding (only 0 or 1 are valid).
        // A token returning any other non-zero value (e.g. 2, 0xff...ff) would cause abi.decode(..., (bool)) to revert
        // — propagating out of _nonRevertingPushFunds and breaking its invariant
        if (returnData.length >= 32) return abi.decode(returnData, (uint256)) != 0;
        return false;
    }
}

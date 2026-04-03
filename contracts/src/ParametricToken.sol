// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "./interfaces/IParametricToken.sol";
import "forge-std/console.sol";

contract ParametricToken is ERC20, IParametricToken {
    uint8 public constant NUMBER_OF_PARAMETERS = 1;

    struct Account {
        AccountType accountType;
        uint256 balance;
        uint64[NUMBER_OF_PARAMETERS] parameters;
    }

    struct ParamConfig {
        bytes32 name;
        uint8 decimals;
        bool isMutable;
    }

    struct SubAccount {
        uint256 balance;
        uint64[NUMBER_OF_PARAMETERS] parameters;
    }

    struct SuperAccount {
        SubAccount[] subs;
        uint48 subsCount;
    }

    struct Allowance {
        uint256 total;
        uint256 sub;
        uint48 subId;
    }

    ParamConfig[NUMBER_OF_PARAMETERS] public PARAM_CONFIG;
    uint64 constant IMMUTABLE_PARAMETER = 1;

    mapping(address => Account) private _accounts;
    mapping(address => SuperAccount) private _supers;
    mapping(address => mapping(address => Allowance)) private _allowances;

    uint64[NUMBER_OF_PARAMETERS] private _parametersInit;

    modifier onlyNormal(address account) {
        require(_accounts[account].accountType == AccountType.Normal, "Not a normal account");
        _;
    }

    modifier onlySuper(address account) {
        require(_accounts[account].accountType == AccountType.Super, "Not a super account");
        _;
    }

    modifier onlyValidSub(address account, uint48 subId) {
        require(_accounts[account].accountType == AccountType.Super, "Not a super account");
        require(subId < _supers[account].subsCount, "Sub-account doesn't exist");
        _;
    }

    constructor(string memory _name, string memory _symbol) ERC20(_name, _symbol) {
        PARAM_CONFIG = [ParamConfig({name: bytes32("myParam"), decimals: 0, isMutable: true})];
        for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
            _parametersInit[i] = 0;
        }
    }

    // ========== ERC20 Overrides ==========

    function transfer(address to, uint256 amount) public override(ERC20, IERC20) returns (bool) {
        address from = _msgSender();

        Account storage fromAcc = _accounts[from];
        Account storage toAcc = _accounts[to];

        require(_noParamsConflict(from, 0, to, 0), "Conflict of parameters");

        if (fromAcc.accountType == AccountType.Normal && toAcc.accountType == AccountType.Normal) {
            // Normal transfer
            bool success = super.transfer(to, amount);
            if (success) {
                fromAcc.balance -= amount;
                toAcc.balance += amount;
            }
            return success;
        }

        revert("Standard transfer not allowed for super accounts");
    }

    function transferFrom(address from, address to, uint256 amount) public override(ERC20, IERC20) returns (bool) {
        Account storage fromAcc = _accounts[from];
        Account storage toAcc = _accounts[to];

        require(_noParamsConflict(from, 0, to, 0), "Conflict of parameters");

        if (fromAcc.accountType == AccountType.Normal && toAcc.accountType == AccountType.Normal) {
            // Check allowance
            uint256 allowed = _allowances[from][_msgSender()].total;
            require(allowed >= amount, "Insufficient allowance");

            bool success = super.transferFrom(from, to, amount);
            if (success) {
                fromAcc.balance -= amount;
                toAcc.balance += amount;
                _allowances[from][_msgSender()].total -= amount;
            }
            return success;
        }

        revert("TransferFrom not allowed for super accounts");
    }

    function allowance(address owner, address spender) public view override(ERC20, IERC20) returns (uint256) {
        return _allowances[owner][spender].total;
    }

    function approve(address spender, uint256 amount) public override(ERC20, IERC20) returns (bool) {
        address owner = _msgSender();

        _allowances[owner][spender].total = amount;

        if (_accounts[owner].accountType == AccountType.Super && _allowances[owner][spender].sub > amount) {
            _allowances[owner][spender].sub = amount;
        }

        super.approve(spender, amount);

        emit Approval(owner, spender, amount);
        return true;
    }

    // ========== Account Queries ==========

    function name() public view override(ERC20) returns (string memory) {
        return super.name();
    }

    function symbol() public view override(ERC20) returns (string memory) {
        return super.symbol();
    }

    function decimals() public view override(ERC20) returns (uint8) {
        return super.decimals();
    }

    function totalSupply() public view override(ERC20, IERC20) returns (uint256) {
        return super.totalSupply();
    }

    function balanceOf(address account) public view override(ERC20, IERC20) returns (uint256) {
        return _accounts[account].balance;
    }

    // ========== Account Management ==========

    function accountType(address account) external view returns (AccountType) {
        return _accounts[account].accountType;
    }

    function convertToSuper(address account) external onlyNormal(account) returns (bool) {
        require(_msgSender() == account, "Only owner can convert");

        Account storage acc = _accounts[account];

        // Convert to super account
        acc.accountType = AccountType.Super;

        // Create subId 0 with current balance
        _supers[account].subs.push(SubAccount({balance: acc.balance, parameters: acc.parameters}));
        _supers[account].subsCount = 1;

        // Clear normal parameters
        acc.parameters = _parametersInit;

        emit AccountConvertedToSuper(account);
        emit SubAccountCreated(account, 0);

        return true;
    }

    function createSubAccount(address account) external onlySuper(account) returns (uint48) {
        require(_msgSender() == account, "Only owner can create");

        SuperAccount storage acc = _supers[account];

        acc.subs.push(SubAccount({balance: 0, parameters: _parametersInit}));
        acc.subsCount = uint48(acc.subs.length);
        uint48 newSubId = acc.subsCount - 1;

        emit SubAccountCreated(account, newSubId);

        return newSubId;
    }

    // ========== Sub-account Queries ==========

    function balanceOfSub(address account, uint48 subId) external view onlySuper(account) returns (uint256) {
        require(subId < _supers[account].subsCount, "Sub-account doesn't exist");
        return _supers[account].subs[subId].balance;
    }

    function subsCountOf(address account) external view onlySuper(account) returns (uint48) {
        return _supers[account].subsCount;
    }

    function numberOfParameters() external pure returns (uint8) {
        return NUMBER_OF_PARAMETERS;
    }

    function parameterOf(uint8 paramIndex, address account) external view onlyNormal(account) returns (uint64) {
        require(paramIndex < NUMBER_OF_PARAMETERS, "Index exceeds number of parameters");
        return _accounts[account].parameters[paramIndex];
    }

    function parameterOfSub(
        uint8 paramIndex,
        address account,
        uint48 subId
    ) external view onlyValidSub(account, subId) returns (uint64) {
        require(paramIndex < NUMBER_OF_PARAMETERS, "Index exceeds number of parameters");
        return _supers[account].subs[subId].parameters[paramIndex];
    }

    // ========== Allowances ==========

    function allowanceForSub(
        address owner,
        uint48 subId,
        address spender
    ) external view onlyValidSub(owner, subId) returns (uint256) {
        Allowance storage al = _allowances[owner][spender];
        if (al.subId == subId) {
            return al.sub;
        }
        return al.total - al.sub;
    }

    function approveForSub(uint48 ownerSubId, address spender, uint256 amount) external returns (bool) {
        address owner = _msgSender();
        Account storage acc = _accounts[owner];
        require(acc.accountType == AccountType.Super, "Not a super account");
        require(ownerSubId < _supers[owner].subsCount, "Sub-account doesn't exist");

        Allowance storage al = _allowances[owner][spender];

        al.subId = ownerSubId;
        al.sub = amount;

        // Adjust total if needed
        if (amount > al.total) {
            al.total = amount; // Total becomes at least the sub-amount
        }

        emit ApprovalForSub(owner, ownerSubId, spender, amount);

        return true;
    }

    // Helper to check and consume allowance for a specific subId
    function _sufficientAllowanceForSub(
        address owner,
        address spender,
        uint48 fromSubId,
        uint256 amount
    ) internal view onlyValidSub(owner, fromSubId) returns (bool) {
        Allowance storage al = _allowances[owner][spender];
        return fromSubId == al.subId ? al.sub >= amount : al.total - al.sub >= amount;
    }

    function _consumeAllowanceForSub(
        address owner,
        address spender,
        uint48 fromSubId,
        uint256 amount
    ) internal onlyValidSub(owner, fromSubId) {
        Allowance storage al = _allowances[owner][spender];

        if (fromSubId == al.subId) {
            // Use sub-allowance first
            require(al.sub >= amount, "Insufficient sub-allowance");
            al.sub -= amount;
            al.total -= amount;
        } else {
            // Use from remaining total allowance (total - sub)
            uint256 remaining = al.total - al.sub;
            require(remaining >= amount, "Insufficient allowance for this subId");
            al.total -= amount;
        }
    }

    function _noParamsConflict(address from, uint48 fromSubId, address to, uint48 toSubId) private view returns (bool) {
        uint64[NUMBER_OF_PARAMETERS] memory fromParams;
        uint64[NUMBER_OF_PARAMETERS] memory toParams;

        if (_accounts[from].balance == 0 || _accounts[to].balance == 0) return true;

        if (_accounts[from].accountType == AccountType.Normal) {
            fromParams = _accounts[from].parameters;
        } else {
            require(fromSubId < _supers[from].subsCount, "Subaccount doesn't exist");
            fromParams = _supers[from].subs[fromSubId].parameters;
        }

        if (_accounts[to].accountType == AccountType.Normal) {
            toParams = _accounts[to].parameters;
        } else {
            require(toSubId < _supers[to].subsCount, "Subaccount doesn't exist");
            toParams = _supers[to].subs[toSubId].parameters;
        }

        for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
            if (!PARAM_CONFIG[i].isMutable && fromParams[i] != toParams[i]) return false;
        }

        return true;
    }

    function _weightedAverage(
        uint64 param1,
        uint256 amount1,
        uint64 param2,
        uint256 amount2
    ) private pure returns (uint64) {
        require(amount1 > 0 && amount2 > 0, "Invalid amounts");
        uint256 sumProduct = uint256(param1) * amount1 + uint256(param2) * amount2;
        uint256 sum = amount1 + amount2;

        return uint64(sumProduct / sum);
    }

    // ========== Transfers ==========

    function transferToSub(
        address toSuper,
        uint48 toSubId,
        uint256 amount
    ) external onlyValidSub(toSuper, toSubId) returns (bool) {
        address from = _msgSender();
        Account storage fromAcc = _accounts[from];
        Account storage toAcc = _accounts[toSuper];

        require(amount > 0, "Void amount");
        require(fromAcc.accountType == AccountType.Normal, "Sender must be normal");

        require(_noParamsConflict(from, 0, toSuper, toSubId), "Conflict of parameters");
        require(fromAcc.balance >= amount, "Insufficient balance");
        fromAcc.balance -= amount;

        SubAccount storage toSubAcc = _supers[toSuper].subs[toSubId];
        uint256 oldSubBalance = toSubAcc.balance;
        toSubAcc.balance += amount;
        toAcc.balance += amount;

        // Update toSubAcc parameters
        if (oldSubBalance > 0) {
            for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
                if (PARAM_CONFIG[i].isMutable) {
                    toSubAcc.parameters[i] = _weightedAverage(
                        toSubAcc.parameters[i],
                        oldSubBalance,
                        fromAcc.parameters[i],
                        amount
                    );
                }
            }
        } else {
            toSubAcc.parameters = fromAcc.parameters;
        }

        // Update fromAcc parameters
        if (fromAcc.balance == 0) fromAcc.parameters = _parametersInit;

        emit TransferToSub(from, toSuper, toSubId, amount);
        emit Transfer(from, toSuper, amount);

        return true;
    }

    function transferFromSub(uint48 fromSubId, address to, uint256 amount) external returns (bool) {
        address fromSuper = _msgSender();
        Account storage fromAcc = _accounts[fromSuper];
        Account storage toAcc = _accounts[to];

        require(amount > 0, "Void amount");
        require(fromAcc.accountType == AccountType.Super, "Not a super account");
        SuperAccount storage fromSuperAcc = _supers[fromSuper];
        require(toAcc.accountType == AccountType.Normal, "Recipient must be normal");
        require(fromSubId < fromSuperAcc.subsCount, "Sub-account doesn't exist");

        require(_noParamsConflict(fromSuper, fromSubId, to, 0), "Conflict of parameters");
        require(fromSuperAcc.subs[fromSubId].balance >= amount, "Insufficient balance");
        fromSuperAcc.subs[fromSubId].balance -= amount;
        fromAcc.balance -= amount;

        uint256 oldToBalance = toAcc.balance;
        toAcc.balance += amount;

        // Update toAcc parameters
        if (oldToBalance > 0) {
            for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
                if (PARAM_CONFIG[i].isMutable) {
                    toAcc.parameters[i] = _weightedAverage(
                        toAcc.parameters[i],
                        oldToBalance,
                        fromSuperAcc.subs[fromSubId].parameters[i],
                        amount
                    );
                }
            }
        } else {
            toAcc.parameters = fromSuperAcc.subs[fromSubId].parameters;
        }

        // Update fromAcc parameters
        if (fromAcc.balance == 0) fromSuperAcc.subs[fromSubId].parameters = _parametersInit;

        emit TransferFromSub(fromSuper, fromSubId, to, amount);
        emit Transfer(fromSuper, to, amount);

        return true;
    }

    function transferBetweenSubs(uint48 fromSubId, uint48 toSubId, uint256 amount) external returns (bool) {
        address superAccount = _msgSender();
        Account storage acc = _accounts[superAccount];

        require(acc.accountType == AccountType.Super, "Not a super account");
        SuperAccount storage superAcc = _supers[superAccount];

        require(fromSubId < superAcc.subsCount && toSubId < superAcc.subsCount, "Sub-account doesn't exist");
        require(_noParamsConflict(superAccount, fromSubId, superAccount, toSubId), "Conflict of parameters");

        require(superAcc.subs[fromSubId].balance >= amount, "Insufficient balance");

        // Update fromSub
        superAcc.subs[fromSubId].balance -= amount;

        // Update toSub
        uint256 oldSubBalance = superAcc.subs[toSubId].balance;
        superAcc.subs[toSubId].balance += amount;

        // Update toSubId parameters
        if (oldSubBalance > 0) {
            for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
                if (PARAM_CONFIG[i].isMutable) {
                    superAcc.subs[toSubId].parameters[i] = _weightedAverage(
                        superAcc.subs[toSubId].parameters[i],
                        oldSubBalance,
                        superAcc.subs[fromSubId].parameters[i],
                        amount
                    );
                }
            }
        } else {
            superAcc.subs[toSubId].parameters = superAcc.subs[fromSubId].parameters;
        }

        // Update fromSubId parameters
        if (superAcc.subs[fromSubId].balance == 0) superAcc.subs[fromSubId].parameters = _parametersInit;

        emit TransferBetweenSubs(superAccount, fromSubId, toSubId, amount);

        return true;
    }

    // ========== Approved Transfers ==========

    function approvedTransferToSub(
        address from,
        address toSuper,
        uint48 toSubId,
        uint256 amount
    ) external onlyValidSub(toSuper, toSubId) returns (bool) {
        address spender = _msgSender();

        // Execute transfer
        Account storage fromAcc = _accounts[from];
        Allowance storage al = _allowances[from][spender];

        require(amount > 0, "Void amount");
        require(fromAcc.accountType == AccountType.Normal, "From must be normal");
        require(_noParamsConflict(from, 0, toSuper, toSubId), "Conflict of parameters");
        require(fromAcc.balance >= amount, "Insufficient balance");
        require(al.total >= amount, "Insufficient allowance");
        fromAcc.balance -= amount;

        SubAccount storage toSubAcc = _supers[toSuper].subs[toSubId];
        uint256 oldSubBalance = toSubAcc.balance;
        toSubAcc.balance += amount;
        _accounts[toSuper].balance += amount;

        // Update toSubAcc parameters
        if (oldSubBalance > 0) {
            for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
                if (PARAM_CONFIG[i].isMutable) {
                    toSubAcc.parameters[i] = _weightedAverage(
                        toSubAcc.parameters[i],
                        oldSubBalance,
                        fromAcc.parameters[i],
                        amount
                    );
                }
            }
        } else {
            toSubAcc.parameters = fromAcc.parameters;
        }

        // Update fromAcc parameters
        if (fromAcc.balance == 0) fromAcc.parameters = _parametersInit;

        // Consume allowance
        al.total -= amount;

        emit TransferToSub(from, toSuper, toSubId, amount);
        emit Transfer(from, toSuper, amount);

        return true;
    }

    function approvedTransferFromSubToSub(
        address fromSuper,
        uint48 fromSubId,
        address toSuper,
        uint48 toSubId,
        uint256 amount
    ) external onlyValidSub(fromSuper, fromSubId) onlyValidSub(toSuper, toSubId) returns (bool) {
        address spender = _msgSender();

        // Execute transfer from sub to sub
        SuperAccount storage fromSuperAcc = _supers[fromSuper];

        require(amount > 0, "Void amount");
        require(fromSuperAcc.subs[fromSubId].balance >= amount, "Insufficient balance");
        require(_sufficientAllowanceForSub(fromSuper, spender, fromSubId, amount), "Insufficient allowance");

        fromSuperAcc.subs[fromSubId].balance -= amount;
        _accounts[fromSuper].balance -= amount;

        SubAccount storage toSubAcc = _supers[toSuper].subs[toSubId];
        uint256 oldSubBalance = toSubAcc.balance;
        toSubAcc.balance += amount;
        _accounts[toSuper].balance += amount;

        // Update toSubAcc parameters
        if (oldSubBalance > 0) {
            for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
                if (PARAM_CONFIG[i].isMutable) {
                    toSubAcc.parameters[i] = _weightedAverage(
                        toSubAcc.parameters[i],
                        oldSubBalance,
                        fromSuperAcc.subs[fromSubId].parameters[i],
                        amount
                    );
                }
            }
        } else {
            toSubAcc.parameters = fromSuperAcc.subs[fromSubId].parameters;
        }

        // Check and consume allowance
        _consumeAllowanceForSub(fromSuper, spender, fromSubId, amount);

        emit TransferFromSubToSub(fromSuper, fromSubId, toSuper, toSubId, amount);
        if (fromSuper != toSuper) emit Transfer(fromSuper, toSuper, amount);

        return true;
    }

    // ========== Mint/Burn Helpers ==========

    function mint(uint256 amount) external {
        require(amount > 0, "Void amount");
        address to = _msgSender();
        _mintParametric(to, amount);
    }

    // function _mintParametric(address account, uint256 amount) internal {
    //     super._mint(account, amount);
    //     _accounts[account].balance += amount;
    //     // parameter logic...
    // }

    function _mintParametric(address account, uint256 amount) internal {
        console.log(">> _mintParametric called");
        console.log(">> account:", account);

        Account storage acc = _accounts[account];

        if (acc.accountType == AccountType.Super) {
            // Mint to sub-account 0
            SuperAccount storage superAcc = _supers[account];
            require(superAcc.subsCount > 0, "No sub-accounts");

            SubAccount storage sub0 = superAcc.subs[0];

            // Calculate new weighted average parameters
            uint256 oldBalance = sub0.balance;

            for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
                if (PARAM_CONFIG[i].isMutable) {
                    if (oldBalance == 0) {
                        // First mint to this sub - set to block.timestamp
                        sub0.parameters[i] = uint64(block.timestamp);
                    } else {
                        // Weighted average for non-zero
                        sub0.parameters[i] = _weightedAverage(
                            sub0.parameters[i],
                            oldBalance,
                            uint64(block.timestamp),
                            amount
                        );
                    }
                } else {
                    // Immutable parameter
                    if (oldBalance == 0) {
                        // First mint - set to constant
                        sub0.parameters[i] = IMMUTABLE_PARAMETER;
                    }
                    // If oldBalance > 0, immutable parameter stays as is (no change)
                }
            }

            // Update balances
            sub0.balance += amount;
            acc.balance += amount;

            super._mint(account, amount);
        } else {
            // Normal account
            uint256 oldBalance = acc.balance;

            for (uint256 i = 0; i < NUMBER_OF_PARAMETERS; i++) {
                if (PARAM_CONFIG[i].isMutable) {
                    if (oldBalance == 0) {
                        // First mint - set to block.timestamp
                        acc.parameters[i] = uint64(block.timestamp);
                    } else {
                        // Non-zero balance
                        acc.parameters[i] = _weightedAverage(
                            acc.parameters[i],
                            oldBalance,
                            uint64(block.timestamp),
                            amount
                        );
                    }
                } else {
                    // Immutable parameter
                    if (oldBalance == 0) {
                        // First mint - set to constant
                        acc.parameters[i] = IMMUTABLE_PARAMETER;
                    }
                    // If oldBalance > 0, immutable parameter stays as is
                }
            }

            acc.balance += amount;
            super._mint(account, amount);
        }
    }

    // function _burn(address account, uint256 amount) internal {
    //     super._burn(account, amount);
    //     // Your custom logic here
    //     _accounts[account].balance -= amount;
    // }

    function _burnParametric(address account, uint256 amount) internal {
        Account storage acc = _accounts[account];
        require(amount > 0, "Void amount");
        require(account == _msgSender(), "Burn allowed only from own account");

        if (acc.accountType == AccountType.Super) {
            // Burn from sub-account 0
            SuperAccount storage superAcc = _supers[account];
            require(superAcc.subsCount > 0, "No sub-accounts");

            SubAccount storage sub0 = superAcc.subs[0];
            require(sub0.balance >= amount, "Insufficient balance in sub-account 0");

            // Calculate new balance after burn
            uint256 newBalance = sub0.balance - amount;

            // Update parameters if balance becomes zero
            if (newBalance == 0) sub0.parameters = _parametersInit;

            // Note: When balance > 0 after burn, parameters remain unchanged
            // because burning doesn't introduce new tokens with different parameters

            // Update balances
            sub0.balance = newBalance;
            acc.balance -= amount;

            super._burn(account, amount);
        } else {
            // Normal account
            require(acc.balance >= amount, "Insufficient balance");

            uint256 newBalance = acc.balance - amount;

            // Update parameters if balance becomes zero
            if (newBalance == 0) acc.parameters = _parametersInit;

            // Note: When balance > 0 after burn, parameters remain unchanged

            acc.balance = newBalance;
            super._burn(account, amount);
        }
    }
}

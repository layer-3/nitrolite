/**
 * Parametric token contract ABI
 * Parametric token interface
 */

export const ParametricTokenAbi = [
  {
    type: 'function',
    name: 'convertToSuper',
    inputs: [{ name: 'account', type: 'address' }],
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'createSubAccount',
    inputs: [{ name: 'account', type: 'address' }],
    outputs: [{ name: '', type: 'uint48' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'accountType',
    inputs: [{ name: 'account', type: 'address' }],
    outputs: [{ name: '', type: 'uint8' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'balanceOfSub',
    inputs: [
      { name: 'superAccount', type: 'address' },
      { name: 'subId', type: 'uint48' },
    ],
    outputs: [{ name: '', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'subsCountOf',
    inputs: [{ name: 'superAccount', type: 'address' }],
    outputs: [{ name: '', type: 'uint48' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'getSubParameter',
    inputs: [
      { name: 'superAccount', type: 'address' },
      { name: 'subId', type: 'uint48' },
    ],
    outputs: [{ name: '', type: 'bytes32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'allowanceForSub',
    inputs: [
      { name: 'owner', type: 'address' },
      { name: 'subId', type: 'uint48' },
      { name: 'spender', type: 'address' },
    ],
    outputs: [{ name: '', type: 'uint256' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'transferToSub',
    inputs: [
      { name: 'toSuper', type: 'address' },
      { name: 'toSubId', type: 'uint48' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'transferFromSub',
    inputs: [
      { name: 'fromSubId', type: 'uint48' },
      { name: 'to', type: 'address' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'transferBetweenSubs',
    inputs: [
      { name: 'fromSubId', type: 'uint48' },
      { name: 'toSubId', type: 'uint48' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'approveForSub',
    inputs: [
      { name: 'ownerSubId', type: 'uint48' },
      { name: 'spender', type: 'address' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'approvedTransferToSub',
    inputs: [
      { name: 'from', type: 'address' },
      { name: 'toSuper', type: 'address' },
      { name: 'toSubId', type: 'uint48' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'approvedTransferFromSubToSub',
    inputs: [
      { name: 'fromSuper', type: 'address' },
      { name: 'fromSubId', type: 'uint48' },
      { name: 'toSuper', type: 'address' },
      { name: 'toSubId', type: 'uint48' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [{ name: '', type: 'bool' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'event',
    name: 'AccountConvertedToSuper',
    inputs: [{ name: 'account', type: 'address', indexed: true }],
  },
  {
    type: 'event',
    name: 'SubAccountCreated',
    inputs: [
      { name: 'superAccount', type: 'address', indexed: true },
      { name: 'subId', type: 'uint48', indexed: true },
    ],
  },
  {
    type: 'event',
    name: 'TransferToSub',
    inputs: [
      { name: 'from', type: 'address', indexed: true },
      { name: 'toSuper', type: 'address', indexed: true },
      { name: 'toSubId', type: 'uint48', indexed: true },
      { name: 'amount', type: 'uint256', indexed: false },
    ],
  },
  {
    type: 'event',
    name: 'TransferFromSub',
    inputs: [
      { name: 'fromSuper', type: 'address', indexed: true },
      { name: 'fromSubId', type: 'uint48', indexed: true },
      { name: 'to', type: 'address', indexed: true },
      { name: 'amount', type: 'uint256', indexed: false },
    ],
  },
  {
    type: 'event',
    name: 'TransferBetweenSubs',
    inputs: [
      { name: 'superAccount', type: 'address', indexed: true },
      { name: 'fromSubId', type: 'uint48', indexed: true },
      { name: 'toSubId', type: 'uint48', indexed: true },
      { name: 'amount', type: 'uint256', indexed: false },
    ],
  },
  {
    type: 'event',
    name: 'ApprovalForSub',
    inputs: [
      { name: 'owner', type: 'address', indexed: true },
      { name: 'subId', type: 'uint48', indexed: true },
      { name: 'spender', type: 'address', indexed: true },
      { name: 'amount', type: 'uint256', indexed: false },
    ],
  },
] as const;

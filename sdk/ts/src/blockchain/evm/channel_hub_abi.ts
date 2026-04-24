/**
 * ChannelHub contract ABI
 * Generated from contracts/src/ChannelHub.sol
 */

import { Abi } from 'abitype';

export const ChannelHubAbi = [
  {
    type: 'constructor',
    inputs: [
      {
        name: '_defaultSigValidator',
        type: 'address',
        internalType: 'contract ISignatureValidator'
      },
      {
        name: '_node',
        type: 'address',
        internalType: 'address'
      }
    ],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'DEFAULT_SIG_VALIDATOR',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'address',
        internalType: 'contract ISignatureValidator'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'ESCROW_DEPOSIT_UNLOCK_DELAY',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'uint32',
        internalType: 'uint32'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'MAX_DEPOSIT_ESCROW_STEPS',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'uint32',
        internalType: 'uint32'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'MIN_CHALLENGE_DURATION',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'uint32',
        internalType: 'uint32'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'NODE',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'address',
        internalType: 'address'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'TRANSFER_GAS_LIMIT',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'VALIDATOR_ACTIVATION_DELAY',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'uint64',
        internalType: 'uint64'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'VERSION',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'uint8',
        internalType: 'uint8'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'challengeChannel',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      },
      {
        name: 'challengerSig',
        type: 'bytes',
        internalType: 'bytes'
      },
      {
        name: 'challengerIdx',
        type: 'uint8',
        internalType: 'enum ParticipantIndex'
      }
    ],
    outputs: [],
    stateMutability: 'payable'
  },
  {
    type: 'function',
    name: 'challengeEscrowDeposit',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'challengerSig',
        type: 'bytes',
        internalType: 'bytes'
      },
      {
        name: 'challengerIdx',
        type: 'uint8',
        internalType: 'enum ParticipantIndex'
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'challengeEscrowWithdrawal',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'challengerSig',
        type: 'bytes',
        internalType: 'bytes'
      },
      {
        name: 'challengerIdx',
        type: 'uint8',
        internalType: 'enum ParticipantIndex'
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'checkpointChannel',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'claimFunds',
    inputs: [
      {
        name: 'token',
        type: 'address',
        internalType: 'address'
      },
      {
        name: 'destination',
        type: 'address',
        internalType: 'address'
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'closeChannel',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'createChannel',
    inputs: [
      {
        name: 'def',
        type: 'tuple',
        internalType: 'struct ChannelDefinition',
        components: [
          {
            name: 'challengeDuration',
            type: 'uint32',
            internalType: 'uint32'
          },
          {
            name: 'user',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'node',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'nonce',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'approvedSignatureValidators',
            type: 'uint256',
            internalType: 'uint256'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          }
        ]
      },
      {
        name: 'initState',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'payable'
  },
  {
    type: 'function',
    name: 'depositToChannel',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'payable'
  },
  {
    type: 'function',
    name: 'depositToNode',
    inputs: [
      {
        name: 'token',
        type: 'address',
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    outputs: [],
    stateMutability: 'payable'
  },
  {
    type: 'function',
    name: 'escrowHead',
    inputs: [],
    outputs: [
      {
        name: '',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'finalizeEscrowDeposit',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'escrowId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'finalizeEscrowWithdrawal',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'escrowId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'finalizeMigration',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'getChannelData',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      }
    ],
    outputs: [
      {
        name: 'status',
        type: 'uint8',
        internalType: 'enum ChannelStatus'
      },
      {
        name: 'definition',
        type: 'tuple',
        internalType: 'struct ChannelDefinition',
        components: [
          {
            name: 'challengeDuration',
            type: 'uint32',
            internalType: 'uint32'
          },
          {
            name: 'user',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'node',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'nonce',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'approvedSignatureValidators',
            type: 'uint256',
            internalType: 'uint256'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          }
        ]
      },
      {
        name: 'lastState',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      },
      {
        name: 'challengeExpiry',
        type: 'uint256',
        internalType: 'uint256'
      },
      {
        name: 'lockedFunds',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getChannelIds',
    inputs: [
      {
        name: 'user',
        type: 'address',
        internalType: 'address'
      }
    ],
    outputs: [
      {
        name: '',
        type: 'bytes32[]',
        internalType: 'bytes32[]'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getEscrowDepositData',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        internalType: 'bytes32'
      }
    ],
    outputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'status',
        type: 'uint8',
        internalType: 'enum EscrowStatus'
      },
      {
        name: 'unlockAt',
        type: 'uint64',
        internalType: 'uint64'
      },
      {
        name: 'challengeExpiry',
        type: 'uint64',
        internalType: 'uint64'
      },
      {
        name: 'lockedAmount',
        type: 'uint256',
        internalType: 'uint256'
      },
      {
        name: 'initState',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getEscrowDepositIds',
    inputs: [
      {
        name: 'page',
        type: 'uint256',
        internalType: 'uint256'
      },
      {
        name: 'pageSize',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    outputs: [
      {
        name: 'ids',
        type: 'bytes32[]',
        internalType: 'bytes32[]'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getEscrowWithdrawalData',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        internalType: 'bytes32'
      }
    ],
    outputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'status',
        type: 'uint8',
        internalType: 'enum EscrowStatus'
      },
      {
        name: 'challengeExpiry',
        type: 'uint64',
        internalType: 'uint64'
      },
      {
        name: 'lockedAmount',
        type: 'uint256',
        internalType: 'uint256'
      },
      {
        name: 'initState',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getNodeBalance',
    inputs: [
      {
        name: 'token',
        type: 'address',
        internalType: 'address'
      }
    ],
    outputs: [
      {
        name: '',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getNodeValidator',
    inputs: [
      {
        name: 'validatorId',
        type: 'uint8',
        internalType: 'uint8'
      }
    ],
    outputs: [
      {
        name: 'validator',
        type: 'address',
        internalType: 'contract ISignatureValidator'
      },
      {
        name: 'registeredAt',
        type: 'uint64',
        internalType: 'uint64'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getOpenChannels',
    inputs: [
      {
        name: 'user',
        type: 'address',
        internalType: 'address'
      }
    ],
    outputs: [
      {
        name: 'openChannels',
        type: 'bytes32[]',
        internalType: 'bytes32[]'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getReclaimBalance',
    inputs: [
      {
        name: 'account',
        type: 'address',
        internalType: 'address'
      },
      {
        name: 'token',
        type: 'address',
        internalType: 'address'
      }
    ],
    outputs: [
      {
        name: '',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'getUnlockableEscrowDepositStats',
    inputs: [],
    outputs: [
      {
        name: 'count',
        type: 'uint256',
        internalType: 'uint256'
      },
      {
        name: 'totalAmount',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    stateMutability: 'view'
  },
  {
    type: 'function',
    name: 'initiateEscrowDeposit',
    inputs: [
      {
        name: 'def',
        type: 'tuple',
        internalType: 'struct ChannelDefinition',
        components: [
          {
            name: 'challengeDuration',
            type: 'uint32',
            internalType: 'uint32'
          },
          {
            name: 'user',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'node',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'nonce',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'approvedSignatureValidators',
            type: 'uint256',
            internalType: 'uint256'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          }
        ]
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'payable'
  },
  {
    type: 'function',
    name: 'initiateEscrowWithdrawal',
    inputs: [
      {
        name: 'def',
        type: 'tuple',
        internalType: 'struct ChannelDefinition',
        components: [
          {
            name: 'challengeDuration',
            type: 'uint32',
            internalType: 'uint32'
          },
          {
            name: 'user',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'node',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'nonce',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'approvedSignatureValidators',
            type: 'uint256',
            internalType: 'uint256'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          }
        ]
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'initiateMigration',
    inputs: [
      {
        name: 'def',
        type: 'tuple',
        internalType: 'struct ChannelDefinition',
        components: [
          {
            name: 'challengeDuration',
            type: 'uint32',
            internalType: 'uint32'
          },
          {
            name: 'user',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'node',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'nonce',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'approvedSignatureValidators',
            type: 'uint256',
            internalType: 'uint256'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          }
        ]
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'purgeEscrowDeposits',
    inputs: [
      {
        name: 'maxSteps',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'registerNodeValidator',
    inputs: [
      {
        name: 'validatorId',
        type: 'uint8',
        internalType: 'uint8'
      },
      {
        name: 'validator',
        type: 'address',
        internalType: 'contract ISignatureValidator'
      },
      {
        name: 'signature',
        type: 'bytes',
        internalType: 'bytes'
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'withdrawFromChannel',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'withdrawFromNode',
    inputs: [
      {
        name: 'to',
        type: 'address',
        internalType: 'address'
      },
      {
        name: 'token',
        type: 'address',
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        internalType: 'uint256'
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'event',
    name: 'ChannelChallenged',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      },
      {
        name: 'challengeExpireAt',
        type: 'uint64',
        indexed: false,
        internalType: 'uint64'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelCheckpointed',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelClosed',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'finalState',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelCreated',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'user',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'definition',
        type: 'tuple',
        indexed: false,
        internalType: 'struct ChannelDefinition',
        components: [
          {
            name: 'challengeDuration',
            type: 'uint32',
            internalType: 'uint32'
          },
          {
            name: 'user',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'node',
            type: 'address',
            internalType: 'address'
          },
          {
            name: 'nonce',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'approvedSignatureValidators',
            type: 'uint256',
            internalType: 'uint256'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          }
        ]
      },
      {
        name: 'initialState',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelDeposited',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelWithdrawn',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'candidate',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'Deposited',
    inputs: [
      {
        name: 'token',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        indexed: false,
        internalType: 'uint256'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowDepositChallenged',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      },
      {
        name: 'challengeExpireAt',
        type: 'uint64',
        indexed: false,
        internalType: 'uint64'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowDepositFinalized',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowDepositFinalizedOnHome',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowDepositInitiated',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowDepositInitiatedOnHome',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowDepositsPurged',
    inputs: [
      {
        name: 'purgedCount',
        type: 'uint256',
        indexed: false,
        internalType: 'uint256'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowWithdrawalChallenged',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      },
      {
        name: 'challengeExpireAt',
        type: 'uint64',
        indexed: false,
        internalType: 'uint64'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowWithdrawalFinalized',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowWithdrawalFinalizedOnHome',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowWithdrawalInitiated',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'EscrowWithdrawalInitiatedOnHome',
    inputs: [
      {
        name: 'escrowId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'FundsClaimed',
    inputs: [
      {
        name: 'account',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'token',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'destination',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        indexed: false,
        internalType: 'uint256'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'MigrationInFinalized',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'MigrationInInitiated',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'MigrationOutFinalized',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'MigrationOutInitiated',
    inputs: [
      {
        name: 'channelId',
        type: 'bytes32',
        indexed: true,
        internalType: 'bytes32'
      },
      {
        name: 'state',
        type: 'tuple',
        indexed: false,
        internalType: 'struct State',
        components: [
          {
            name: 'version',
            type: 'uint64',
            internalType: 'uint64'
          },
          {
            name: 'intent',
            type: 'uint8',
            internalType: 'enum StateIntent'
          },
          {
            name: 'metadata',
            type: 'bytes32',
            internalType: 'bytes32'
          },
          {
            name: 'homeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'nonHomeLedger',
            type: 'tuple',
            internalType: 'struct Ledger',
            components: [
              {
                name: 'chainId',
                type: 'uint64',
                internalType: 'uint64'
              },
              {
                name: 'token',
                type: 'address',
                internalType: 'address'
              },
              {
                name: 'decimals',
                type: 'uint8',
                internalType: 'uint8'
              },
              {
                name: 'userAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'userNetFlow',
                type: 'int256',
                internalType: 'int256'
              },
              {
                name: 'nodeAllocation',
                type: 'uint256',
                internalType: 'uint256'
              },
              {
                name: 'nodeNetFlow',
                type: 'int256',
                internalType: 'int256'
              }
            ]
          },
          {
            name: 'userSig',
            type: 'bytes',
            internalType: 'bytes'
          },
          {
            name: 'nodeSig',
            type: 'bytes',
            internalType: 'bytes'
          }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'NodeBalanceUpdated',
    inputs: [
      {
        name: 'token',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        indexed: false,
        internalType: 'uint256'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'TransferFailed',
    inputs: [
      {
        name: 'recipient',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'token',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        indexed: false,
        internalType: 'uint256'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ValidatorRegistered',
    inputs: [
      {
        name: 'validatorId',
        type: 'uint8',
        indexed: true,
        internalType: 'uint8'
      },
      {
        name: 'validator',
        type: 'address',
        indexed: true,
        internalType: 'contract ISignatureValidator'
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'Withdrawn',
    inputs: [
      {
        name: 'token',
        type: 'address',
        indexed: true,
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        indexed: false,
        internalType: 'uint256'
      }
    ],
    anonymous: false
  },
  {
    type: 'error',
    name: 'AddressCollision',
    inputs: [
      {
        name: 'collision',
        type: 'address',
        internalType: 'address'
      }
    ]
  },
  {
    type: 'error',
    name: 'ChallengerVersionTooLow',
    inputs: []
  },
  {
    type: 'error',
    name: 'ECDSAInvalidSignature',
    inputs: []
  },
  {
    type: 'error',
    name: 'ECDSAInvalidSignatureLength',
    inputs: [
      {
        name: 'length',
        type: 'uint256',
        internalType: 'uint256'
      }
    ]
  },
  {
    type: 'error',
    name: 'ECDSAInvalidSignatureS',
    inputs: [
      {
        name: 's',
        type: 'bytes32',
        internalType: 'bytes32'
      }
    ]
  },
  {
    type: 'error',
    name: 'EmptySignature',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectAmount',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectChallengeDuration',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectChannelId',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectChannelStatus',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectMsgSender',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectNode',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectSignature',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectStateIntent',
    inputs: []
  },
  {
    type: 'error',
    name: 'IncorrectValue',
    inputs: []
  },
  {
    type: 'error',
    name: 'InsufficientBalance',
    inputs: []
  },
  {
    type: 'error',
    name: 'InvalidAddress',
    inputs: []
  },
  {
    type: 'error',
    name: 'InvalidValidatorId',
    inputs: []
  },
  {
    type: 'error',
    name: 'NativeTransferFailed',
    inputs: [
      {
        name: 'to',
        type: 'address',
        internalType: 'address'
      },
      {
        name: 'amount',
        type: 'uint256',
        internalType: 'uint256'
      }
    ]
  },
  {
    type: 'error',
    name: 'NoChannelIdFoundForEscrow',
    inputs: []
  },
  {
    type: 'error',
    name: 'ReentrancyGuardReentrantCall',
    inputs: []
  },
  {
    type: 'error',
    name: 'SafeCastOverflowedIntToUint',
    inputs: [
      {
        name: 'value',
        type: 'int256',
        internalType: 'int256'
      }
    ]
  },
  {
    type: 'error',
    name: 'SafeERC20FailedOperation',
    inputs: [
      {
        name: 'token',
        type: 'address',
        internalType: 'address'
      }
    ]
  },
  {
    type: 'error',
    name: 'ValidatorAlreadyRegistered',
    inputs: [
      {
        name: 'validatorId',
        type: 'uint8',
        internalType: 'uint8'
      }
    ]
  },
  {
    type: 'error',
    name: 'ValidatorNotActive',
    inputs: [
      {
        name: 'validatorId',
        type: 'uint8',
        internalType: 'uint8'
      },
      {
        name: 'activatesAt',
        type: 'uint64',
        internalType: 'uint64'
      }
    ]
  },
  {
    type: 'error',
    name: 'ValidatorNotApproved',
    inputs: []
  },
  {
    type: 'error',
    name: 'ValidatorNotRegistered',
    inputs: [
      {
        name: 'validatorId',
        type: 'uint8',
        internalType: 'uint8'
      }
    ]
  }
] as const satisfies Abi;

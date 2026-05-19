import Decimal from 'decimal.js';
import {
  validateDecimalPrecision,
  decimalToBigInt,
  getHomeChannelId,
  getEscrowChannelId,
  generateChannelMetadata,
  transitionToIntent,
  getStateTransitionHash,
  SESSION_KEY_AUTH_TYPEHASH,
  packChannelKeyStateV1,
} from '../../../src/core/utils';
import {
  TransitionType,
  newTransition,
  INTENT_OPERATE,
  INTENT_CLOSE,
  INTENT_DEPOSIT,
  INTENT_WITHDRAW,
  INTENT_INITIATE_ESCROW_DEPOSIT,
  INTENT_FINALIZE_ESCROW_DEPOSIT,
  INTENT_INITIATE_ESCROW_WITHDRAWAL,
  INTENT_FINALIZE_ESCROW_WITHDRAWAL,
  INTENT_INITIATE_MIGRATION,
} from '../../../src/core/types';

describe('validateDecimalPrecision', () => {
  const tests = [
    {
      name: 'valid_6_decimals',
      amount: '1.123456',
      maxDecimals: 6,
      expectError: false,
      description: 'Amount with exactly 6 decimals should be valid for 6 decimal limit',
    },
    {
      name: 'valid_less_than_max',
      amount: '1.123',
      maxDecimals: 6,
      expectError: false,
      description: 'Amount with 3 decimals should be valid for 6 decimal limit',
    },
    {
      name: 'valid_whole_number',
      amount: '100',
      maxDecimals: 6,
      expectError: false,
      description: 'Whole number should be valid for any decimal limit',
    },
    {
      name: 'valid_zero',
      amount: '0',
      maxDecimals: 6,
      expectError: false,
      description: 'Zero should be valid for any decimal limit',
    },
    {
      name: 'invalid_too_many_decimals',
      amount: '1.1234567',
      maxDecimals: 6,
      expectError: true,
      description: 'Amount with 7 decimals should be invalid for 6 decimal limit',
    },
    {
      name: 'invalid_8_decimals',
      amount: '0.12345678',
      maxDecimals: 6,
      expectError: true,
      description: 'Amount with 8 decimals should be invalid for 6 decimal limit',
    },
    {
      name: 'valid_18_decimals_eth',
      amount: '1.123456789012345678',
      maxDecimals: 18,
      expectError: false,
      description: 'ETH amount with 18 decimals should be valid for 18 decimal limit',
    },
    {
      name: 'invalid_19_decimals_eth',
      amount: '1.1234567890123456789',
      maxDecimals: 18,
      expectError: true,
      description: 'Amount with 19 decimals should be invalid for 18 decimal limit',
    },
    {
      name: 'valid_usdc_6_decimals',
      amount: '1000.123456',
      maxDecimals: 6,
      expectError: false,
      description: 'USDC amount with 6 decimals should be valid',
    },
    {
      name: 'valid_small_amount',
      amount: '0.000001',
      maxDecimals: 6,
      expectError: false,
      description: 'Very small amount with 6 decimals should be valid',
    },
    {
      name: 'invalid_one_over_limit',
      amount: '0.0000001',
      maxDecimals: 6,
      expectError: true,
      description: 'Amount with one more decimal than allowed should be invalid',
    },
    {
      name: 'valid_large_number_no_decimals',
      amount: '1000000000',
      maxDecimals: 2,
      expectError: false,
      description: 'Large whole number should be valid regardless of decimal limit',
    },
    {
      name: 'valid_2_decimals',
      amount: '99.99',
      maxDecimals: 2,
      expectError: false,
      description: 'Amount with 2 decimals should be valid for 2 decimal limit',
    },
    {
      name: 'invalid_3_decimals_when_2_allowed',
      amount: '99.999',
      maxDecimals: 2,
      expectError: true,
      description: 'Amount with 3 decimals should be invalid for 2 decimal limit',
    },
  ];

  tests.forEach((tt) => {
    test(tt.name, () => {
      const amount = new Decimal(tt.amount);

      if (tt.expectError) {
        expect(() => validateDecimalPrecision(amount, tt.maxDecimals)).toThrow(
          'amount exceeds maximum decimal precision'
        );
      } else {
        expect(() => validateDecimalPrecision(amount, tt.maxDecimals)).not.toThrow();
      }
    });
  });

  describe('edge cases', () => {
    test('negative_amount', () => {
      const amount = new Decimal(-1.123456);
      expect(() => validateDecimalPrecision(amount, 6)).not.toThrow();
    });

    test('negative_amount_too_many_decimals', () => {
      const amount = new Decimal(-1.1234567);
      expect(() => validateDecimalPrecision(amount, 6)).toThrow();
    });

    test('very_large_amount', () => {
      const amount = new Decimal('999999999999999999.123456');
      expect(() => validateDecimalPrecision(amount, 6)).not.toThrow();
    });

    test('zero_decimal_limit', () => {
      const amount = new Decimal(100);
      expect(() => validateDecimalPrecision(amount, 0)).not.toThrow();

      const amountWithDecimals = new Decimal(100.1);
      expect(() => validateDecimalPrecision(amountWithDecimals, 0)).toThrow();
    });
  });
});

describe('decimalToBigInt', () => {
  const tests = [
    {
      name: 'usdc_whole_number',
      amount: '100',
      decimals: 6,
      expected: '100000000',
      description: '100 USDC should be 100000000 in smallest unit',
    },
    {
      name: 'usdc_with_decimals',
      amount: '1.23',
      decimals: 6,
      expected: '1230000',
      description: '1.23 USDC should be 1230000 in smallest unit',
    },
    {
      name: 'usdc_max_decimals',
      amount: '1.123456',
      decimals: 6,
      expected: '1123456',
      description: '1.123456 USDC should be 1123456 in smallest unit',
    },
    {
      name: 'usdc_small_amount',
      amount: '0.000001',
      decimals: 6,
      expected: '1',
      description: '0.000001 USDC (smallest unit) should be 1',
    },
    {
      name: 'eth_whole_number',
      amount: '1',
      decimals: 18,
      expected: '1000000000000000000',
      description: '1 ETH should be 1000000000000000000 wei',
    },
    {
      name: 'eth_with_decimals',
      amount: '1.5',
      decimals: 18,
      expected: '1500000000000000000',
      description: '1.5 ETH should be 1500000000000000000 wei',
    },
    {
      name: 'eth_gwei',
      amount: '0.000000001',
      decimals: 18,
      expected: '1000000000',
      description: '1 gwei (0.000000001 ETH) should be 1000000000 wei',
    },
    {
      name: 'eth_max_precision',
      amount: '1.123456789012345678',
      decimals: 18,
      expected: '1123456789012345678',
      description: 'ETH with 18 decimals should preserve full precision',
    },
    {
      name: 'zero_amount',
      amount: '0',
      decimals: 6,
      expected: '0',
      description: 'Zero amount should be zero',
    },
    {
      name: 'large_amount',
      amount: '1000000',
      decimals: 6,
      expected: '1000000000000',
      description: 'Large amount should be handled correctly',
    },
    {
      name: 'btc_like_8_decimals',
      amount: '0.00000001',
      decimals: 8,
      expected: '1',
      description: '1 satoshi (0.00000001 BTC) should be 1',
    },
    {
      name: 'btc_like_full_amount',
      amount: '21.12345678',
      decimals: 8,
      expected: '2112345678',
      description: 'BTC amount with 8 decimals should convert correctly',
    },
    {
      name: 'two_decimals_currency',
      amount: '99.99',
      decimals: 2,
      expected: '9999',
      description: 'Currency with 2 decimals (like cents) should convert correctly',
    },
    {
      name: 'zero_decimals',
      amount: '100',
      decimals: 0,
      expected: '100',
      description: 'Token with 0 decimals should remain unchanged',
    },
    {
      name: 'fractional_less_than_decimals',
      amount: '1.1',
      decimals: 6,
      expected: '1100000',
      description: 'Amount with fewer decimals than max should be scaled correctly',
    },
  ];

  tests.forEach((tt) => {
    test(tt.name, () => {
      const amount = new Decimal(tt.amount);
      const result = decimalToBigInt(amount, tt.decimals);
      expect(result.toString()).toBe(tt.expected);
    });
  });

  describe('negative amounts', () => {
    test('negative_usdc', () => {
      const amount = new Decimal(-1.23);
      const result = decimalToBigInt(amount, 6);
      expect(result.toString()).toBe('-1230000');
    });

    test('negative_eth', () => {
      const amount = new Decimal(-0.5);
      const result = decimalToBigInt(amount, 18);
      expect(result.toString()).toBe('-500000000000000000');
    });

    test('negative_zero', () => {
      const amount = new Decimal(0);
      const result = decimalToBigInt(amount, 6);
      expect(result.toString()).toBe('0');
    });
  });

  describe('edge cases', () => {
    test('very_large_amount', () => {
      const amount = new Decimal('999999999999999999.123456');
      const result = decimalToBigInt(amount, 6);
      expect(result.toString()).toBe('999999999999999999123456');
    });

    test('very_small_amount_exceeds_precision', () => {
      const amount = new Decimal(0.0000001); // 7 decimals
      expect(() => decimalToBigInt(amount, 6)).toThrow('precision');
    });

    test('max_uint8_decimals', () => {
      const amount = new Decimal(1);
      const result = decimalToBigInt(amount, 255);
      const expected = BigInt(10) ** BigInt(255);
      expect(result).toBe(expected);
    });

    test('precision_preservation', () => {
      const amount = new Decimal('123.456789');
      const result = decimalToBigInt(amount, 6);
      expect(result.toString()).toBe('123456789');
    });
  });

  describe('round trip', () => {
    test('usdc_round_trip', () => {
      const original = '1.123456';
      const amount = new Decimal(original);

      const bigIntValue = decimalToBigInt(amount, 6);
      const divisor = new Decimal(10).pow(6);
      const recovered = new Decimal(bigIntValue.toString()).div(divisor);

      expect(recovered.toString()).toBe(original);
    });
  });
});

describe('getHomeChannelId', () => {
  test('match_solidity_implementation', () => {
    const node = '0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a' as any;
    const user = '0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045' as any;
    const asset = 'ether';
    const nonce = 42n;
    const challengeDuration = 86400;

    const channelId = getHomeChannelId(node, user, asset, nonce, challengeDuration);

    const expected = '0x011d32827760cd3fa7dfb3934eb4ddb4a05f47e327581d4fd1585f4dc9a8c490';
    expect(channelId).toBe(expected);
  });

  test('different_assets_produce_different_ids', () => {
    const node = '0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a' as any;
    const user = '0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045' as any;
    const nonce = 42n;
    const challengeDuration = 86400;

    const channelId1 = getHomeChannelId(node, user, 'ether', nonce, challengeDuration);
    const channelId2 = getHomeChannelId(node, user, 'usdc', nonce, challengeDuration);

    expect(channelId1).not.toBe(channelId2);
  });

  test('different_nonces_produce_different_ids', () => {
    const node = '0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a' as any;
    const user = '0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045' as any;
    const asset = 'ether';
    const challengeDuration = 86400;

    const channelId1 = getHomeChannelId(node, user, asset, 1n, challengeDuration);
    const channelId2 = getHomeChannelId(node, user, asset, 2n, challengeDuration);

    expect(channelId1).not.toBe(channelId2);
  });
});

describe('getEscrowChannelId', () => {
  test('match_solidity_implementation', () => {
    const homeChannelId = '0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66';
    const version = 42n;

    const escrowId = getEscrowChannelId(homeChannelId, version);

    const expected = '0xe4d925dcf63add647f25c757d6ff0e74ba31401da91d8c7bafa4846c97a92ac2';
    expect(escrowId).toBe(expected);
  });

  test('different_versions_produce_different_ids', () => {
    const homeChannelId = '0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66';

    const escrowId1 = getEscrowChannelId(homeChannelId, 1n);
    const escrowId2 = getEscrowChannelId(homeChannelId, 2n);

    expect(escrowId1).not.toBe(escrowId2);
  });

  test('different_channels_produce_different_ids', () => {
    const version = 42n;

    const escrowId1 = getEscrowChannelId(
      '0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66',
      version
    );
    const escrowId2 = getEscrowChannelId(
      '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
      version
    );

    expect(escrowId1).not.toBe(escrowId2);
  });
});

describe('generateChannelMetadata', () => {
  test('match_solidity_implementation', () => {
    const asset = 'ether';
    const metadata = generateChannelMetadata(asset);

    // Expected from Solidity test: first 8 bytes of keccak256("ether")
    // keccak256("ether") = 0x13730b0d8e1bdbdc293b62ba010b1eede56b412ea2980defabe3d0b6c7844c3a
    // First 8 bytes: 0x13730b0d8e1bdbdc
    const expected = new Uint8Array([0x13, 0x73, 0x0b, 0x0d, 0x8e, 0x1b, 0xdb, 0xdc]);

    // Check metadata starts with 0x and convert
    expect(metadata.startsWith('0x')).toBe(true);
    const metadataBytes = Buffer.from(metadata.slice(2), 'hex');

    // Check first 8 bytes
    for (let i = 0; i < 8; i++) {
      expect(metadataBytes[i]).toBe(expected[i]);
    }

    // Rest should be zeros
    for (let i = 8; i < 32; i++) {
      expect(metadataBytes[i]).toBe(0);
    }
  });

  test('different_assets_produce_different_metadata', () => {
    const metadata1 = generateChannelMetadata('ether');
    const metadata2 = generateChannelMetadata('usdc');

    expect(metadata1).not.toBe(metadata2);
  });
});

describe('transitionToIntent', () => {
  describe('operate intents', () => {
    const tests = [
      { name: 'TransferSend', transitionType: TransitionType.TransferSend, expectedIntent: INTENT_OPERATE },
      {
        name: 'TransferReceive',
        transitionType: TransitionType.TransferReceive,
        expectedIntent: INTENT_OPERATE,
      },
      { name: 'Commit', transitionType: TransitionType.Commit, expectedIntent: INTENT_OPERATE },
      { name: 'Release', transitionType: TransitionType.Release, expectedIntent: INTENT_OPERATE },
    ];

    tests.forEach((tt) => {
      test(tt.name, () => {
        const transition = newTransition(tt.transitionType, '', '', new Decimal(0));
        const intent = transitionToIntent(transition);
        expect(intent).toBe(tt.expectedIntent);
      });
    });
  });

  describe('all transition types', () => {
    const tests = [
      { name: 'Finalize', transitionType: TransitionType.Finalize, expectedIntent: INTENT_CLOSE },
      { name: 'HomeDeposit', transitionType: TransitionType.HomeDeposit, expectedIntent: INTENT_DEPOSIT },
      {
        name: 'HomeWithdrawal',
        transitionType: TransitionType.HomeWithdrawal,
        expectedIntent: INTENT_WITHDRAW,
      },
      {
        name: 'MutualLock',
        transitionType: TransitionType.MutualLock,
        expectedIntent: INTENT_INITIATE_ESCROW_DEPOSIT,
      },
      {
        name: 'EscrowDeposit',
        transitionType: TransitionType.EscrowDeposit,
        expectedIntent: INTENT_FINALIZE_ESCROW_DEPOSIT,
      },
      {
        name: 'EscrowLock',
        transitionType: TransitionType.EscrowLock,
        expectedIntent: INTENT_INITIATE_ESCROW_WITHDRAWAL,
      },
      {
        name: 'EscrowWithdraw',
        transitionType: TransitionType.EscrowWithdraw,
        expectedIntent: INTENT_FINALIZE_ESCROW_WITHDRAWAL,
      },
      {
        name: 'Migrate',
        transitionType: TransitionType.Migrate,
        expectedIntent: INTENT_INITIATE_MIGRATION,
      },
    ];

    tests.forEach((tt) => {
      test(tt.name, () => {
        const transition = newTransition(tt.transitionType, '', '', new Decimal(0));
        const intent = transitionToIntent(transition);
        expect(intent).toBe(tt.expectedIntent);
      });
    });
  });

  test('void_transition', () => {
    const transition = newTransition(TransitionType.Void, '', '', new Decimal(0));
    const intent = transitionToIntent(transition);
    expect(intent).toBe(INTENT_OPERATE);
  });
});

describe('getStateTransitionHash', () => {
  test('hash_for_void_transition', () => {
    const transition = newTransition(TransitionType.Void, '', '', new Decimal(0));
    const hash = getStateTransitionHash(transition);
    expect(hash).toBeDefined();
  });

  test('hash_for_single_transition', () => {
    const transition = newTransition(
      TransitionType.HomeDeposit,
      '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', // 32-byte txId
      '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0', // 20-byte address
      new Decimal(1000)
    );
    const hash = getStateTransitionHash(transition);
    expect(hash).toBeDefined();
  });

  test('hash_with_negative_amount', () => {
    const transition = newTransition(
      TransitionType.HomeWithdrawal,
      '0x4444444444444444444444444444444444444444444444444444444444444444', // 32-byte txId
      '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0', // 20-byte address
      new Decimal(-100)
    );
    const hash = getStateTransitionHash(transition);
    expect(hash).toBeDefined();
  });
});

describe('SESSION_KEY_AUTH_TYPEHASH', () => {
    // Golden value cross-checked against Solidity: keccak256("Nitrolite.SessionKey(address sessionKey,bytes32 metadataHash)")
    // Verified by running Solidity test test_log_toSigningData in SessionKeyValidator.t.sol.
    const EXPECTED = '0x251773da8b8949935ef07284d20cc8605ad7d6f4cf6b5e040ce07dae857f0b6c';

    test('matches_solidity_constant', () => {
        expect(SESSION_KEY_AUTH_TYPEHASH).toBe(EXPECTED);
    });
});

describe('packChannelKeyStateV1', () => {
    const sessionKey = '0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF' as `0x${string}`;
    const metadataHash = '0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890' as `0x${string}`;
    // abi.encode(SESSION_KEY_AUTH_TYPEHASH, sessionKey, metadataHash)
    const EXPECTED =
        '0x' +
        '251773da8b8949935ef07284d20cc8605ad7d6f4cf6b5e040ce07dae857f0b6c' + // typehash
        '000000000000000000000000deadbeefdeadbeefdeadbeefdeadbeefdeadbeef' + // address padded
        'abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890';  // metadataHash

    test('golden_value', () => {
        const packed = packChannelKeyStateV1(sessionKey, metadataHash);
        expect(packed.toLowerCase()).toBe(EXPECTED);
    });
});

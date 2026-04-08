package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeHexAddress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid_lowercase",
			input: "0x1234567890abcdef1234567890abcdef12345678",
			want:  "0x1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:  "valid_uppercase_normalized",
			input: "0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
			want:  "0xabcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:  "valid_mixed_case",
			input: "0xAbCdEf1234567890aBcDeF1234567890AbCdEf12",
			want:  "0xabcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:  "valid_checksum_address",
			input: "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
			want:  "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed",
		},
		{
			name:    "missing_0x_prefix",
			input:   "1234567890abcdef1234567890abcdef12345678",
			wantErr: true,
		},
		{
			name:    "too_short",
			input:   "0x1234",
			wantErr: true,
		},
		{
			name:    "too_long",
			input:   "0x1234567890abcdef1234567890abcdef1234567890",
			wantErr: true,
		},
		{
			name:    "empty_string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid_hex_char",
			input:   "0x1234567890abcdef1234567890abcdef1234567g",
			wantErr: true,
		},
		{
			name:  "all_zeros",
			input: "0x0000000000000000000000000000000000000000",
			want:  "0x0000000000000000000000000000000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeHexAddress(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateDecimalPrecision(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		amount      string
		maxDecimals uint8
		expectError bool
		description string
	}{
		{
			name:        "valid_6_decimals",
			amount:      "1.123456",
			maxDecimals: 6,
			expectError: false,
			description: "Amount with exactly 6 decimals should be valid for 6 decimal limit",
		},
		{
			name:        "valid_less_than_max",
			amount:      "1.123",
			maxDecimals: 6,
			expectError: false,
			description: "Amount with 3 decimals should be valid for 6 decimal limit",
		},
		{
			name:        "valid_whole_number",
			amount:      "100",
			maxDecimals: 6,
			expectError: false,
			description: "Whole number should be valid for any decimal limit",
		},
		{
			name:        "valid_zero",
			amount:      "0",
			maxDecimals: 6,
			expectError: false,
			description: "Zero should be valid for any decimal limit",
		},
		{
			name:        "invalid_too_many_decimals",
			amount:      "1.1234567",
			maxDecimals: 6,
			expectError: true,
			description: "Amount with 7 decimals should be invalid for 6 decimal limit",
		},
		{
			name:        "invalid_8_decimals",
			amount:      "0.12345678",
			maxDecimals: 6,
			expectError: true,
			description: "Amount with 8 decimals should be invalid for 6 decimal limit",
		},
		{
			name:        "valid_18_decimals_eth",
			amount:      "1.123456789012345678",
			maxDecimals: 18,
			expectError: false,
			description: "ETH amount with 18 decimals should be valid for 18 decimal limit",
		},
		{
			name:        "invalid_19_decimals_eth",
			amount:      "1.1234567890123456789",
			maxDecimals: 18,
			expectError: true,
			description: "Amount with 19 decimals should be invalid for 18 decimal limit",
		},
		{
			name:        "valid_usdc_6_decimals",
			amount:      "1000.123456",
			maxDecimals: 6,
			expectError: false,
			description: "USDC amount with 6 decimals should be valid",
		},
		{
			name:        "valid_small_amount",
			amount:      "0.000001",
			maxDecimals: 6,
			expectError: false,
			description: "Very small amount with 6 decimals should be valid",
		},
		{
			name:        "invalid_one_over_limit",
			amount:      "0.0000001",
			maxDecimals: 6,
			expectError: true,
			description: "Amount with one more decimal than allowed should be invalid",
		},
		{
			name:        "valid_large_number_no_decimals",
			amount:      "1000000000",
			maxDecimals: 2,
			expectError: false,
			description: "Large whole number should be valid regardless of decimal limit",
		},
		{
			name:        "valid_2_decimals",
			amount:      "99.99",
			maxDecimals: 2,
			expectError: false,
			description: "Amount with 2 decimals should be valid for 2 decimal limit",
		},
		{
			name:        "invalid_3_decimals_when_2_allowed",
			amount:      "99.999",
			maxDecimals: 2,
			expectError: true,
			description: "Amount with 3 decimals should be invalid for 2 decimal limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			amount, err := decimal.NewFromString(tt.amount)
			assert.NoError(t, err, "Test setup error: invalid amount string")

			err = ValidateDecimalPrecision(amount, tt.maxDecimals)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "amount exceeds maximum decimal precision")
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestValidateDecimalPrecision_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("negative_amount", func(t *testing.T) {
		t.Parallel()
		amount := decimal.NewFromFloat(-1.123456)
		err := ValidateDecimalPrecision(amount, 6)
		assert.NoError(t, err, "Negative amounts should be validated the same as positive")
	})

	t.Run("negative_amount_too_many_decimals", func(t *testing.T) {
		t.Parallel()
		amount := decimal.NewFromFloat(-1.1234567)
		err := ValidateDecimalPrecision(amount, 6)
		assert.Error(t, err, "Negative amount with too many decimals should fail")
	})

	t.Run("very_large_amount", func(t *testing.T) {
		t.Parallel()
		amount, err := decimal.NewFromString("999999999999999999.123456")
		assert.NoError(t, err)
		err = ValidateDecimalPrecision(amount, 6)
		// This should pass as long as decimals are within limit
		assert.NoError(t, err)
	})

	t.Run("zero_decimal_limit", func(t *testing.T) {
		t.Parallel()
		amount := decimal.NewFromInt(100)
		err := ValidateDecimalPrecision(amount, 0)
		assert.NoError(t, err, "Whole number should be valid for 0 decimal limit")

		amountWithDecimals := decimal.NewFromFloat(100.1)
		err = ValidateDecimalPrecision(amountWithDecimals, 0)
		assert.Error(t, err, "Amount with decimals should fail for 0 decimal limit")
	})
}

func TestDecimalToBigInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		amount      string
		decimals    uint8
		expected    string
		description string
	}{
		{
			name:        "usdc_whole_number",
			amount:      "100",
			decimals:    6,
			expected:    "100000000", // 100 * 10^6
			description: "100 USDC should be 100000000 in smallest unit",
		},
		{
			name:        "usdc_with_decimals",
			amount:      "1.23",
			decimals:    6,
			expected:    "1230000", // 1.23 * 10^6
			description: "1.23 USDC should be 1230000 in smallest unit",
		},
		{
			name:        "usdc_max_decimals",
			amount:      "1.123456",
			decimals:    6,
			expected:    "1123456", // 1.123456 * 10^6
			description: "1.123456 USDC should be 1123456 in smallest unit",
		},
		{
			name:        "usdc_small_amount",
			amount:      "0.000001",
			decimals:    6,
			expected:    "1", // 0.000001 * 10^6 = 1
			description: "0.000001 USDC (smallest unit) should be 1",
		},
		{
			name:        "eth_whole_number",
			amount:      "1",
			decimals:    18,
			expected:    "1000000000000000000", // 1 * 10^18
			description: "1 ETH should be 1000000000000000000 wei",
		},
		{
			name:        "eth_with_decimals",
			amount:      "1.5",
			decimals:    18,
			expected:    "1500000000000000000", // 1.5 * 10^18
			description: "1.5 ETH should be 1500000000000000000 wei",
		},
		{
			name:        "eth_gwei",
			amount:      "0.000000001",
			decimals:    18,
			expected:    "1000000000", // 1 gwei = 10^9 wei
			description: "1 gwei (0.000000001 ETH) should be 1000000000 wei",
		},
		{
			name:        "eth_max_precision",
			amount:      "1.123456789012345678",
			decimals:    18,
			expected:    "1123456789012345678", // 1.123456789012345678 * 10^18
			description: "ETH with 18 decimals should preserve full precision",
		},
		{
			name:        "zero_amount",
			amount:      "0",
			decimals:    6,
			expected:    "0",
			description: "Zero amount should be zero",
		},
		{
			name:        "large_amount",
			amount:      "1000000",
			decimals:    6,
			expected:    "1000000000000", // 1000000 * 10^6
			description: "Large amount should be handled correctly",
		},
		{
			name:        "btc_like_8_decimals",
			amount:      "0.00000001",
			decimals:    8,
			expected:    "1", // 0.00000001 * 10^8 = 1 satoshi
			description: "1 satoshi (0.00000001 BTC) should be 1",
		},
		{
			name:        "btc_like_full_amount",
			amount:      "21.12345678",
			decimals:    8,
			expected:    "2112345678", // 21.12345678 * 10^8
			description: "BTC amount with 8 decimals should convert correctly",
		},
		{
			name:        "two_decimals_currency",
			amount:      "99.99",
			decimals:    2,
			expected:    "9999", // 99.99 * 10^2
			description: "Currency with 2 decimals (like cents) should convert correctly",
		},
		{
			name:        "zero_decimals",
			amount:      "100",
			decimals:    0,
			expected:    "100", // 100 * 10^0 = 100
			description: "Token with 0 decimals should remain unchanged",
		},
		{
			name:        "fractional_less_than_decimals",
			amount:      "1.1",
			decimals:    6,
			expected:    "1100000", // 1.1 * 10^6
			description: "Amount with fewer decimals than max should be scaled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			amount, err := decimal.NewFromString(tt.amount)
			assert.NoError(t, err, "Test setup error: invalid amount string")

			result, err := DecimalToBigInt(amount, tt.decimals)
			assert.NoError(t, err)

			expected, ok := new(big.Int).SetString(tt.expected, 10)
			assert.True(t, ok, "Test setup error: invalid expected value")

			assert.Equal(t, expected.String(), result.String(), tt.description)
		})
	}
}

func TestDecimalToBigInt_NegativeAmounts(t *testing.T) {
	t.Parallel()
	t.Run("negative_usdc", func(t *testing.T) {
		t.Parallel()
		amount := decimal.NewFromFloat(-1.23)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		expected := big.NewInt(-1230000)
		assert.Equal(t, expected.String(), result.String(), "Negative amounts should be handled correctly")
	})

	t.Run("negative_eth", func(t *testing.T) {
		t.Parallel()
		amount := decimal.NewFromFloat(-0.5)
		result, err := DecimalToBigInt(amount, 18)
		assert.NoError(t, err)
		expected, _ := new(big.Int).SetString("-500000000000000000", 10)
		assert.Equal(t, expected.String(), result.String(), "Negative ETH amount should convert correctly")
	})

	t.Run("negative_zero", func(t *testing.T) {
		t.Parallel()
		amount := decimal.NewFromInt(0)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		expected := big.NewInt(0)
		assert.Equal(t, expected.String(), result.String(), "Zero should always be zero")
	})
}

func TestDecimalToBigInt_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("very_large_amount", func(t *testing.T) {
		t.Parallel()
		// Test with a very large amount
		amount, err := decimal.NewFromString("999999999999999999.123456")
		assert.NoError(t, err)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		// 999999999999999999.123456 * 10^6 = 999999999999999999123456
		expected, ok := new(big.Int).SetString("999999999999999999123456", 10)
		assert.True(t, ok)
		assert.Equal(t, expected.String(), result.String(), "Very large amounts should be handled")
	})

	t.Run("very_small_amount", func(t *testing.T) {
		t.Parallel()
		// Test with an amount that has more decimals than supported
		amount := decimal.NewFromFloat(0.0000001) // 7 decimals
		_, err := DecimalToBigInt(amount, 6)      // Only 6 decimal precision
		// This should return an error because the amount has more decimal places than allowed
		// After scaling: 0.0000001 * 10^6 = 0.1, which still has a fractional part
		assert.Error(t, err, "Amount with more decimals than supported should return an error")
		assert.Contains(t, err.Error(), "precision", "Error should mention precision")
	})

	t.Run("max_uint8_decimals", func(t *testing.T) {
		t.Parallel()
		// Test with maximum uint8 value for decimals (not practical, but edge case)
		amount := decimal.NewFromInt(1)
		result, err := DecimalToBigInt(amount, 255)
		assert.NoError(t, err)
		// 1 * 10^255 should work
		expected := new(big.Int).Exp(big.NewInt(10), big.NewInt(255), nil)
		assert.Equal(t, expected.String(), result.String(), "Maximum decimals should work")
	})

	t.Run("precision_preservation", func(t *testing.T) {
		t.Parallel()
		// Test that we don't lose precision during conversion
		amount, err := decimal.NewFromString("123.456789")
		assert.NoError(t, err)
		result, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		// 123.456789 * 10^6 = 123456789
		expected := big.NewInt(123456789)
		assert.Equal(t, expected.String(), result.String(), "Precision should be preserved")
	})
}

func TestDecimalToBigInt_RoundTrip(t *testing.T) {
	t.Parallel()
	t.Run("usdc_round_trip", func(t *testing.T) {
		t.Parallel()
		// Test that we can convert back and forth without losing precision
		original := "1.123456"
		amount, err := decimal.NewFromString(original)
		assert.NoError(t, err)

		// Convert to big.Int
		bigIntValue, err := DecimalToBigInt(amount, 6)
		assert.NoError(t, err)
		// Convert back to decimal
		divisor := decimal.New(1, 6) // 10^6
		recovered := decimal.NewFromBigInt(bigIntValue, 0).Div(divisor)

		assert.Equal(t, original, recovered.String(), "Round trip conversion should preserve value")
	})
}

func TestGetHomeChannelID(t *testing.T) {
	t.Parallel()
	t.Run("match_solidity_implementation", func(t *testing.T) {
		t.Parallel()
		// Test values from contracts/test/Utils.t.sol:test_log_calculateChannelId
		node := "0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a"
		user := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
		asset := "ether"
		nonce := uint64(42)
		challengeDuration := uint32(86400)

		channelID, err := GetHomeChannelID(node, user, asset, nonce, challengeDuration, "0x0")
		assert.NoError(t, err)

		// Expected value from Solidity test (with version byte 0x01, approvedSignatureValidators: 0)
		expected := "0x011d32827760cd3fa7dfb3934eb4ddb4a05f47e327581d4fd1585f4dc9a8c490"
		assert.Equal(t, expected, channelID, "Channel ID should match Solidity implementation")
	})

	t.Run("different_versions_produce_different_ids", func(t *testing.T) {
		t.Parallel()
		// This test matches contracts/test/Utils.t.sol:test_channelId_forDifferentVersions_differ
		node := "0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a"
		user := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
		asset := "ether"
		nonce := uint64(42)
		challengeDuration := uint32(86400)

		channelIDV1, err := getHomeChannelID(node, user, asset, nonce, challengeDuration, "0x0", 1)
		assert.NoError(t, err)

		channelIDV2, err := getHomeChannelID(node, user, asset, nonce, challengeDuration, "0x0", 2)
		assert.NoError(t, err)

		channelIDV255, err := getHomeChannelID(node, user, asset, nonce, challengeDuration, "0x0", 255)
		assert.NoError(t, err)

		// Channel IDs must differ for different versions
		assert.NotEqual(t, channelIDV1, channelIDV2, "channelIdV1 should differ from channelIdV2")
		assert.NotEqual(t, channelIDV1, channelIDV255, "channelIdV1 should differ from channelIdV255")
		assert.NotEqual(t, channelIDV2, channelIDV255, "channelIdV2 should differ from channelIdV255")

		// Expected values (with approvedSignatureValidators: 0)
		expectedV1 := "0x011d32827760cd3fa7dfb3934eb4ddb4a05f47e327581d4fd1585f4dc9a8c490"
		expectedV2 := "0x021d32827760cd3fa7dfb3934eb4ddb4a05f47e327581d4fd1585f4dc9a8c490"
		expectedV255 := "0xff1d32827760cd3fa7dfb3934eb4ddb4a05f47e327581d4fd1585f4dc9a8c490"

		assert.Equal(t, expectedV1, channelIDV1, "Version 1 channel ID should match Solidity")
		assert.Equal(t, expectedV2, channelIDV2, "Version 2 channel ID should match Solidity")
		assert.Equal(t, expectedV255, channelIDV255, "Version 255 channel ID should match Solidity")

		// First byte should match the version
		assert.Equal(t, byte(1), common.HexToHash(channelIDV1)[0], "First byte of V1 should be 0x01")
		assert.Equal(t, byte(2), common.HexToHash(channelIDV2)[0], "First byte of V2 should be 0x02")
		assert.Equal(t, byte(255), common.HexToHash(channelIDV255)[0], "First byte of V255 should be 0xFF")

		// All other bytes should be the same (derived from the same base hash)
		v1Hash := common.HexToHash(channelIDV1)
		v2Hash := common.HexToHash(channelIDV2)
		v255Hash := common.HexToHash(channelIDV255)

		assert.Equal(t, v1Hash[1:], v2Hash[1:], "Bytes 1-31 should be the same for V1 and V2")
		assert.Equal(t, v1Hash[1:], v255Hash[1:], "Bytes 1-31 should be the same for V1 and V255")
	})

	t.Run("different_assets_produce_different_ids", func(t *testing.T) {
		t.Parallel()
		node := "0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a"
		user := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
		nonce := uint64(42)
		challengeDuration := uint32(86400)

		channelID1, err := GetHomeChannelID(node, user, "ether", nonce, challengeDuration, "0x0")
		assert.NoError(t, err)

		channelID2, err := GetHomeChannelID(node, user, "usdc", nonce, challengeDuration, "0x0")
		assert.NoError(t, err)

		assert.NotEqual(t, channelID1, channelID2, "Different assets should produce different channel IDs")
	})

	t.Run("different_nonces_produce_different_ids", func(t *testing.T) {
		t.Parallel()
		node := "0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a"
		user := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
		asset := "ether"
		challengeDuration := uint32(86400)

		channelID1, err := GetHomeChannelID(node, user, asset, 1, challengeDuration, "0x0")
		assert.NoError(t, err)

		channelID2, err := GetHomeChannelID(node, user, asset, 2, challengeDuration, "0x0")
		assert.NoError(t, err)

		assert.NotEqual(t, channelID1, channelID2, "Different nonces should produce different channel IDs")
	})
}

func TestGetEscrowChannelID(t *testing.T) {
	t.Parallel()
	t.Run("match_solidity_implementation", func(t *testing.T) {
		t.Parallel()
		// Test values from contracts-new/test/Utils.t.sol:test_log_calculateEscrowId
		homeChannelID := "0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66"
		version := uint64(42)

		escrowID, err := GetEscrowChannelID(homeChannelID, version)
		assert.NoError(t, err)

		// Expected value from Solidity test
		expected := "0xe4d925dcf63add647f25c757d6ff0e74ba31401da91d8c7bafa4846c97a92ac2"
		assert.Equal(t, expected, escrowID, "Escrow ID should match Solidity implementation")
	})

	t.Run("different_versions_produce_different_ids", func(t *testing.T) {
		t.Parallel()
		homeChannelID := "0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66"

		escrowID1, err := GetEscrowChannelID(homeChannelID, 1)
		assert.NoError(t, err)

		escrowID2, err := GetEscrowChannelID(homeChannelID, 2)
		assert.NoError(t, err)

		assert.NotEqual(t, escrowID1, escrowID2, "Different versions should produce different escrow IDs")
	})

	t.Run("different_channels_produce_different_ids", func(t *testing.T) {
		t.Parallel()
		version := uint64(42)

		escrowID1, err := GetEscrowChannelID("0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66", version)
		assert.NoError(t, err)

		escrowID2, err := GetEscrowChannelID("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", version)
		assert.NoError(t, err)

		assert.NotEqual(t, escrowID1, escrowID2, "Different home channel IDs should produce different escrow IDs")
	})
}

func TestGenerateChannelMetadata(t *testing.T) {
	t.Parallel()
	t.Run("match_solidity_implementation", func(t *testing.T) {
		t.Parallel()
		// Test that metadata generation matches Solidity
		asset := "ether"
		metadata := GenerateChannelMetadata(asset)

		// Expected from Solidity test: first 8 bytes of keccak256("ether")
		// keccak256("ether") = 0x13730b0d8e1bdbdc293b62ba010b1eede56b412ea2980defabe3d0b6c7844c3a
		// First 8 bytes: 0x13730b0d8e1bdbdc
		expected := [32]byte{0x13, 0x73, 0x0b, 0x0d, 0x8e, 0x1b, 0xdb, 0xdc}

		assert.Equal(t, expected[:8], metadata[:8], "First 8 bytes should match asset hash")

		// Rest should be zeros
		for i := 8; i < 32; i++ {
			assert.Equal(t, byte(0), metadata[i], "Bytes after index 8 should be zero")
		}
	})

	t.Run("different_assets_produce_different_metadata", func(t *testing.T) {
		t.Parallel()
		metadata1 := GenerateChannelMetadata("ether")
		metadata2 := GenerateChannelMetadata("usdc")

		assert.NotEqual(t, metadata1, metadata2, "Different assets should produce different metadata")
	})
}

func TestTransitionToIntent_OperateIntents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		transitionType TransitionType
		expectedIntent uint8
	}{
		{"TransferSend", TransitionTypeTransferSend, INTENT_OPERATE},
		{"TransferReceive", TransitionTypeTransferReceive, INTENT_OPERATE},
		{"Commit", TransitionTypeCommit, INTENT_OPERATE},
		{"Release", TransitionTypeRelease, INTENT_OPERATE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			transition := Transition{Type: tt.transitionType}
			intent := TransitionToIntent(transition)
			require.Equal(t, tt.expectedIntent, intent)
		})
	}
}

func TestTransitionToIntent_AllTransitionTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		transitionType TransitionType
		expectedIntent uint8
	}{
		{"Finalize", TransitionTypeFinalize, INTENT_CLOSE},
		{"HomeDeposit", TransitionTypeHomeDeposit, INTENT_DEPOSIT},
		{"HomeWithdrawal", TransitionTypeHomeWithdrawal, INTENT_WITHDRAW},
		{"MutualLock", TransitionTypeMutualLock, INTENT_INITIATE_ESCROW_DEPOSIT},
		{"EscrowDeposit", TransitionTypeEscrowDeposit, INTENT_FINALIZE_ESCROW_DEPOSIT},
		{"EscrowLock", TransitionTypeEscrowLock, INTENT_INITIATE_ESCROW_WITHDRAWAL},
		{"EscrowWithdraw", TransitionTypeEscrowWithdraw, INTENT_FINALIZE_ESCROW_WITHDRAWAL},
		{"Migrate", TransitionTypeMigrate, INTENT_INITIATE_MIGRATION},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			transition := &Transition{Type: tt.transitionType}
			intent := TransitionToIntent(*transition)
			require.Equal(t, tt.expectedIntent, intent)
		})
	}
}

func TestGetStateTransitionsHash(t *testing.T) {
	t.Parallel()
	t.Run("print_hash_for_single_transition", func(t *testing.T) {
		t.Parallel()
		hash, err := GetStateTransitionHash(*NewTransition(TransitionTypeHomeDeposit,
			"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", // 32-byte txId
			"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",                         // 20-byte address
			decimal.NewFromInt(1000)))
		assert.NoError(t, err)

		t.Logf("Hash for single transition: 0x%x", hash)
	})

	t.Run("print_hash_with_negative_amounts", func(t *testing.T) {
		t.Parallel()
		_, err := GetStateTransitionHash(*NewTransition(TransitionTypeHomeWithdrawal,
			"0x4444444444444444444444444444444444444444444444444444444444444444", // 32-byte txId
			"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",                         // 20-byte address
			decimal.NewFromInt(-100)))
		assert.NoError(t, err)
	})
}

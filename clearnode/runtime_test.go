package main

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/layer-3/nitrolite/pkg/core"
)

func TestCheckChannelHubVersion_Manual(t *testing.T) {
	t.Skip("Manual test - uncomment test cases and set values before running")

	blockchainRPC := "https://rpc.example.com"                                                    // TODO: Set RPC URL
	correctChannelHubAddress := common.HexToAddress("0x0000000000000000000000000000000000000000") // TODO: Set ChannelHub address

	testCases := []struct {
		name              string
		blockchainRPC     string
		channelHubAddress common.Address
		expectedVersion   uint8
		expectError       bool
		errorContains     string // Expected substring in error message
	}{
		{
			name:              "valid ChannelHub with correct version",
			blockchainRPC:     blockchainRPC,
			channelHubAddress: correctChannelHubAddress,
			expectedVersion:   core.ChannelHubVersion,
			expectError:       false,
		},
		{
			name:              "invalid contract address (no code)",
			blockchainRPC:     blockchainRPC,
			channelHubAddress: common.HexToAddress("0x0000000000000000000000000000000000004242"), // Address with no contract
			expectedVersion:   core.ChannelHubVersion,
			expectError:       true,
			errorContains:     "failed to get ChannelHub version",
		},
		{
			name:              "version mismatch",
			blockchainRPC:     blockchainRPC,
			channelHubAddress: correctChannelHubAddress,
			expectedVersion:   99, // Wrong version to trigger mismatch
			expectError:       true,
			errorContains:     "configured and fetched ChannelHub version mismatch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate test case is properly configured
			if tc.blockchainRPC == "" {
				t.Fatal("blockchainRPC must be set for this test case")
			}
			if tc.channelHubAddress == (common.Address{}) {
				t.Fatal("channelHubAddress must be set for this test case")
			}

			// Run the version check
			err := checkChannelHubVersion(tc.blockchainRPC, tc.channelHubAddress, tc.expectedVersion)

			// Verify results
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errorContains)
				}
				if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Fatalf("expected error containing %q, got: %v", tc.errorContains, err)
				}
				t.Logf("Got expected error: %v", err)
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				t.Log("ChannelHub version check passed successfully")
			}
		})
	}
}

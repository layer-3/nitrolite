package main

import (
	"strings"
	"testing"
)

func TestValidateChannelChallengeConfig(t *testing.T) {
	tests := []struct {
		name          string
		minChallenge  uint32
		maxChallenge  uint32
		wantErr       bool
		errorContains string
	}{
		{
			name:         "default range passes",
			minChallenge: 86400,
			maxChallenge: 604800,
		},
		{
			name:         "stricter range passes",
			minChallenge: 172800,
			maxChallenge: 345600,
		},
		{
			name:          "max above contract limit fails",
			minChallenge:  86400,
			maxChallenge:  604801,
			wantErr:       true,
			errorContains: "NITRONODE_CHANNEL_MAX_CHALLENGE_DURATION",
		},
		{
			name:          "min below contract limit fails",
			minChallenge:  86399,
			maxChallenge:  604800,
			wantErr:       true,
			errorContains: "NITRONODE_CHANNEL_MIN_CHALLENGE_DURATION",
		},
		{
			name:          "min greater than max fails",
			minChallenge:  604800,
			maxChallenge:  86400,
			wantErr:       true,
			errorContains: "must be <=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChannelChallengeConfig(tt.minChallenge, tt.maxChallenge)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

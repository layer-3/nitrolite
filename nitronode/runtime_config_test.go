package main

import (
	"strings"
	"testing"

	"github.com/layer-3/nitrolite/pkg/core"
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
			minChallenge: core.ChannelMinChallengeDuration,
			maxChallenge: core.ChannelMaxChallengeDuration,
		},
		{
			name:         "stricter range passes",
			minChallenge: 172800,
			maxChallenge: 345600,
		},
		{
			name:          "max above contract limit fails",
			minChallenge:  core.ChannelMinChallengeDuration,
			maxChallenge:  core.ChannelMaxChallengeDuration + 1,
			wantErr:       true,
			errorContains: "NITRONODE_CHANNEL_MAX_CHALLENGE_DURATION",
		},
		{
			name:          "min below contract limit fails",
			minChallenge:  core.ChannelMinChallengeDuration - 1,
			maxChallenge:  core.ChannelMaxChallengeDuration,
			wantErr:       true,
			errorContains: "NITRONODE_CHANNEL_MIN_CHALLENGE_DURATION",
		},
		{
			name:          "min greater than max fails",
			minChallenge:  core.ChannelMaxChallengeDuration,
			maxChallenge:  core.ChannelMinChallengeDuration,
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

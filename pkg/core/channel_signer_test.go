package core

import "testing"

func TestIsChannelSignerSupported(t *testing.T) {
	tests := []struct {
		name       string
		bitmap     string
		signerType ChannelSignerType
		want       bool
	}{
		{
			name:       "default signer supported when set in bitmap",
			bitmap:     "0x01",
			signerType: ChannelSignerType_Default,
			want:       true,
		},
		{
			name:       "default signer supported even when not in bitmap",
			bitmap:     "0x02",
			signerType: ChannelSignerType_Default,
			want:       true,
		},
		{
			name:       "default signer supported with empty bitmap",
			bitmap:     "0x00",
			signerType: ChannelSignerType_Default,
			want:       true,
		},
		{
			name:       "session key supported when set in bitmap",
			bitmap:     "0x02",
			signerType: ChannelSignerType_SessionKey,
			want:       true,
		},
		{
			name:       "session key not supported when not in bitmap",
			bitmap:     "0x01",
			signerType: ChannelSignerType_SessionKey,
			want:       false,
		},
		{
			name:       "session key supported when both set in bitmap",
			bitmap:     "0x03",
			signerType: ChannelSignerType_SessionKey,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsChannelSignerSupported(tt.bitmap, tt.signerType)
			if got != tt.want {
				t.Errorf("IsChannelSignerSupported(%q, %v) = %v, want %v",
					tt.bitmap, tt.signerType, got, tt.want)
			}
		})
	}
}

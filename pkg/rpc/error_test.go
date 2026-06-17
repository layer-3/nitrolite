package rpc

import (
	"encoding/json"
	"testing"
)

func TestNewErrorPayload(t *testing.T) {
	cases := map[string]string{
		"plain":          "invalid address format",
		"embedded quote": `duplicate key value violates unique constraint "app_sessions_v1_pkey"`,
		"backslash":      `path\to\thing`,
		"newline":        "line1\nline2",
		"control char":   "tab\there",
		"leading quote":  `"quoted"`,
		"empty":          "",
	}

	for name, msg := range cases {
		t.Run(name, func(t *testing.T) {
			payload := NewErrorPayload(msg)

			// The stored value must be valid JSON that round-trips to the
			// original string.
			var got string
			if err := json.Unmarshal(payload[errorParamKey], &got); err != nil {
				t.Fatalf("stored value is not valid JSON: %v", err)
			}
			if got != msg {
				t.Fatalf("round-trip mismatch: got %q, want %q", got, msg)
			}

			// The whole payload must marshal without error. This is the
			// regression guard for the websocket-response marshal failure.
			if _, err := json.Marshal(payload); err != nil {
				t.Fatalf("payload failed to marshal: %v", err)
			}
		})
	}
}

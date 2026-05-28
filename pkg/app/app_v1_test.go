package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidApplicationID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"simple", "my-app", true},
		{"underscore", "my_app", true},
		{"digits", "app-123", true},
		{"dash only", "-", true},
		{"underscore only", "_", true},
		{"single char", "a", true},
		{"exactly 66 chars", strings.Repeat("a", 66), true},

		{"empty", "", false},
		{"67 chars", strings.Repeat("a", 67), false},
		{"uppercase", "MyApp", false},
		{"space", "my app", false},
		{"dot", "my.app", false},
		{"slash", "my/app", false},
		{"newline", "my\napp", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsValidApplicationID(tc.id))
		})
	}
}

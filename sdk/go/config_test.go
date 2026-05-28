package sdk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	assert.Equal(t, 5*time.Second, DefaultConfig.HandshakeTimeout)
	assert.Equal(t, 15*time.Second, DefaultConfig.PingTimeout)
	assert.NotNil(t, DefaultConfig.ErrorHandler)
}

func TestWithHandshakeTimeout(t *testing.T) {
	c := &Config{}
	opt := WithHandshakeTimeout(10 * time.Second)
	opt(c)
	assert.Equal(t, 10*time.Second, c.HandshakeTimeout)
}

func TestWithPingTimeout(t *testing.T) {
	c := &Config{}
	opt := WithPingTimeout(20 * time.Second)
	opt(c)
	assert.Equal(t, 20*time.Second, c.PingTimeout)
}

func TestWithErrorHandler(t *testing.T) {
	c := &Config{}
	called := false
	handler := func(err error) {
		called = true
	}
	opt := WithErrorHandler(handler)
	opt(c)
	
	assert.NotNil(t, c.ErrorHandler)
	c.ErrorHandler(nil)
	assert.True(t, called)
}

func TestDefaultErrorHandler(t *testing.T) {
	// Just ensure it doesn't panic when called with nil or error
	// Capturing stderr is harder and maybe overkill for this simple function
	defaultErrorHandler(nil)
	// defaultErrorHandler(errors.New("test error")) // verification would require capturing stderr
}

func TestWithApplicationID(t *testing.T) {
	c := &Config{}
	WithApplicationID("my-app")(c)
	assert.Equal(t, "my-app", c.ApplicationID)
}

func TestAppendApplicationIDQueryParam(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		appID   string
		want    string
		wantErr bool
	}{
		{name: "empty app_id returns url unchanged", url: "ws://host/path", appID: "", want: "ws://host/path"},
		{name: "adds app_id when no query", url: "ws://host/path", appID: "my-app", want: "ws://host/path?app_id=my-app"},
		{name: "adds app_id alongside existing query", url: "ws://host/path?foo=bar", appID: "my-app", want: "ws://host/path?app_id=my-app&foo=bar"},
		{name: "overwrites existing app_id", url: "ws://host/path?app_id=old", appID: "new", want: "ws://host/path?app_id=new"},
		{name: "url-escapes the value", url: "ws://host/", appID: "a b&c", want: "ws://host/?app_id=a+b%26c"},
		{name: "invalid url returns error", url: "://", appID: "x", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := appendApplicationIDQueryParam(tc.url, tc.appID)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

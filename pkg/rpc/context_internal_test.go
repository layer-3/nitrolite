package rpc

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_Next(t *testing.T) {
	t.Parallel()

	t.Run("Missed", func(t *testing.T) {
		valueMap := make(map[string]bool)
		ctx := &Context{
			handlers: []Handler{
				func(c *Context) {
					valueMap["step1"] = true
				},
				func(c *Context) {
					valueMap["step2"] = true
				},
			},
		}
		ctx.Next()

		assert.True(t, valueMap["step1"], "First handler should have been executed")
		assert.False(t, valueMap["step2"], "Second handler should not have been executed")
		assert.Len(t, ctx.handlers, 1, "One handler should remain")
	})

	t.Run("Correct", func(t *testing.T) {
		valueMap := make(map[string]bool)
		ctx := &Context{
			handlers: []Handler{
				func(c *Context) {
					valueMap["step1"] = true
					c.Next()
				},
				func(c *Context) {
					valueMap["step2"] = true
				},
			},
		}
		ctx.Next()

		assert.True(t, valueMap["step1"], "First handler should have been executed")
		assert.True(t, valueMap["step2"], "Second handler should have been executed")
		assert.Len(t, ctx.handlers, 0, "No handlers should remain")
	})
}

func TestContext_Succeed(t *testing.T) {
	t.Parallel()

	ctx := &Context{
		Request: Message{
			Type:      MsgTypeReq,
			RequestID: 1,
		},
	}

	method := "method1"
	payload := Payload{
		"key": json.RawMessage(`"value"`),
	}
	ctx.Succeed(method, payload)

	assert.Equal(t, ctx.Request.RequestID, ctx.Response.RequestID, "Response RequestID should match Request RequestID")
	assert.Equal(t, method, ctx.Response.Method, "Response Method should match the expected method")
	assert.Equal(t, payload, ctx.Response.Payload, "Response Params should match the expected params")
}

func TestContext_Fail(t *testing.T) {
	t.Parallel()

	t.Run("With rpc.Error", func(t *testing.T) {
		ctx := &Context{
			Request: Message{
				Type:      MsgTypeReq,
				RequestID: 2,
			},
		}

		rpcErr := Errorf("RPC error occurred")
		ctx.Fail(rpcErr, "This message should be ignored")

		assert.Equal(t, ctx.Request.RequestID, ctx.Response.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, "RPC error occurred", ctx.Response.Error().Error(), "Response Message should match the rpc.Error message")
	})

	t.Run("With standard error and fallback message", func(t *testing.T) {
		ctx := &Context{
			Request: Message{
				Type:      MsgTypeReq,
				RequestID: 3,
			},
		}

		stdErr := assert.AnError
		fallbackMessage := "A standard error occurred"
		ctx.Fail(stdErr, fallbackMessage)

		assert.Equal(t, ctx.Request.RequestID, ctx.Response.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, fallbackMessage, ctx.Response.Error().Error(), "Response Message should match the fallback message")
	})

	t.Run("With nil error and fallback message", func(t *testing.T) {
		ctx := &Context{
			Request: Message{
				Type:      MsgTypeReq,
				RequestID: 4,
			},
		}

		fallbackMessage := "An error occurred"
		ctx.Fail(nil, fallbackMessage)

		assert.Equal(t, ctx.Request.RequestID, ctx.Response.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, fallbackMessage, ctx.Response.Error().Error(), "Response Message should match the fallback message")
	})

	t.Run("With nil error and empty fallback message", func(t *testing.T) {
		ctx := &Context{
			Request: Message{
				Type:      MsgTypeReq,
				RequestID: 5,
			},
		}

		ctx.Fail(nil, "")

		assert.Equal(t, ctx.Request.RequestID, ctx.Response.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, defaultNodeErrorMessage, ctx.Response.Error().Error(), "Response Message should match the default error message")
	})
}

func TestGetApplicationID(t *testing.T) {
	t.Parallel()

	t.Run("returns empty for nil context", func(t *testing.T) {
		assert.Equal(t, "", GetApplicationID(nil))
	})

	t.Run("returns empty for nil storage", func(t *testing.T) {
		assert.Equal(t, "", GetApplicationID(&Context{}))
	})

	t.Run("returns empty when unset", func(t *testing.T) {
		ctx := &Context{Storage: NewSafeStorage()}
		assert.Equal(t, "", GetApplicationID(ctx))
	})

	t.Run("returns value when set", func(t *testing.T) {
		ctx := &Context{Storage: NewSafeStorage()}
		ctx.Storage.Set(ApplicationIDQueryParam, "my-app")
		assert.Equal(t, "my-app", GetApplicationID(ctx))
	})

	t.Run("returns empty when value is not a string", func(t *testing.T) {
		ctx := &Context{Storage: NewSafeStorage()}
		ctx.Storage.Set(ApplicationIDQueryParam, 123)
		assert.Equal(t, "", GetApplicationID(ctx))
	})
}

func TestSafeStorage(t *testing.T) {
	t.Parallel()

	storage := NewSafeStorage()

	key := "testKey"
	value := "testValue"

	storage.Set(key, value)
	retrievedValue, ok := storage.Get(key)
	require.True(t, ok, "Key should exist in storage")
	require.Equal(t, value, retrievedValue, "Retrieved value should match the set value")

	// Test concurrent access
	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()

	go func() {
		defer wg.Done()

		for range 100 {
			storage.Set(key, value)
		}
	}()

	go func() {
		defer wg.Done()

		for range 100 {
			value, ok := storage.Get(key)
			assert.True(t, ok, "Key should exist in storage")
			assert.Equal(t, value, value, "Value should match the set value")
		}
	}()
}

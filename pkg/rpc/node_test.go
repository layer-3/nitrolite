package rpc

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/log"
)

func newTestNode(t *testing.T) *WebsocketNode {
	t.Helper()
	node, err := NewWebsocketNode(WebsocketNodeConfig{Logger: log.NewNoopLogger()})
	require.NoError(t, err)
	return node
}

func TestRegisteredMethods_EmptyNode(t *testing.T) {
	node := newTestNode(t)
	assert.Empty(t, node.RegisteredMethods())
}

func TestRegisteredMethods_RootHandle(t *testing.T) {
	node := newTestNode(t)
	noop := func(c *Context) {}

	node.Handle("ping", noop)
	node.Handle("status", noop)

	got := node.RegisteredMethods()
	sort.Strings(got)
	assert.Equal(t, []string{"ping", "status"}, got)
}

// Group-level Handle must also surface through RegisteredMethods(); the metrics
// seeder relies on a complete list across the root group and all subgroups.
func TestRegisteredMethods_IncludesGroupRegistrations(t *testing.T) {
	node := newTestNode(t)
	noop := func(c *Context) {}

	node.Handle("ping", noop)

	auth := node.NewGroup("auth")
	auth.Handle("auth.login", noop)
	auth.Handle("auth.logout", noop)

	nested := auth.NewGroup("admin")
	nested.Handle("auth.admin.kick", noop)

	got := node.RegisteredMethods()
	sort.Strings(got)
	assert.Equal(t,
		[]string{"auth.admin.kick", "auth.login", "auth.logout", "ping"},
		got,
	)
}

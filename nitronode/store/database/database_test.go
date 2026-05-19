package database

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresqlDbUrl(t *testing.T) {
	base := DatabaseConfig{
		Driver:   "postgres",
		Username: "user",
		Password: "pass",
		Host:     "db.example.com",
		Port:     "5432",
		Name:     "nitronode",
	}

	t.Run("DefaultsToRequireWhenSSLModeEmpty", func(t *testing.T) {
		dsn, err := postgresqlDbUrl(base)
		require.NoError(t, err)
		assert.Contains(t, dsn, "sslmode=require")
		assert.NotContains(t, dsn, "sslmode=disable")
	})

	t.Run("HonorsExplicitSSLMode", func(t *testing.T) {
		cnf := base
		cnf.SSLMode = "verify-full"
		dsn, err := postgresqlDbUrl(cnf)
		require.NoError(t, err)
		assert.Contains(t, dsn, "sslmode=verify-full")
	})

	t.Run("AllowsDisableForLocalDev", func(t *testing.T) {
		cnf := base
		cnf.SSLMode = "disable"
		dsn, err := postgresqlDbUrl(cnf)
		require.NoError(t, err)
		assert.Contains(t, dsn, "sslmode=disable")
	})

	t.Run("RejectsInvalidSSLMode", func(t *testing.T) {
		cnf := base
		cnf.SSLMode = "totally-bogus"
		_, err := postgresqlDbUrl(cnf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sslmode")
	})

	t.Run("AppendsSearchPathWhenSchemaSet", func(t *testing.T) {
		cnf := base
		cnf.Schema = "tenant_a"
		dsn, err := postgresqlDbUrl(cnf)
		require.NoError(t, err)
		assert.Contains(t, dsn, "search_path=tenant_a")
	})

	t.Run("URLOverridesIndividualFields", func(t *testing.T) {
		cnf := base
		cnf.URL = "postgres://override:secret@otherhost:6543/otherdb?sslmode=verify-ca"
		cnf.SSLMode = "disable" // ignored when URL set
		dsn, err := postgresqlDbUrl(cnf)
		require.NoError(t, err)
		assert.Equal(t, cnf.URL, dsn)
		assert.False(t, strings.Contains(dsn, "user=user"), "URL must be returned verbatim")
	})

	t.Run("RejectsUnsupportedDriver", func(t *testing.T) {
		cnf := base
		cnf.Driver = "mysql"
		_, err := postgresqlDbUrl(cnf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported driver")
	})
}

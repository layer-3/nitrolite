package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculatePaginationMetadata_ZeroLimit(t *testing.T) {
	_, err := calculatePaginationMetadata(100, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "limit must be greater than 0")
}

func TestCalculatePaginationMetadata_Normal(t *testing.T) {
	meta, err := calculatePaginationMetadata(25, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, uint32(3), meta.PageCount)
	assert.Equal(t, uint32(1), meta.Page)
	assert.Equal(t, uint32(10), meta.PerPage)
	assert.Equal(t, uint32(25), meta.TotalCount)
}

func TestCalculatePaginationMetadata_ExactPage(t *testing.T) {
	meta, err := calculatePaginationMetadata(20, 10, 10)
	require.NoError(t, err)
	assert.Equal(t, uint32(2), meta.PageCount)
	assert.Equal(t, uint32(2), meta.Page)
}

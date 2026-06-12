package repository_common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPasswordAndCompare(t *testing.T) {
	hash, err := HashPassword("12345678")

	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, "12345678", hash)
	require.True(t, ComparePasswordHash("12345678", hash))
	require.False(t, ComparePasswordHash("wrong-password", hash))
}

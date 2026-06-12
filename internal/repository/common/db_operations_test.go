package repository_common

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestIsPgError(t *testing.T) {
	err := &pgconn.PgError{Code: "23505"}

	require.True(t, IsPgError(err, "23505"))
	require.False(t, IsPgError(err, "23503"))
	require.False(t, IsPgError(errors.New("regular error"), "23505"))
}

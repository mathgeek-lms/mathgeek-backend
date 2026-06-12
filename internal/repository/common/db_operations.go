package repository_common

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

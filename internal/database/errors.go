package database

import (
	"errors"

	"github.com/lib/pq"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

// IsUniqueViolation reports whether err is a Postgres unique constraint violation.
func IsUniqueViolation(err error) bool {
	return hasPgCode(err, pgUniqueViolation)
}

// IsForeignKeyViolation reports whether err is a Postgres foreign key violation.
func IsForeignKeyViolation(err error) bool {
	return hasPgCode(err, pgForeignKeyViolation)
}

func hasPgCode(err error, code string) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == code
}

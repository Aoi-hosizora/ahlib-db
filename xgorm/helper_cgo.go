//go:build cgo
// +build cgo

package xgorm

import (
	"github.com/mattn/go-sqlite3"
)

// IsSQLiteUniqueConstraintError checks whether err is SQLite's ErrConstraintUnique error, whose extended code is SQLiteUniqueConstraintErrno.
func IsSQLiteUniqueConstraintError(err error) bool {
	e, ok := err.(sqlite3.Error)
	if ok {
		return int(e.ExtendedCode) == SQLiteUniqueConstraintErrno
	}
	pe, ok := err.(*sqlite3.Error)
	return ok && int(pe.ExtendedCode) == SQLiteUniqueConstraintErrno
}

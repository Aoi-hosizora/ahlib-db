//go:build cgo
// +build cgo

package xgormv2

import (
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// IsSQLiteUniqueConstraintError checks if err is SQLite's ErrConstraintUnique error, its error extended code is SQLiteUniqueConstraintErrno.
func IsSQLiteUniqueConstraintError(err error) bool {
	e, ok := err.(sqlite3.Error)
	if ok {
		return e.ExtendedCode == SQLiteUniqueConstraintErrno
	}
	pe, ok := err.(*sqlite3.Error)
	return ok && pe.ExtendedCode == SQLiteUniqueConstraintErrno
}

// CreateErr checks gorm.DB after create operated, will only return xstatus.DbSuccess, xstatus.DbExisted and xstatus.DbFailed.
func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error),
		IsSQLite(rdb) && IsSQLiteUniqueConstraintError(rdb.Error),
		IsPostgreSQL(rdb) && IsPostgreSQLUniqueViolationError != nil && IsPostgreSQLUniqueViolationError(rdb.Error):
		return xstatus.DbExisted, rdb.Error // duplicate
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	}
	return xstatus.DbSuccess, nil
}

// UpdateErr checks gorm.DB after update operated, will only return xstatus.DbSuccess, xstatus.DbNotFound, xstatus.DbExisted and xstatus.DbFailed.
func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error),
		IsSQLite(rdb) && IsSQLiteUniqueConstraintError(rdb.Error),
		IsPostgreSQL(rdb) && IsPostgreSQLUniqueViolationError != nil && IsPostgreSQLUniqueViolationError(rdb.Error):
		return xstatus.DbExisted, rdb.Error // duplicate
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	case rdb.RowsAffected == 0:
		return xstatus.DbNotFound, nil // not found
	}
	return xstatus.DbSuccess, nil
}

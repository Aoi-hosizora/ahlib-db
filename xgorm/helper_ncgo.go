// +build !cgo

package xgorm

import (
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/jinzhu/gorm"
)

// CreateErr checks gorm.DB create result, will only return xstatus.DbExisted, xstatus.DbFailed and xstatus.DbSuccess.
func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error),
		IsPostgreSQL(rdb) && IsPostgreSQLUniqueViolationError(rdb.Error):
		return xstatus.DbExisted, rdb.Error // duplicate
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	}
	return xstatus.DbSuccess, nil
}

// UpdateErr checks gorm.DB update result, will only return xstatus.DbExisted, xstatus.DbFailed, xstatus.DbNotFound and xstatus.DbSuccess.
func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error),
		IsPostgreSQL(rdb) && IsPostgreSQLUniqueViolationError(rdb.Error):
		return xstatus.DbExisted, rdb.Error // duplicate
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	case rdb.RowsAffected == 0:
		return xstatus.DbNotFound, nil // not found
	}
	return xstatus.DbSuccess, nil
}

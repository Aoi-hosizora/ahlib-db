//go:build cgo
// +build cgo

package xgorm

import (
	"github.com/Aoi-hosizora/ahlib/xtesting"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres" // dummy
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mattn/go-sqlite3"
	"testing"
)

func TestMess2(t *testing.T) {
	xtesting.True(t, IsSQLiteUniqueConstraintError(sqlite3.Error{ExtendedCode: sqlite3.ErrNoExtended(SQLiteUniqueConstraintErrno)}))
	xtesting.True(t, IsSQLiteUniqueConstraintError(&sqlite3.Error{ExtendedCode: sqlite3.ErrNoExtended(SQLiteUniqueConstraintErrno)}))
}

func TestHook(t *testing.T) {
	for _, tc := range []struct {
		giveDialect string
		giveParam   string
	}{
		{MySQL, mysqlDsn},
		{SQLite, sqliteFile},
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			testHook(t, tc.giveDialect, tc.giveParam)
		})
	}
}

func TestHelper(t *testing.T) {
	for _, tc := range []struct {
		giveDialect string
		giveParam   string
	}{
		{MySQL, mysqlDsn},
		{SQLite, sqliteFile},
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			testHelper(t, tc.giveDialect, tc.giveParam)
		})
	}
}

func TestLogger(t *testing.T) {
	for _, tc := range []struct {
		giveDialect string
		giveParam   string
	}{
		{MySQL, mysqlDsn},
		{SQLite, sqliteFile},
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			testLogger(t, tc.giveDialect, tc.giveParam)
		})
	}
}

//go:build cgo
// +build cgo

package xgormv2

import (
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/mattn/go-sqlite3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
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
			switch tc.giveDialect {
			case MySQL:
				testHook(t, mysql.Open(tc.giveParam))
			case SQLite:
				testHook(t, sqlite.Open(tc.giveParam))
			}
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
			switch tc.giveDialect {
			case MySQL:
				testHelper(t, mysql.Open(tc.giveParam))
			case SQLite:
				testHelper(t, sqlite.Open(tc.giveParam))
			}
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
			switch tc.giveDialect {
			case MySQL:
				testLogger(t, mysql.Open(tc.giveParam))
			case SQLite:
				testLogger(t, sqlite.Open(tc.giveParam))
			}
		})
	}
}

//go:build !cgo
// +build !cgo

package xgorm

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"testing"
)

func TestHook(t *testing.T) {
	for _, tc := range []struct {
		giveDialect string
		giveParam   string
	}{
		{MySQL, mysqlDsn},
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
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			testLogger(t, tc.giveDialect, tc.giveParam)
		})
	}
}

//go:build !cgo
// +build !cgo

package xgormv2

import (
	"gorm.io/driver/mysql"
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
			switch tc.giveDialect {
			case MySQL:
				testHook(t, mysql.Open(tc.giveParam))
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
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			switch tc.giveDialect {
			case MySQL:
				testHelper(t, mysql.Open(tc.giveParam))
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
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			switch tc.giveDialect {
			case MySQL:
				testLogger(t, mysql.Open(tc.giveParam))
			}
		})
	}
}

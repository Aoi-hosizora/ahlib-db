package xdbutils_sqlite

import (
	"errors"
	"github.com/Aoi-hosizora/ahlib-db/xdbutils/internal"
	"testing"
)

type stringError string
type structError struct{ msg string }
type fakeSQLiteError struct{ ExtendedCode ErrNoExtended }

func (n stringError) Error() string     { return string(n) }
func (n structError) Error() string     { return n.msg }
func (f fakeSQLiteError) Error() string { return "x" }

func TestCheckSQLiteErrorExtendedCodeByReflect(t *testing.T) {
	for _, tc := range []struct {
		giveError error
		giveCode  int
		want      bool
	}{
		{nil, 0, false},
		{errors.New("x"), 0, false},
		{stringError("x"), 0, false},
		{(*structError)(nil), 0, false},
		{structError{"x"}, 0, false},
		{&structError{"x"}, 0, false},
		{fakeSQLiteError{0}, 0, true},
		{&fakeSQLiteError{0}, 0, true},
		{fakeSQLiteError{ErrConstraintUnique}, 0, false},
		{&fakeSQLiteError{ErrConstraintUnique}, 0, false},
		{fakeSQLiteError{ErrConstraintUnique}, int(ErrConstraintUnique), true},
		{&fakeSQLiteError{ErrConstraintUnique}, int(ErrConstraintUnique), true},
	} {
		internal.TestEqual(t, CheckSQLiteErrorExtendedCodeByReflect(tc.giveError, tc.giveCode), tc.want)
	}
}

func TestErrNo(t *testing.T) {
	for _, tc := range []struct {
		giveExtend ErrNoExtended
		wantErrNo  ErrNo
		wantBy     int
	}{
		// ErrIoErr = ErrNo(10)
		{ErrIoErrRead, ErrIoErr, 1},
		{ErrIoErrShortRead, ErrIoErr, 2},
		{ErrIoErrWrite, ErrIoErr, 3},
		{ErrIoErrFsync, ErrIoErr, 4},
		{ErrIoErrDirFsync, ErrIoErr, 5},
		{ErrIoErrTruncate, ErrIoErr, 6},
		{ErrIoErrFstat, ErrIoErr, 7},
		{ErrIoErrUnlock, ErrIoErr, 8},
		{ErrIoErrRDlock, ErrIoErr, 9},
		{ErrIoErrDelete, ErrIoErr, 10},
		{ErrIoErrBlocked, ErrIoErr, 11},
		{ErrIoErrNoMem, ErrIoErr, 12},
		{ErrIoErrAccess, ErrIoErr, 13},
		{ErrIoErrCheckReservedLock, ErrIoErr, 14},
		{ErrIoErrLock, ErrIoErr, 15},
		{ErrIoErrClose, ErrIoErr, 16},
		{ErrIoErrDirClose, ErrIoErr, 17},
		{ErrIoErrSHMOpen, ErrIoErr, 18},
		{ErrIoErrSHMSize, ErrIoErr, 19},
		{ErrIoErrSHMLock, ErrIoErr, 20},
		{ErrIoErrSHMMap, ErrIoErr, 21},
		{ErrIoErrSeek, ErrIoErr, 22},
		{ErrIoErrDeleteNoent, ErrIoErr, 23},
		{ErrIoErrMMap, ErrIoErr, 24},
		{ErrIoErrGetTempPath, ErrIoErr, 25},
		{ErrIoErrConvPath, ErrIoErr, 26},

		// ErrConstraint = ErrNo(19)
		{ErrConstraintCheck, ErrConstraint, 1},
		{ErrConstraintCommitHook, ErrConstraint, 2},
		{ErrConstraintForeignKey, ErrConstraint, 3},
		{ErrConstraintFunction, ErrConstraint, 4},
		{ErrConstraintNotNull, ErrConstraint, 5},
		{ErrConstraintPrimaryKey, ErrConstraint, 6},
		{ErrConstraintTrigger, ErrConstraint, 7},
		{ErrConstraintUnique, ErrConstraint, 8},
		{ErrConstraintVTab, ErrConstraint, 9},
		{ErrConstraintRowID, ErrConstraint, 10},
	} {
		internal.TestEqual(t, int(tc.giveExtend), int(tc.wantErrNo.Extend(tc.wantBy)))
	}
}

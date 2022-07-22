package xdbutils_sqlite

import (
	"reflect"
)

// CheckSQLiteErrorExtendedCodeByReflect checks whether given err can be regarded as SQLite error (github.com/mattn/go-sqlite3), and whether its extended code is given code.
func CheckSQLiteErrorExtendedCodeByReflect(err error, code int) bool {
	val := reflect.ValueOf(err)
	if !val.IsValid() {
		return false // nil error
	}
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return false // nil pointer
		}
		// *sqlite3.Error -> sqlite3.Error
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false // non-struct and non-struct-pointer
	}

	field := val.FieldByName("ExtendedCode")
	if field.IsValid() && field.Kind() == reflect.Int {
		extendedCode := field.Interface().(int)
		return extendedCode == code
	}
	return false
}

// The following code are almost referred from https://github.com/mattn/go-sqlite3/blob/master/error.go.

// ErrNo represents sqlite3.Error's Error field, inherits errno.
type ErrNo int

// ErrNoExtended represents sqlite3.Error's ExtendedCode field, is extended errno.
type ErrNoExtended int

// Extend returns extended errno.
func (err ErrNo) Extend(by int) ErrNoExtended {
	return ErrNoExtended(int(err) | (by << 8))
}

// result codes from http://www.sqlite.org/c3ref/c_abort.html
const (
	ErrError      = ErrNo(1)  /* SQL error or missing database */
	ErrInternal   = ErrNo(2)  /* Internal logic error in SQLite */
	ErrPerm       = ErrNo(3)  /* Access permission denied */
	ErrAbort      = ErrNo(4)  /* Callback routine requested an abort */
	ErrBusy       = ErrNo(5)  /* The database file is locked */
	ErrLocked     = ErrNo(6)  /* A table in the database is locked */
	ErrNomem      = ErrNo(7)  /* A malloc() failed */
	ErrReadonly   = ErrNo(8)  /* Attempt to write a readonly database */
	ErrInterrupt  = ErrNo(9)  /* Operation terminated by sqlite3_interrupt() */
	ErrIoErr      = ErrNo(10) /* Some kind of disk I/O error occurred */
	ErrCorrupt    = ErrNo(11) /* The database disk image is malformed */
	ErrNotFound   = ErrNo(12) /* Unknown opcode in sqlite3_file_control() */
	ErrFull       = ErrNo(13) /* Insertion failed because database is full */
	ErrCantOpen   = ErrNo(14) /* Unable to open the database file */
	ErrProtocol   = ErrNo(15) /* Database lock protocol error */
	ErrEmpty      = ErrNo(16) /* Database is empty */
	ErrSchema     = ErrNo(17) /* The database schema changed */
	ErrTooBig     = ErrNo(18) /* String or BLOB exceeds size limit */
	ErrConstraint = ErrNo(19) /* Abort due to constraint violation */
	ErrMismatch   = ErrNo(20) /* Data type mismatch */
	ErrMisuse     = ErrNo(21) /* Library used incorrectly */
	ErrNoLFS      = ErrNo(22) /* Uses OS features not supported on host */
	ErrAuth       = ErrNo(23) /* Authorization denied */
	ErrFormat     = ErrNo(24) /* Auxiliary database format error */
	ErrRange      = ErrNo(25) /* 2nd parameter to sqlite3_bind out of range */
	ErrNotADB     = ErrNo(26) /* File opened that is not a database file */
	ErrNotice     = ErrNo(27) /* Notifications from sqlite3_log() */
	ErrWarning    = ErrNo(28) /* Warnings from sqlite3_log() */
)

// result codes from http://www.sqlite.org/c3ref/c_abort_rollback.html
const (
	ErrIoErrRead              = ErrNoExtended(10 | 1<<8)  // ErrIoErr.Extend(1) => ErrIoErr = ErrNo(10)
	ErrIoErrShortRead         = ErrNoExtended(10 | 2<<8)  // ErrIoErr.Extend(2)
	ErrIoErrWrite             = ErrNoExtended(10 | 3<<8)  // ErrIoErr.Extend(3)
	ErrIoErrFsync             = ErrNoExtended(10 | 4<<8)  // ErrIoErr.Extend(4)
	ErrIoErrDirFsync          = ErrNoExtended(10 | 5<<8)  // ErrIoErr.Extend(5)
	ErrIoErrTruncate          = ErrNoExtended(10 | 6<<8)  // ErrIoErr.Extend(6)
	ErrIoErrFstat             = ErrNoExtended(10 | 7<<8)  // ErrIoErr.Extend(7)
	ErrIoErrUnlock            = ErrNoExtended(10 | 8<<8)  // ErrIoErr.Extend(8)
	ErrIoErrRDlock            = ErrNoExtended(10 | 9<<8)  // ErrIoErr.Extend(9)
	ErrIoErrDelete            = ErrNoExtended(10 | 10<<8) // ErrIoErr.Extend(10)
	ErrIoErrBlocked           = ErrNoExtended(10 | 11<<8) // ErrIoErr.Extend(11)
	ErrIoErrNoMem             = ErrNoExtended(10 | 12<<8) // ErrIoErr.Extend(12)
	ErrIoErrAccess            = ErrNoExtended(10 | 13<<8) // ErrIoErr.Extend(13)
	ErrIoErrCheckReservedLock = ErrNoExtended(10 | 14<<8) // ErrIoErr.Extend(14)
	ErrIoErrLock              = ErrNoExtended(10 | 15<<8) // ErrIoErr.Extend(15)
	ErrIoErrClose             = ErrNoExtended(10 | 16<<8) // ErrIoErr.Extend(16)
	ErrIoErrDirClose          = ErrNoExtended(10 | 17<<8) // ErrIoErr.Extend(17)
	ErrIoErrSHMOpen           = ErrNoExtended(10 | 18<<8) // ErrIoErr.Extend(18)
	ErrIoErrSHMSize           = ErrNoExtended(10 | 19<<8) // ErrIoErr.Extend(19)
	ErrIoErrSHMLock           = ErrNoExtended(10 | 20<<8) // ErrIoErr.Extend(20)
	ErrIoErrSHMMap            = ErrNoExtended(10 | 21<<8) // ErrIoErr.Extend(21)
	ErrIoErrSeek              = ErrNoExtended(10 | 22<<8) // ErrIoErr.Extend(22)
	ErrIoErrDeleteNoent       = ErrNoExtended(10 | 23<<8) // ErrIoErr.Extend(23)
	ErrIoErrMMap              = ErrNoExtended(10 | 24<<8) // ErrIoErr.Extend(24)
	ErrIoErrGetTempPath       = ErrNoExtended(10 | 25<<8) // ErrIoErr.Extend(25)
	ErrIoErrConvPath          = ErrNoExtended(10 | 26<<8) // ErrIoErr.Extend(26)
	ErrLockedSharedCache      = ErrNoExtended(6 | 1<<8)   // ErrLocked.Extend(1) => ErrLocked = ErrNo(6)
	ErrBusyRecovery           = ErrNoExtended(5 | 1<<8)   // ErrBusy.Extend(1) => ErrBusy = ErrNo(5)
	ErrBusySnapshot           = ErrNoExtended(5 | 2<<8)   // ErrBusy.Extend(2)
	ErrCantOpenNoTempDir      = ErrNoExtended(14 | 1<<8)  // ErrCantOpen.Extend(1) => ErrCantOpen = ErrNo(14)
	ErrCantOpenIsDir          = ErrNoExtended(14 | 2<<8)  // ErrCantOpen.Extend(2)
	ErrCantOpenFullPath       = ErrNoExtended(14 | 3<<8)  // ErrCantOpen.Extend(3)
	ErrCantOpenConvPath       = ErrNoExtended(14 | 4<<8)  // ErrCantOpen.Extend(4)
	ErrCorruptVTab            = ErrNoExtended(11 | 1<<8)  // ErrCorrupt.Extend(1) => ErrCorrupt = ErrNo(11)
	ErrReadonlyRecovery       = ErrNoExtended(8 | 1<<8)   // ErrReadonly.Extend(1) => ErrReadonly = ErrNo(8)
	ErrReadonlyCantLock       = ErrNoExtended(8 | 2<<8)   // ErrReadonly.Extend(2)
	ErrReadonlyRollback       = ErrNoExtended(8 | 3<<8)   // ErrReadonly.Extend(3)
	ErrReadonlyDbMoved        = ErrNoExtended(8 | 4<<8)   // ErrReadonly.Extend(4)
	ErrAbortRollback          = ErrNoExtended(4 | 2<<8)   // ErrAbort.Extend(2) => ErrAbort = ErrNo(4)
	ErrConstraintCheck        = ErrNoExtended(19 | 1<<8)  // ErrConstraint.Extend(1) => ErrConstraint = ErrNo(19)
	ErrConstraintCommitHook   = ErrNoExtended(19 | 2<<8)  // ErrConstraint.Extend(2)
	ErrConstraintForeignKey   = ErrNoExtended(19 | 3<<8)  // ErrConstraint.Extend(3)
	ErrConstraintFunction     = ErrNoExtended(19 | 4<<8)  // ErrConstraint.Extend(4)
	ErrConstraintNotNull      = ErrNoExtended(19 | 5<<8)  // ErrConstraint.Extend(5)
	ErrConstraintPrimaryKey   = ErrNoExtended(19 | 6<<8)  // ErrConstraint.Extend(6)
	ErrConstraintTrigger      = ErrNoExtended(19 | 7<<8)  // ErrConstraint.Extend(7)
	ErrConstraintUnique       = ErrNoExtended(19 | 8<<8)  // ErrConstraint.Extend(8)
	ErrConstraintVTab         = ErrNoExtended(19 | 9<<8)  // ErrConstraint.Extend(9)
	ErrConstraintRowID        = ErrNoExtended(19 | 10<<8) // ErrConstraint.Extend(10)
	ErrNoticeRecoverWAL       = ErrNoExtended(27 | 1<<8)  // ErrNotice.Extend(1) => ErrNotice = ErrNo(27)
	ErrNoticeRecoverRollback  = ErrNoExtended(27 | 2<<8)  // ErrNotice.Extend(2)
	ErrWarningAutoIndex       = ErrNoExtended(28 | 1<<8)  // ErrWarning.Extend(1) => ErrWarning = ErrNo(28)
)

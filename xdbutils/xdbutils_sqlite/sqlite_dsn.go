package xdbutils_sqlite

import (
	"fmt"
	"net/url"
	"strconv"
)

// SQLiteConfig is a configuration for SQLite, can be used to generate DSN by FormatDSN method. Please visit
// https://github.com/mattn/go-sqlite3#connection-string and https://www.sqlite.org/c3ref/open.html for more information.
type SQLiteConfig struct {
	// SQLite specified parameters
	Filename  string // Database filename (<filepath>, ":memory:", empty string)
	VFS       string // The name of a VFS object
	Mode      string // Access Mode (ro, rw, rwc, memory)
	Cache     string // Shared-Cache Mode (shared, private)
	Psow      *bool  // ...
	NoLock    *bool  // ...
	Immutable *bool  // ...

	// go-sqlite3 package parameters
	Auth                   bool   // UA - Create, defaults to false
	AuthUser               string // UA - Username
	AuthPass               string // UA - Password
	AuthCrypt              string // UA - Crypt (SHA1, SSHA1, SHA256, SSHA256, SHA384, SSHA384, SHA512, SSHA512)
	AuthSalt               string // UA - Salt
	AutoVacuum             string // Auto Vacuum (none, full, incremental), defaults to ignore this option
	BusyTimeout            *int   // Busy Timeout, default to 5000
	CaseSensitiveLike      *bool  // Case Sensitive LIKE, defaults to ignore this option
	DeferForeignKeys       *bool  // Defer Foreign Keys, defaults to ignore this option
	ForeignKeys            *bool  // Foreign Keys, defaults to ignore this option
	IgnoreCheckConstraints *bool  // Ignore CHECK Constraints, defaults to ignore this option
	JournalMode            string // Journal Mode (DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF), defaults to ignore this option
	LockingMode            string // Locking Mode (NORMAL, EXCLUSIVE), defaults to NORMAL
	Mutex                  string // Mutex Locking (no, full), defaults to full
	QueryOnly              *bool  // Query Only, defaults to ignore this option
	RecursiveTriggers      *bool  // Recursive Triggers, defaults to ignore this option
	SecureDelete           string // Secure Delete (default, yes, no, fast), defaults to default
	Synchronous            string // Synchronous (OFF, NORMAL, FULL, EXTRA), defaults to normal
	Loc                    string // Time Zone Location (auto, ...), defaults to ignore this option
	TransactionLock        string // Transaction Lock (immediate, deferred, exclusive), defaults to deferred
	WritableSchema         *bool  // Writable Schema, defaults to ignore this option
	CacheSize              *int   // Cache Size, defaults to ignore this option
}

// FormatDSN generates formatted DSN string from SQLiteConfig.
func (s *SQLiteConfig) FormatDSN() string {
	values := url.Values{}

	if s.Auth {
		values.Set("_auth", "")
	}
	if s.AuthUser != "" {
		values.Set("_auth_user", s.AuthUser)
	}
	if s.AuthPass != "" {
		values.Set("_auth_pass", s.AuthPass)
	}
	if s.AuthCrypt != "" {
		values.Set("_auth_crypt", s.AuthCrypt)
	}
	if s.AuthSalt != "" {
		values.Set("_auth_salt", s.AuthSalt)
	}
	if s.AutoVacuum != "" {
		values.Set("_auto_vacuum", s.AutoVacuum)
	}
	if s.BusyTimeout != nil {
		values.Set("_busy_timeout", strconv.Itoa(*s.BusyTimeout))
	}
	if s.CaseSensitiveLike != nil {
		values.Set("_case_sensitive_like", boolString(*s.CaseSensitiveLike))
	}
	if s.DeferForeignKeys != nil {
		values.Set("_defer_foreign_keys", boolString(*s.DeferForeignKeys))
	}
	if s.ForeignKeys != nil {
		values.Set("_foreign_keys", boolString(*s.ForeignKeys))
	}
	if s.IgnoreCheckConstraints != nil {
		values.Set("_ignore_check_constraints", boolString(*s.IgnoreCheckConstraints))
	}
	if s.JournalMode != "" {
		values.Set("_journal_mode", s.JournalMode)
	}
	if s.LockingMode != "" {
		values.Set("_locking_mode", s.LockingMode)
	}
	if s.Mutex != "" {
		values.Set("_mutex", s.Mutex)
	}
	if s.QueryOnly != nil {
		values.Set("_query_only", boolString(*s.QueryOnly))
	}
	if s.RecursiveTriggers != nil {
		values.Set("_recursive_triggers", boolString(*s.RecursiveTriggers))
	}
	if s.SecureDelete != "" {
		values.Set("_secure_delete", s.SecureDelete)
	}
	if s.Synchronous != "" {
		values.Set("_synchronous", s.Synchronous)
	}
	if s.Loc != "" {
		values.Set("_loc", s.Loc)
	}
	if s.TransactionLock != "" {
		values.Set("_txlock", s.TransactionLock)
	}
	if s.WritableSchema != nil {
		values.Set("_writable_schema", boolString(*s.WritableSchema))
	}
	if s.CacheSize != nil {
		values.Set("_busy_timeout", strconv.Itoa(*s.CacheSize))
	}

	if s.VFS != "" {
		values.Set("vfs", s.VFS)
	}
	if s.Mode != "" {
		values.Set("mode", s.Mode)
	}
	if s.Cache != "" {
		values.Set("cache", s.Cache)
	}
	if s.Psow != nil {
		values.Set("psow", boolString(*s.Psow))
	}
	if s.NoLock != nil {
		values.Set("nolock", boolString(*s.NoLock))
	}
	if s.Immutable != nil {
		values.Set("immutable", boolString(*s.Immutable))
	}

	query := values.Encode()
	if query == "" {
		// 1. If the filename is ":memory:", then a private, temporary in-memory database is created for the connection.
		//    This in-memory database will vanish when the database connection is closed. It is recommended that when a
		//    database filename actually does begin with a ":" character you should prefix the filename with a pathname
		//    such as "./" to avoid ambiguity.
		// 2. If the filename is an empty string, then a private, temporary on-disk database will be created. This private
		//    database will be automatically deleted as soon as the database connection is closed.
		return url.QueryEscape(s.Filename)
	}

	// 3. If URI filename interpretation is enabled, and the filename argument begins with "file:", then the filename is
	//    interpreted as a URI (URI Filenames). The query component of a URI may contain parameters that are interpreted
	//    either by SQLite itself, or by a custom VFS implementation. SQLite and its built-in VFSes interpret the following
	//    query parameters: vfs, mode, cache, psow, nolock, immutable.s
	return fmt.Sprintf("file:%s?%s", url.QueryEscape(s.Filename), query)
}

// boolString returns string from bool value, in "0" and "1" format.
func boolString(b bool) string {
	if b {
		return "1" // or true
	}
	return "0" // or false
}

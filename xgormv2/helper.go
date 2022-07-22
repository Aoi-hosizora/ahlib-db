package xgormv2

import (
	"errors"
	"fmt"
	"github.com/Aoi-hosizora/ahlib-db/xdbutils"
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// ========
// dialects
// ========

const (
	// MySQL is MySQL dialect for gorm, remember to use https://github.com/go-gorm/mysql to open a mysql gorm.DB.
	MySQL = "mysql"

	// SQLite is SQLite dialect for gorm, remember to use https://github.com/go-gorm/sqlite to open a sqlite gorm.DB.
	SQLite = "sqlite"

	// Postgres is PostgreSQL dialect for gorm, remember to use https://github.com/go-gorm/postgres to open a postgres gorm.DB.
	Postgres = "postgres"
)

// IsMySQL checks whether the dialect of given gorm.DB is "mysql".
func IsMySQL(db *gorm.DB) bool {
	return db.Dialector.Name() == MySQL
}

// IsSQLite checks whether the dialect of given gorm.DB is "sqlite".
func IsSQLite(db *gorm.DB) bool {
	return db.Dialector.Name() == SQLite
}

// IsPostgreSQL checks whether the dialect of given gorm.DB is "postgres".
func IsPostgreSQL(db *gorm.DB) bool {
	return db.Dialector.Name() == Postgres
}

// MySQLConfig is an alias type of mysql.Config, can be used to generate dsl by FormatDSN method.
type MySQLConfig = mysql.Config

// MySQLDefaultCharsetTimeLocParam returns a map as mysql.Config's Param value, it contains the default "utf8mb4" charset, "True" parseTime, and "Local" loc.
func MySQLDefaultCharsetTimeLocParam() map[string]string {
	return map[string]string{"charset": "utf8mb4", "parseTime": "True", "loc": "Local"}
}

// MySQLDefaultDsn returns the MySQL dsn from given parameters with "utf8mb4" charset and "local" location. If you want to set more options in dsn,
// please use mysql.Config or xgormv2.MySQLConfig. For more information, please visit https://github.com/go-sql-driver/mysql#dsn-data-source-name.
func MySQLDefaultDsn(username, password, address, database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, address, database)
}

// SQLiteDefaultDsn returns the SQLite dsn from given username. For more information, please visit https://github.com/mattn/go-sqlite3#connection-string.
func SQLiteDefaultDsn(filename string) string {
	return filename // fmt.Sprintf("file:%s", filename)
}

// PostgresDefaultDsn returns the Postgres dsn from given parameters. For more information, please visit
// https://www.postgresql.org/docs/current/libpq-connect.html#id-1.7.3.8.3.3.
func PostgresDefaultDsn(username, password, host string, port int, database string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", host, port, username, password, database)
}

const (
	// MySQLDuplicateEntryErrno is MySQL's ER_DUP_ENTRY errno, referred from https://github.com/VividCortex/mysqlerr/blob/master/mysqlerr.go and
	// https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.htm.
	MySQLDuplicateEntryErrno = mysqlerr.ER_DUP_ENTRY // 1062

	// SQLiteUniqueConstraintErrno is SQLite's CONSTRAINT_UNIQUE extended errno, referred from https://github.com/mattn/go-sqlite3/blob/master/error.go
	// and http://www.sqlite.org/c3ref/c_abort_rollback.html.
	SQLiteUniqueConstraintErrno = 19 | 8<<8

	// PostgreSQLUniqueViolationErrno is PostgreSQL's unique_violation errno, referred from https://github.com/lib/pq/blob/master/error.go and
	// https://www.postgresql.org/docs/10/errcodes-appendix.html
	PostgreSQLUniqueViolationErrno = "23505"
)

// IsMySQLDuplicateEntryError checks whether err is MySQL's ER_DUP_ENTRY error, whose error code is MySQLDuplicateEntryErrno.
func IsMySQLDuplicateEntryError(err error) bool {
	e, ok := err.(*mysql.MySQLError)
	return ok && e.Number == MySQLDuplicateEntryErrno
}

// IsPostgreSQLUniqueViolationError is a variable that used to check whether err is PostgreSQL's unique_violation error, whose error code is PostgreSQLUniqueViolationErrno.
//
// Example:
// 	import "github.com/jackc/pgconn"
// 	xgormv2.IsPostgreSQLUniqueViolationError = func(err error) bool {
// 		perr, ok := err.(*pgconn.PgError)
// 		return ok && perr.Code == xgormv2.PostgreSQLUniqueViolationErrno
// 	}
var IsPostgreSQLUniqueViolationError func(err error) bool

// ==============
// CRUD and other
// ==============

// IsRecordNotFound checks whether given error from gorm.DB is gorm.ErrRecordNotFound.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// QueryErr checks gorm.DB after query operated, will only return xstatus.DbSuccess, xstatus.DbNotFound and xstatus.DbFailed.
func QueryErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case IsRecordNotFound(rdb.Error):
		return xstatus.DbNotFound, nil // not found
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	}
	return xstatus.DbSuccess, nil
}

// DeleteErr checks gorm.DB after delete operated, will only return xstatus.DbSuccess, xstatus.DbNotFound and xstatus.DbFailed.
func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	case rdb.RowsAffected == 0:
		return xstatus.DbNotFound, nil // not found
	}
	return xstatus.DbSuccess, nil
}

// PropertyValue is a struct type of database entity's property mapping rule, used in GenerateOrderByExpr.
type PropertyValue = xdbutils.PropertyValue

// PropertyDict is a dictionary type to store pairs from data transfer object to database entity's PropertyValue, used in GenerateOrderByExpr.
type PropertyDict = xdbutils.PropertyDict

// NewPropertyValue creates a PropertyValue by given reverse and destinations, used to describe database entity's property mapping rule.
//
// Here:
// 1. `destinations` represents mapping property destination array, use `property_name` directly for sql, use `returned_name.property_name` for cypher.
// 2. `reverse` represents the flag whether you need to revert the order or not.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	return xdbutils.NewPropertyValue(reverse, destinations...)
}

// GenerateOrderByExpr returns a generated order-by expression by given source (query string) order string (such as "name desc, age asc") and PropertyDict.
// The generated expression is in mysql-sql or neo4j-cypher style (such as "xxx ASC" or "xxx.yyy DESC").
//
// Example:
// 	dict := PropertyDict{
// 		"uid":  NewPropertyValue(false, "uid"),
// 		"name": NewPropertyValue(false, "firstname", "lastname"),
// 		"age":  NewPropertyValue(true, "birthday"),
// 	}
// 	_ = GenerateOrderByExpr(`uid, age desc`, dict) // => uid ASC, birthday ASC
// 	_ = GenerateOrderByExpr(`age, username desc`, dict) // => birthday DESC, firstname DESC, lastname DESC
func GenerateOrderByExpr(source string, dict PropertyDict) string {
	return xdbutils.GenerateOrderByExpr(source, dict)
}

package xgorm

import (
	"fmt"
	"github.com/Aoi-hosizora/ahlib-db/internal"
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// Gorm db dialect.
const (
	MySQL    = "mysql"
	SQLite   = "sqlite3"
	Postgres = "postgres"
)

// MySQLConfig is an alias type of mysql.Config.
type MySQLConfig = mysql.Config

// MySQLDefaultDsn returns the MySQL dsn from given parameters with "utf8mb4" charset, parseTime and Local location.
// For more options in dsn, please use mysql.Config or xgorm.MySQLConfig.
func MySQLDefaultDsn(username, password, address, database string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, address, database,
	)
}

// PostgresDefaultDsn returns the Postgres dsn from given parameters.
func PostgresDefaultDsn(username, password, host string, port int, database string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		host, port, username, password, database,
	)
}

// IsMySQL checks if the dialect of given gorm.DB is "mysql".
func IsMySQL(db *gorm.DB) bool {
	return db.Dialect().GetName() == MySQL
}

// IsSQLite checks if the dialect of given gorm.DB is "sqlite3".
func IsSQLite(db *gorm.DB) bool {
	return db.Dialect().GetName() == SQLite
}

// IsPostgreSQL checks if the dialect of given gorm.DB is "postgres".
func IsPostgreSQL(db *gorm.DB) bool {
	return db.Dialect().GetName() == Postgres
}

// Reference from http://go-database-sql.org/errors.html.
//
// MySQL: https://github.com/VividCortex/mysqlerr/blob/master/mysqlerr.go and https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.htm,
// SQLite: https://github.com/mattn/go-sqlite3/blob/master/error.go and http://www.sqlite.org/c3ref/c_abort_rollback.html,
// PostgreSQL: https://github.com/lib/pq/blob/master/error.go and https://www.postgresql.org/docs/10/errcodes-appendix.html.
const (
	MySQLDuplicateEntryErrno       = 1062      // MySQLDuplicateEntryErrno is MySQL's ER_DUP_ENTRY errno.
	SQLiteUniqueConstraintErrno    = 19 | 8<<8 // SQLiteUniqueConstraintErrno is SQLite's CONSTRAINT_UNIQUE extended errno.
	PostgreSQLUniqueViolationErrno = "23505"   // PostgreSQLUniqueViolationErrno is PostgreSQL's unique_violation errno.
)

// IsMySQLDuplicateEntryError checks if err is MySQL's ER_DUP_ENTRY error.
func IsMySQLDuplicateEntryError(err error) bool {
	mysqlErr, ok := err.(*mysql.MySQLError)
	return ok && mysqlErr.Number == MySQLDuplicateEntryErrno
}

// IsPostgreSQLUniqueViolationError checks if err is PostgreSQL's unique_violation error.
func IsPostgreSQLUniqueViolationError(err error) bool {
	postgresErr, ok := err.(pq.Error)
	return ok && postgresErr.Code == PostgreSQLUniqueViolationErrno
}

// QueryErr checks gorm.DB query result, will only return xstatus.DbNotFound, xstatus.DbFailed and xstatus.DbSuccess.
func QueryErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case rdb.RecordNotFound():
		return xstatus.DbNotFound, nil // not found
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	}
	return xstatus.DbSuccess, nil
}

// DeleteErr checks gorm.DB delete result, will only return xstatus.DbFailed, xstatus.DbNotFound and xstatus.DbSuccess.
func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	case rdb.RowsAffected == 0:
		return xstatus.DbNotFound, nil // not found
	}
	return xstatus.DbSuccess, nil
}

// PropertyValue represents a PO entity's property mapping rule.
type PropertyValue = internal.PropertyValue

// PropertyDict represents a DTO-PO PropertyValue dictionary, used in GenerateOrderByExp.
type PropertyDict = internal.PropertyDict

// NewPropertyValue creates a PropertyValue by given reverse and destinations.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	return internal.NewPropertyValue(reverse, destinations...)
}

// GenerateOrderByExp returns a generated orderBy expression by given source dto order string (split by ",", such as "name desc, age asc") and PropertyDict.
// The generated expression is in mysql-sql or neo4j-cypher style, that is "xx ASC", "xx.yy DESC".
func GenerateOrderByExp(source string, dict PropertyDict) string {
	return internal.GenerateOrderByExp(source, dict)
}

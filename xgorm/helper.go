package xgorm

import (
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

// IsMySQL checks if the dialect of given gorm.DB is "mysql".
func IsMySQL(db *gorm.DB) bool {
	return db.Dialect().GetName() == "mysql"
}

// IsSQLite checks if the dialect of given gorm.DB is "sqlite".
// func IsSQLite(db *gorm.DB) bool {
// 	return db.Dialect().GetName() == "sqlite"
// }

// IsPostgreSQL checks if the dialect of given gorm.DB is "postgres".
func IsPostgreSQL(db *gorm.DB) bool {
	return db.Dialect().GetName() == "postgres"
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

// IsSQLiteUniqueConstraintError checks if err is SQLite's ErrConstraintUnique.
// func IsSQLiteUniqueConstraintError(err error) bool {
// 	sqliteErr, ok := err.(*sqlite3.Error)
// 	return ok && sqliteErr.ExtendedCode == SQLiteUniqueConstraintErrno
// }

// IsPostgreSQLUniqueViolationError checks if err is PostgreSQL's unique_violation.
func IsPostgreSQLUniqueViolationError(err error) bool {
	postgresErr, ok := err.(*pq.Error)
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

// CreateErr checks gorm.DB create result, will only return xstatus.DbExisted, xstatus.DbFailed and xstatus.DbSuccess.
func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	switch {
	case IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error),
		// IsSQLite(rdb) && IsSQLiteUniqueConstraintError(rdb.Error),
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
		// IsSQLite(rdb) && IsSQLiteUniqueConstraintError(rdb.Error),
		IsPostgreSQL(rdb) && IsPostgreSQLUniqueViolationError(rdb.Error):
		return xstatus.DbExisted, rdb.Error // duplicate
	case rdb.Error != nil:
		return xstatus.DbFailed, rdb.Error // failed
	case rdb.RowsAffected == 0:
		return xstatus.DbNotFound, nil // not found
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
type PropertyValue struct {
	destinations []string // mapping destinations
	reverse      bool     // reverse order mapping
}

// NewPropertyValue creates a PropertyValue by given reverse and destinations.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	finalDestinations := make([]string, 0, len(destinations))
	for _, dest := range destinations {
		dest = strings.TrimSpace(dest)
		if dest != "" {
			finalDestinations = append(finalDestinations, dest) // filter empty destination
		}
	}
	return &PropertyValue{reverse: reverse, destinations: finalDestinations}
}

// Destinations returns the destinations of PropertyValue.
func (p *PropertyValue) Destinations() []string {
	return p.destinations
}

// Reverse returns the reverse of PropertyValue.
func (p *PropertyValue) Reverse() bool {
	return p.reverse
}

// PropertyDict represents a DTO-PO PropertyValue dictionary, used in GenerateOrderByExp.
type PropertyDict map[string]*PropertyValue

// GenerateOrderByExp returns a generated orderBy expresion by given source dto order string (split by ",") and PropertyDict.
func GenerateOrderByExp(source string, dict PropertyDict) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}

	result := make([]string, 0)
	for _, src := range strings.Split(source, ",") {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}
		srcSp := strings.Split(src, " ") // xxx / yyy asc / zzz desc
		if len(srcSp) > 2 {
			continue
		}

		src = srcSp[0]
		desc := len(srcSp) == 2 && strings.ToUpper(srcSp[1]) == "DESC"
		value, ok := dict[src] // property mapping rule
		if !ok || value == nil || len(value.destinations) == 0 {
			continue
		}

		if value.reverse {
			desc = !desc
		}
		for _, prop := range value.destinations {
			direction := " ASC"
			if !desc {
				direction = " DESC"
			}
			result = append(result, prop+direction)
		}
	}

	return strings.Join(result, ", ")
}

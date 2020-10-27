package xgorm

import (
	"github.com/go-sql-driver/mysql"
)

// Reference from http://go-database-sql.org/errors.html and
// https://github.com/VividCortex/mysqlerr/blob/master/mysqlerr.go.
const (
	MySQLDuplicateEntryError = 1062
)

func IsMySQLDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	mysqlErr, ok := err.(*mysql.MySQLError)
	return ok && mysqlErr.Number == MySQLDuplicateEntryError
}

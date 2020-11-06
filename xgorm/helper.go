package xgorm

import (
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xproperty"
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"strings"
)

// Helper is an extension of gorm.DB.
type Helper struct {
	db *gorm.DB
}

// WithDB creates a Helper with gorm.DB.
func WithDB(db *gorm.DB) *Helper {
	return &Helper{db: db}
}

// GetDB gets the original gorm.DB.
func (h *Helper) GetDB() *gorm.DB {
	return h.db
}

// Pagination: db.Limit(limit).Offset((page - 1) * limit)
func (h *Helper) Pagination(limit int32, page int32) *gorm.DB {
	return h.db.Limit(limit).Offset((page - 1) * limit)
}

// Count: db.Model(model).Where(where).Count(&cnt)
func (h *Helper) Count(model interface{}, where interface{}) (int, error) {
	cnt := 0
	rdb := h.db.Model(model).Where(where).Count(&cnt)
	return cnt, rdb.Error
}

// Exist: db.Model(model).Where(where).Count(&cnt)
func (h *Helper) Exist(model interface{}, where interface{}) (bool, error) {
	cnt, err := h.Count(model, where)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// IsMySQL checks the dialect of db is "mysql".
func IsMySQL(db *gorm.DB) bool {
	return db.Dialect().GetName() == "mysql"
}

// Reference from http://go-database-sql.org/errors.html and https://github.com/VividCortex/mysqlerr/blob/master/mysqlerr.go.
const (
	MySQLDuplicateEntryError = 1062 // ER_DUP_ENTRY
)

// IsMySQLDuplicateEntryError checks if err is mysql ER_DUP_ENTRY.
func IsMySQLDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	mysqlErr, ok := err.(*mysql.MySQLError)
	return ok && mysqlErr.Number == MySQLDuplicateEntryError
}

// QueryErr checks gorm.DB's query result.
func QueryErr(rdb *gorm.DB) (bool, error) {
	if rdb.RecordNotFound() {
		return false, nil // not found
	} else if rdb.Error != nil {
		return false, rdb.Error // failed
	}

	return true, nil
}

// CreateErr checks gorm.DB's create result,
// only return xstatus.DbSuccess, xstatus.DbExisted and xstatus.DbFailed.
func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	if IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error) {
		return xstatus.DbExisted, rdb.Error // duplicate
	} else if rdb.Error != nil {
		return xstatus.DbFailed, rdb.Error // failed
	} else if rdb.RowsAffected == 0 {
		return xstatus.DbFailed, fmt.Errorf("unknown error when create") // failed
	}

	return xstatus.DbSuccess, nil
}

// UpdateErr checks gorm.DB's update result,
// only return xstatus.DbSuccess, xstatus.DbExisted, xstatus.DbFailed and xstatus.DbNotFound.
func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	if IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error) {
		return xstatus.DbExisted, rdb.Error // duplicate
	} else if rdb.Error != nil {
		return xstatus.DbFailed, rdb.Error // failed
	} else if rdb.RowsAffected == 0 {
		return xstatus.DbNotFound, nil // not found
	}

	return xstatus.DbSuccess, nil
}

// DeleteErr checks gorm.DB's delete result,
// only return xstatus.DbSuccess, xstatus.DbFailed and xstatus.DbNotFound.
func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	if rdb.Error != nil {
		return xstatus.DbFailed, rdb.Error // failed
	} else if rdb.RowsAffected == 0 {
		return xstatus.DbNotFound, nil // not found
	}

	return xstatus.DbSuccess, nil
}

// OrderByFunc generates a handler function `func(string) string` from xproperty.PropertyDict.
func OrderByFunc(p xproperty.PropertyDict) func(source string) string {
	return func(source string) string {
		if source == "" {
			return ""
		}

		result := make([]string, 0)
		for _, src := range strings.Split(source, ",") {
			src = strings.TrimSpace(src)
			if src == "" {
				continue
			}
			srcSp := strings.Split(src, " ")
			if len(srcSp) > 2 {
				continue
			}

			src = srcSp[0]
			reverse := false
			if len(srcSp) == 2 {
				reverse = strings.ToUpper(srcSp[1]) == "DESC"
			}

			dest, ok := p[src]
			if !ok || dest == nil || len(dest.Destinations) == 0 {
				continue
			}
			if dest.Revert {
				reverse = !reverse
			}
			for _, prop := range dest.Destinations {
				if !reverse {
					prop += " ASC"
				} else {
					prop += " DESC"
				}
				result = append(result, prop)
			}
		}

		return strings.Join(result, ", ")
	}
}

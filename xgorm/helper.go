package xgorm

import (
	"github.com/Aoi-hosizora/ahlib/xproperty"
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/jinzhu/gorm"
	"strings"
)

type Helper struct {
	db *gorm.DB
}

func WithDB(db *gorm.DB) *Helper {
	return &Helper{db: db}
}

func (h *Helper) Pagination(limit int32, page int32) *gorm.DB {
	return h.db.Limit(limit).Offset((page - 1) * limit)
}

func (h *Helper) Count(model interface{}, where interface{}) (int, error) {
	cnt := 0
	rdb := h.db.Model(model).Where(where).Count(&cnt)
	return cnt, rdb.Error
}

func (h *Helper) Exist(model interface{}, where interface{}) (bool, error) {
	cnt, err := h.Count(model, where)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// db.Model(model).Create(object).
func (h *Helper) Create(model interface{}, object interface{}) (xstatus.DbStatus, error) {
	rdb := h.db.Model(model).Create(object)
	return CreateErr(rdb)
}

// db.Model(model).Where(where).Update(object).
func (h *Helper) Update(model interface{}, where interface{}, object interface{}) (xstatus.DbStatus, error) {
	if where == nil {
		where = object
	}
	rdb := h.db.Model(model).Where(where).Update(object)
	return UpdateErr(rdb)
}

// db.Model(model).Where(where).Delete(object).
func (h *Helper) Delete(model interface{}, where interface{}, object interface{}) (xstatus.DbStatus, error) {
	if where == nil {
		where = object
	}
	rdb := h.db.Model(model).Where(where).Delete(object)
	return DeleteErr(rdb)
}

func IsMySQL(db *gorm.DB) bool {
	return db.Dialect().GetName() == "mysql"
}

func QueryErr(rdb *gorm.DB) (bool, error) {
	if rdb.RecordNotFound() {
		return false, nil
	} else if rdb.Error != nil {
		return false, rdb.Error
	}

	return true, nil
}

func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	if IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error) {
		return xstatus.DbExisted, nil
	} else if rdb.Error != nil || rdb.RowsAffected == 0 {
		return xstatus.DbFailed, rdb.Error
	}

	return xstatus.DbSuccess, nil
}

func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	if IsMySQL(rdb) && IsMySQLDuplicateEntryError(rdb.Error) {
		return xstatus.DbExisted, nil
	} else if rdb.Error != nil {
		return xstatus.DbFailed, rdb.Error
	} else if rdb.RowsAffected == 0 {
		return xstatus.DbNotFound, nil
	}

	return xstatus.DbSuccess, nil
}

func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error) {
	if rdb.Error != nil {
		return xstatus.DbFailed, rdb.Error
	} else if rdb.RowsAffected == 0 {
		return xstatus.DbNotFound, nil
	}

	return xstatus.DbSuccess, nil
}

func OrderByFunc(p xproperty.PropertyDict) func(source string) string {
	return func(source string) string {
		result := make([]string, 0)
		if source == "" {
			return ""
		}

		sources := strings.Split(source, ",")
		for _, src := range sources {
			if src == "" {
				continue
			}

			src = strings.TrimSpace(src)
			reverse := strings.HasSuffix(src, " desc") || strings.HasSuffix(src, " DESC")
			src = strings.Split(src, " ")[0]

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

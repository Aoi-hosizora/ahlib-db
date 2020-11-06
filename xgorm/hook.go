package xgorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

const (
	// DefaultDeletedAtTimestamp is the default deletedAt value.
	DefaultDeletedAtTimestamp = "1970-01-01 00:00:01"
)

// GormTime3 is a structure of CreatedAt, UpdatedAt, DeletedAt with "1970-01-01 00:00:01" default, is a replacement of gorm.Model.
type GormTime3 struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index" gorm:"default:'1970-01-01 00:00:01'"`
}

// GormTime2 is a structure of CreatedAt, UpdatedAt, which can customize the DeletedAt field.
type GormTime2 struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HookDeletedAt changes the soft-delete callback (query, row_query, update, delete) using the new deleteAt timestamp.
func HookDeletedAt(db *gorm.DB, deletedAtTimestamp string) {
	// query
	db.Callback().Query().
		Before("gorm:query").
		Register("new_deleted_at_before_query_callback", deletedAtQueryUpdateCallback(deletedAtTimestamp))

	// row query
	db.Callback().RowQuery().
		Before("gorm:row_query").
		Register("new_deleted_at_before_row_query_callback", deletedAtQueryUpdateCallback(deletedAtTimestamp))

	// update
	db.Callback().Update().
		Before("gorm:update").
		Register("new_deleted_at_before_update_callback", deletedAtQueryUpdateCallback(deletedAtTimestamp))

	// delete !!!
	db.Callback().Delete().
		Replace("gorm:delete", deletedAtDeleteCallback(deletedAtTimestamp))
}

// deletedAtQueryUpdateCallback is a callback used in HookDeletedAt, and as a gorm:query, gorm:row_query, gorm:update callback.
func deletedAtQueryUpdateCallback(defaultDeleteAtTimeStamp string) func(scope *gorm.Scope) {
	// https://qiita.com/touyu/items/f1ac43b186cd6b26b8c7
	return func(scope *gorm.Scope) {
		var (
			quotedTableName                   = scope.QuotedTableName()
			deletedAtField, hasDeletedAtField = scope.FieldByName("DeletedAt")
			defaultTimeStamp                  = defaultDeleteAtTimeStamp
		)

		if !scope.HasError() && !scope.Search.Unscoped && hasDeletedAtField {
			scope.Search.Unscoped = true
			sql := fmt.Sprintf("%v.%v = '%v'", quotedTableName, scope.Quote(deletedAtField.DBName), defaultTimeStamp)
			scope.Search.Where(sql)
		}
	}
}

// deletedAtDeleteCallback is a callback used in HookDeletedAt, and as a gorm:delete new callback.
func deletedAtDeleteCallback(deletedAtTimestamp string) func(scope *gorm.Scope) {
	// https://github.com/jinzhu/gorm/blob/master/callback_delete.go
	return func(scope *gorm.Scope) {
		if scope.HasError() {
			return
		}
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}
		var (
			quotedTableName                   = scope.QuotedTableName()
			deletedAtField, hasDeletedAtField = scope.FieldByName("DeletedAt")
			defaultTimeStamp                  = deletedAtTimestamp
		)

		addExtraSpaceIfExist := func(str string) string {
			if str != "" {
				return " " + str
			}
			return ""
		}

		if !scope.Search.Unscoped && hasDeletedAtField {
			var (
				comb = scope.CombinedConditionSql()
				from = fmt.Sprintf("%s IS NULL", scope.Quote(deletedAtField.DBName))
				to   = fmt.Sprintf("%s = '%s'", scope.Quote(deletedAtField.DBName), defaultTimeStamp)
				now  = time.Now().Format("2006-01-02 15:04:05")
			)
			comb = strings.Replace(comb, from, to, 1)

			sql := fmt.Sprintf(
				"UPDATE %v SET %v='%v'%v%v",
				quotedTableName,
				scope.Quote(deletedAtField.DBName), now,
				addExtraSpaceIfExist(comb),
				addExtraSpaceIfExist(extraOption),
			)
			scope.Raw(sql).Exec()
		} else {
			sql := fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)
			scope.Raw(sql).Exec()
		}
	}
}

package xgorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

// DefaultDeletedAtTimestamp represents the default value in `gorm` tag of GormTime.DeletedAt.
const DefaultDeletedAtTimestamp = "1970-01-01 00:00:01"

// GormTime represents a structure of CreatedAt, UpdatedAt, DeletedAt (defaults to DefaultDeletedAtTimestamp), is a replacement of gorm.Model.
type GormTime struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index" gorm:"default:'1970-01-01 00:00:01'"`
}

// GormTime2 represents a structure of CreatedAt, UpdatedAt, which allow you to customize the DeletedAt field, is a replacement of gorm.Model.
type GormTime2 struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HookDeletedAt hooks gorm.DB to replace the soft-delete callback (including query, row_query, update, delete), using the new deletedAt timestamp,
// such as DefaultDeletedAtTimestamp.
func HookDeletedAt(db *gorm.DB, deletedAtTimestamp string) *gorm.DB {
	// query
	db.Callback().Query().
		Before("gorm:query").
		Register("new_deleted_at_before_query_callback", hookedQueryUpdateCallback(deletedAtTimestamp))

	// row query
	db.Callback().RowQuery().
		Before("gorm:row_query").
		Register("new_deleted_at_before_row_query_callback", hookedQueryUpdateCallback(deletedAtTimestamp))

	// update
	db.Callback().Update().
		Before("gorm:update").
		Register("new_deleted_at_before_update_callback", hookedQueryUpdateCallback(deletedAtTimestamp))

	// delete <<<
	db.Callback().Delete().
		Replace("gorm:delete", hookedDeleteCallback(deletedAtTimestamp))

	return db
}

const deletedAtFieldName = "DeletedAt"

// hookedQueryUpdateCallback is a callback for query, row_query and update, used in HookDeletedAt, referred from https://qiita.com/touyu/items/f1ac43b186cd6b26b8c7.
func hookedQueryUpdateCallback(deletedAtTimestamp string) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		var (
			quotedTableName     = scope.QuotedTableName()
			deletedAtField, has = scope.FieldByName(deletedAtFieldName)
		)

		if !scope.HasError() && !scope.Search.Unscoped && has {
			scope.Search.Unscoped = true
			sql := fmt.Sprintf("%s.%s = '%s'", quotedTableName, scope.Quote(deletedAtField.DBName), deletedAtTimestamp)
			scope.Search.Where(sql)
		}
	}
}

// addExtraSpaceIfNotBlank is a string utility function used in hookedDeleteCallback.
func addExtraSpaceIfNotBlank(s string) string {
	if s != "" {
		return " " + s
	}
	return ""
}

// hookedDeleteCallback is a callback for gorm:delete used in HookDeletedAt, referred from https://github.com/jinzhu/gorm/blob/master/callback_delete.go.
func hookedDeleteCallback(deletedAtTimestamp string) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}
		var (
			quotedTableName     = scope.QuotedTableName()
			deletedAtField, has = scope.FieldByName(deletedAtFieldName)
		)

		if !scope.HasError() {
			if !scope.Search.Unscoped && has {
				// replace `deleted_at IS NULL` to `deleted_at = 'xxx'`
				var (
					quotedFieldName = scope.Quote(deletedAtField.DBName)
					isNullCond      = fmt.Sprintf("%s IS NULL", quotedFieldName)
					equalCond       = fmt.Sprintf("%s = '%s'", quotedFieldName, deletedAtTimestamp)
					combCond        = strings.ReplaceAll(scope.CombinedConditionSql(), isNullCond, equalCond)
				)
				sql := fmt.Sprintf(
					"UPDATE %v SET %v='%v'%v%v",
					quotedTableName,
					quotedFieldName,
					time.Now().Format("2006-01-02 15:04:05"),
					addExtraSpaceIfNotBlank(combCond),
					addExtraSpaceIfNotBlank(extraOption),
				)
				scope.Raw(sql).Exec()
			} else {
				sql := fmt.Sprintf(
					"DELETE FROM %v%v%v",
					scope.QuotedTableName(),
					addExtraSpaceIfNotBlank(scope.CombinedConditionSql()),
					addExtraSpaceIfNotBlank(extraOption),
				)
				scope.Raw(sql).Exec()
			}
		}
	}
}

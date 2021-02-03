package xgorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

const (
	// DefaultDeletedAtTimestamp represents the default value of GormTime.DeletedAt.
	DefaultDeletedAtTimestamp = "1970-01-01 00:00:01"
)

// GormTime represents a structure of CreatedAt, UpdatedAt, DeletedAt (defaults to "1970-01-01 00:00:01"), is a replacement of gorm.Model.
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

// HookDeletedAt hooks gorm.DB to replace the soft-delete callback (including query, row_query, update, delete) using the new deletedAt timestamp.
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

// deletedAtQueryUpdateCallback is a callback for gorm:query, gorm:row_query, gorm:update used in HookDeletedAt.
//
// Reference: https://qiita.com/touyu/items/f1ac43b186cd6b26b8c7.
func deletedAtQueryUpdateCallback(deletedAtTimestamp string) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		var (
			quotedTableName                   = scope.QuotedTableName()
			deletedAtField, hasDeletedAtField = scope.FieldByName("DeletedAt")
		)

		if !scope.HasError() && !scope.Search.Unscoped && hasDeletedAtField {
			scope.Search.Unscoped = true
			sql := fmt.Sprintf("%s.%s = '%s'", quotedTableName, scope.Quote(deletedAtField.DBName), deletedAtTimestamp)
			scope.Search.Where(sql)
		}
	}
}

// addExtraSpaceIfNotBlank is a string util function used in deletedAtDeleteCallback.
func addExtraSpaceIfNotBlank(s string) string {
	if s != "" {
		return " " + s
	}
	return ""
}

// deletedAtDeleteCallback is a callback for gorm:delete used in HookDeletedAt.
//
// Reference: https://github.com/jinzhu/gorm/blob/master/callback_delete.go.
func deletedAtDeleteCallback(deletedAtTimestamp string) func(scope *gorm.Scope) {
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
		)

		if !scope.Search.Unscoped && hasDeletedAtField {
			// replace `deleted_at IS NULL` to `deleted_at = 'xxx'`
			var (
				quotedFieldName = scope.Quote(deletedAtField.DBName)
				isNullCond      = fmt.Sprintf("%s IS NULL", quotedFieldName)
				equalCond       = fmt.Sprintf("%s = '%s'", quotedFieldName, deletedAtTimestamp)
				combCond        = strings.Replace(scope.CombinedConditionSql(), isNullCond, equalCond, 1)
			)
			sql := fmt.Sprintf(
				"UPDATE %v SET %v='%v'%v%v",
				quotedTableName,
				quotedFieldName,
				time.Now().Format("2006-01-02 15:04:05"), // scope.db.nowFunc()
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

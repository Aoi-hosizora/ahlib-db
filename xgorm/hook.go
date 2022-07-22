package xgorm

import (
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xstring"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

const (
	// DefaultDeletedAtTimestamp represents the default value in `gorm` tag of GormTime.DeletedAt.
	DefaultDeletedAtTimestamp = "1970-01-01 00:00:01"

	// deletedAtFieldName is the "DeletedAt" field name constant.
	deletedAtFieldName = "DeletedAt"
)

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

// Gorm's callback names. See gorm.Callback or visit https://github.com/jinzhu/gorm/blob/master/callback.go for more details.
const (
	CreateCallbackName   = "gorm:create"
	UpdateCallbackName   = "gorm:update"
	DeleteCallbackName   = "gorm:delete"
	QueryCallbackName    = "gorm:query"
	RowQueryCallbackName = "gorm:row_query"
)

// HookDeletedAt hooks gorm.DB's callbacks to make soft deleting to use the new default deletedAt timestamp, such as DefaultDeletedAtTimestamp.
func HookDeletedAt(db *gorm.DB, defaultTimestamp string) {
	// query
	db.Callback().Query().
		Before(QueryCallbackName).
		Register("new_before_query_callback_for_deleted_at", hookedCallback(defaultTimestamp))

	// row query
	db.Callback().RowQuery().
		Before(RowQueryCallbackName).
		Register("new_before_row_query_callback_for_deleted_at", hookedCallback(defaultTimestamp))

	// update
	db.Callback().Update().
		Before(UpdateCallbackName).
		Register("new_before_update_callback_for_deleted_at", hookedCallback(defaultTimestamp))

	// delete <<<
	db.Callback().Delete().
		Replace(DeleteCallbackName, hookedDeleteCallback(defaultTimestamp))
}

// hookedCallback is a callback for gorm:query, gorm:row_query and gorm:update, used in HookDeletedAt.
func hookedCallback(defaultTimestamp string) func(scope *gorm.Scope) {
	// https://qiita.com/touyu/items/f1ac43b186cd6b26b8c7
	return func(scope *gorm.Scope) {
		if !scope.HasError() && !scope.Search.Unscoped {
			field, has := scope.FieldByName(deletedAtFieldName)
			if has {
				// unscope and use new condition `deleted_at = 'xxx'`
				scope.Search.Unscoped = true
				sql := fmt.Sprintf("%s.%s = '%s'", scope.QuotedTableName(), scope.Quote(field.DBName), defaultTimestamp)
				scope.Search.Where(sql)
			}
		}
	}
}

// hookedDeleteCallback is a callback for gorm:delete used in HookDeletedAt.
func hookedDeleteCallback(defaultTimestamp string) func(scope *gorm.Scope) {
	// https://github.com/jinzhu/gorm/blob/master/callback_delete.go#L29
	// https://github.com/jinzhu/gorm/blob/master/scope.go#L718
	return func(scope *gorm.Scope) {
		if !scope.HasError() {
			var extraOption string
			if i, ok := scope.Get("gorm:delete_option"); ok {
				extraOption = fmt.Sprint(i)
			}
			field, has := scope.FieldByName(deletedAtFieldName)
			if !scope.Search.Unscoped && has {
				// replace condition `deleted_at IS NULL` to `deleted_at = 'xxx'` and use `SET deleted_at = 'yyy'`
				var (
					tblName   = scope.QuotedTableName()
					fieldName = scope.Quote(field.DBName)
					nowVar    = time.Now().Format("2006-01-02 15:04:05")
					condition = strings.ReplaceAll(scope.CombinedConditionSql(),
						fmt.Sprintf("%s.%s IS NULL", tblName, fieldName), fmt.Sprintf("%s.%s = '%s'", tblName, fieldName, defaultTimestamp))
				)
				sql := fmt.Sprintf("UPDATE %v SET %v='%v'%v%v", tblName, fieldName, nowVar,
					xstring.ExtraSpaceOnLeftIfNotBlank(condition), xstring.ExtraSpaceOnLeftIfNotBlank(extraOption))
				scope.Raw(sql).Exec()
			} else {
				sql := fmt.Sprintf("DELETE FROM %v%v%v", scope.QuotedTableName(),
					xstring.ExtraSpaceOnLeftIfNotBlank(scope.CombinedConditionSql()), xstring.ExtraSpaceOnLeftIfNotBlank(extraOption))
				scope.Raw(sql).Exec()
			}
		}
	}
}

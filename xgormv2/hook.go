package xgormv2

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
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
	DeletedAt gorm.DeletedAt `gorm:"index; not null; default:'1970-01-01 00:00:01'"`
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
	RowQueryCallbackName = "gorm:row"
)

// HookDeletedAt hooks gorm.DB's callbacks to make soft deleting to use the new deletedAt timestamp, such as DefaultDeletedAtTimestamp.
func HookDeletedAt(db *gorm.DB, defaultTimestamp string) {
	// query
	_ = db.Callback().Query().
		Before(QueryCallbackName).
		Register("new_before_query_callback_for_deleted_at", hookedCallback(defaultTimestamp, QueryCallbackName))

	// row query
	_ = db.Callback().Row().
		Before(RowQueryCallbackName).
		Register("new_before_row_query_callback_for_deleted_at", hookedCallback(defaultTimestamp, RowQueryCallbackName))

	// update
	_ = db.Callback().Update().
		Before(UpdateCallbackName).
		Register("new_before_update_callback_for_deleted_at", hookedCallback(defaultTimestamp, UpdateCallbackName))

	// delete <<<
	_ = db.Callback().Delete().
		Before(DeleteCallbackName).
		Replace("new_before_delete_callback_for_deleted_at", hookedDeleteCallback(defaultTimestamp))
}

// hookedCallback is a callback for gorm:query, gorm:row and gorm:update, used in HookDeletedAt.
func hookedCallback(defaultTimestamp string, name string) func(db *gorm.DB) {
	// https://github.com/go-gorm/gorm/blob/master/soft_delete.go#L65
	return func(db *gorm.DB) {
		stmt := db.Statement
		if db.Error == nil && stmt.Schema != nil && !stmt.Unscoped {
			field := stmt.Schema.LookUpField(deletedAtFieldName)
			if field != nil {
				// SoftDeleteQueryClause.ModifyStatement / SoftDeleteUpdateClause.ModifyStatement
				if c, ok := stmt.Clauses["WHERE"]; (name != UpdateCallbackName && ok) || (name == UpdateCallbackName && (stmt.DB.AllowGlobalUpdate || ok)) {
					if where, ok := c.Expression.(clause.Where); ok && len(where.Exprs) > 1 {
						for _, expr := range where.Exprs {
							if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
								where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
								c.Expression = where
								stmt.Clauses["WHERE"] = c
								break
							}
						}
					}
				}
				// unscope and use new condition `deleted_at = 'xxx'`
				stmt.Unscoped = true
				eq := clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: field.DBName}, Value: defaultTimestamp}
				stmt.AddClause(clause.Where{Exprs: []clause.Expression{eq}})
			}
		}
	}
}

// hookedDeleteCallback is a callback for gorm:delete used in HookDeletedAt.
func hookedDeleteCallback(defaultTimestamp string) func(db *gorm.DB) {
	// https://github.com/go-gorm/gorm/blob/master/callbacks/delete.go#L113
	// https://github.com/go-gorm/gorm/blob/master/soft_delete.go#L131
	return func(db *gorm.DB) {
		stmt := db.Statement
		if db.Error == nil && stmt.Schema != nil && !stmt.Unscoped {
			field := db.Statement.Schema.LookUpField(deletedAtFieldName)
			if field != nil {
				// replace condition `deleted_at IS NULL` to `deleted_at = 'xxx'`
				stmt.Unscoped = true
				nowTime := time.Now()
				as := clause.Assignment{Column: clause.Column{Name: field.DBName}, Value: nowTime}
				stmt.AddClause(clause.Set{as})
				stmt.SetColumn(field.DBName, nowTime)

				// SoftDeleteDeleteClause.ModifyStatement
				if stmt.Schema != nil {
					_, queryValues := schema.GetIdentityFieldValuesMap(stmt.ReflectValue, stmt.Schema.PrimaryFields)
					column, values := schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)
					if len(values) > 0 {
						stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
					}
					if stmt.ReflectValue.CanAddr() && stmt.Dest != stmt.Model && stmt.Model != nil {
						_, queryValues = schema.GetIdentityFieldValuesMap(reflect.ValueOf(stmt.Model), stmt.Schema.PrimaryFields)
						column, values = schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)
						if len(values) > 0 {
							stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
						}
					}
				}

				// use `SET deleted_at = 'yyy'`
				if _, ok := stmt.Clauses["WHERE"]; !stmt.DB.AllowGlobalUpdate && !ok {
					_ = stmt.DB.AddError(gorm.ErrMissingWhereClause)
				} else {
					eq := clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: field.DBName}, Value: defaultTimestamp}
					stmt.AddClause(clause.Where{Exprs: []clause.Expression{eq}})
				}
				stmt.AddClauseIfNotExists(clause.Update{})
				stmt.Build(stmt.DB.Callback().Update().Clauses...)
			}
		}
	}
}

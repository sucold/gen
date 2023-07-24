package gen

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen/field"
)

// Dao[T] CRUD methods
type Dao[T any] interface {
	SubQuery[T]
	schema.Tabler
	As(alias string) Dao[T]
	Not(conds ...Condition) Dao[T]
	Or(conds ...Condition) Dao[T]
	Select(columns ...field.Expr) Dao[T]
	Where(conds ...Condition) Dao[T]
	Order(columns ...field.Expr) Dao[T]
	Distinct(columns ...field.Expr) Dao[T]
	Omit(columns ...field.Expr) Dao[T]
	Join(table schema.Tabler, conds ...field.Expr) Dao[T]
	LeftJoin(table schema.Tabler, conds ...field.Expr) Dao[T]
	RightJoin(table schema.Tabler, conds ...field.Expr) Dao[T]
	Group(columns ...field.Expr) Dao[T]
	Having(conds ...Condition) Dao[T]
	Limit(limit int) Dao[T]
	Offset(offset int) Dao[T]
	Scopes(funcs ...func(Dao[T]) Dao[T]) Dao[T]
	Unscoped() Dao[T]
	Attrs(attrs ...field.AssignExpr) Dao[T]
	Assign(attrs ...field.AssignExpr) Dao[T]
	Joins(field field.RelationField) Dao[T]
	Preload(field field.RelationField) Dao[T]
	Clauses(conds ...clause.Expression) Dao[T]
	Create(value ...*T) error
	CreateInBatches(value []*T, batchSize int) error
	Save(value ...*T) error
	First() (result *T, err error)
	Take() (result *T, err error)
	Last() (result *T, err error)
	Find() (results []*T, err error)
	FindInBatches(dest any, batchSize int, fc func(tx Dao[T], batch int) error) error
	FirstOrInit() (result *T, err error)
	FirstOrCreate() (result *T, err error)
	Update(column field.Expr, value any) (info ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info ResultInfo, err error)
	Updates(values any) (info ResultInfo, err error)
	UpdateColumn(column field.Expr, value any) (info ResultInfo, err error)
	UpdateColumns(values any) (info ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info ResultInfo, err error)
	Delete(models ...*T) (info ResultInfo, err error)
	Count() (int64, error)
	Row() *sql.Row
	Rows() (*sql.Rows, error)
	Scan(dest any) error
	Pluck(column field.Expr, dest any) error
	ScanRows(rows *sql.Rows, dest any) error

	Debug() Dao[T]
	WithContext(ctx context.Context) Dao[T]
	WithResult(fc func(tx Dao[T])) ResultInfo
	ReplaceDB(db *gorm.DB)
	Session(config *gorm.Session) Dao[T]
	UpdateFrom(q SubQuery[T]) Dao[T]
	Columns(cols ...field.Expr) Columns
	Returning(value interface{}, columns ...string) Dao[T]
	UnderlyingDB() *gorm.DB
	ReadDB() Dao[T]
	WriteDB() Dao[T]
	WhereStruct(get field.GetField, data any) Dao[T]
	FindInBatch(batchSize int, fc func(tx Dao[T], batch int) error) (results []*T, err error)

	FindByPage(offset int, limit int) (result []*T, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	ReplaceConnPool(pool gorm.ConnPool)
	UseTable(tableName string)
	UseDB(db *gorm.DB, opts ...DOOption)
	UseModel()
}

type (
	// Condition query condition
	// field.Expr and subquery are expect value
	Condition interface {
		BeCond() any
		CondError() error
	}
)

var (
	_ Condition = (field.Expr)(nil)
	_ Condition = (field.Value)(nil)
	_ Condition = (SubQuery[T])(nil)
	_ Condition = (Dao[T])(nil)
)

// SubQuery sub query interface
type SubQuery[T any] interface {
	underlyingDB() *gorm.DB
	underlyingDO() *DO[T]

	Condition
}

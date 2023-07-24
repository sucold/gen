package gen

import (
	"database/sql"

	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Dao[T] CRUD methods
type BaseDo[T any] interface {
	SubQuery[T]
	schema.Tabler
	As(alias string) Dao[T]
	Not(conds ...Condition) BaseDo[T]
	Or(conds ...Condition) BaseDo[T]
	Select(conds ...field.Expr) BaseDo[T]
	Where(conds ...Condition) BaseDo[T]
	Order(conds ...field.Expr) BaseDo[T]
	Distinct(cols ...field.Expr) BaseDo[T]
	Omit(cols ...field.Expr) BaseDo[T]
	Join(table schema.Tabler, on ...field.Expr) BaseDo[T]
	LeftJoin(table schema.Tabler, on ...field.Expr) BaseDo[T]
	RightJoin(table schema.Tabler, on ...field.Expr) BaseDo[T]
	Group(cols ...field.Expr) BaseDo[T]
	Having(conds ...Condition) BaseDo[T]
	Limit(limit int) BaseDo[T]
	Offset(offset int) BaseDo[T]
	Scopes(funcs ...func(Dao[T]) Dao[T]) BaseDo[T]
	Unscoped() BaseDo[T]
	Attrs(attrs ...field.AssignExpr) BaseDo[T]
	Assign(attrs ...field.AssignExpr) BaseDo[T]
	Joins(fields ...field.RelationField) BaseDo[T]
	Preload(fields ...field.RelationField) BaseDo[T]
	Clauses(conds ...clause.Expression) BaseDo[T]
	Create(value ...*T) error
	CreateInBatches(value []*T, batchSize int) error
	Save(value ...*T) error
	First() (result *T, err error)
	Take() (result *T, err error)
	Last() (result *T, err error)
	Find() (results []*T, err error)
	FindInBatches(dest *[]*T, batchSize int, fc func(tx Dao[T], batch int) error) error
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

	// Debug() BaseDo[T]
	// WithContext(ctx context.Context) BaseDo[T]
	// WithResult(fc func(tx Dao[T])) ResultInfo
	// ReplaceDB(db *gorm.DB)
	// ReadDB() BaseDo[T]
	// WriteDB() BaseDo[T]
	// Session(config *gorm.Session) BaseDo[T]
	// Columns(cols ...field.Expr) Columns
	// WhereStruct(get field.GetField, data any) BaseDo[T]
	// FindInBatch(batchSize int, fc func(tx Dao[T], batch int) error) (results []*T, err error)
	// UpdateFrom(q SubQuery[T]) Dao[T]
	// FindByPage(offset int, limit int) (result []*T, count int64, err error)
	// ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	// Returning(value interface{}, columns ...string) BaseDo[T]
	// UnderlyingDB() *gorm.DB
}

func id[T any]() BaseDo[T] {
	data := DO[T]{}
	return &data
}
